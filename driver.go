package main

import (
	"fmt"  // For outputting stuff to the screen
	"log"  // Similar thing, I suppose?
	"time" // Used as part of "setInterval" and for pausing code to allow for data to come back

	"github.com/Grayda/go-orvibo"         // The magic part that lets us control sockets
	"github.com/ninjasphere/go-ninja/api" // Ninja Sphere API
	"github.com/ninjasphere/go-ninja/model"
	"github.com/ninjasphere/go-ninja/support"
)

// package.json is required, otherwise the app just exits and doesn't show any output
var info = ninja.LoadModuleInfo("./package.json")
var serial string        // Declared but not used?
var driver *OrviboDriver // So we can access this in our configuration.go file

// Are we ready to rock? This is sphere-orvibo only code by the way. You don't need to do this in your own driver?
var ready = false
var started = false // Stops us from running theloop twice

// OrviboDriver holds info about our driver, including our configuration
type OrviboDriver struct {
	support.DriverSupport
	config *OrviboDriverConfig // This is how we save and load IR codes and such. Call this by using driver.config
	conn   *ninja.Connection
	device map[int]*OrviboDevice // A list of devices we've found. This is in addition to the list go-orvibo maintains
}

// OrviboIRCode is a struct that holds info about saved IR codes. Used with config
type OrviboIRCode struct {
	ID          int    // The index of our code
	Name        string // A short name for the IR code
	Description string
	Code        string // The IR code itself
	AllOne      string // Which AllOne to blast through (MACAddress)
	Group       string // Which group does this code belong to?
}

type OrviboRFCode struct {
	ID          string // The Channel of our code
	Name        string // A short name for the IR code
	Description string
	Code        string // The RF code itself
	AllOne      string // Which AllOne to blast through (MACAddress)
	Group       string // Which group does this code belong to?
}

// OrviboIRCodeGroup is a struct that defines an IR code. This makes it easier to pass to a saveGroup function
// Seriously. Structs are awesome. You should use them all the time.
type OrviboIRCodeGroup struct { // Also applies to RF!
	ID          int    // The index of our code
	Name        string // A short name for the IR code
	Description string
}

// OrviboDriverConfig holds config info. The learningIR* stuff should be in its own struct, but I haven't got that far yet.
type OrviboDriverConfig struct {
	Initialised           bool                // Has our driver run once before?
	Codes                 []OrviboIRCode      // Saved IR codes
	CodeGroups            []OrviboIRCodeGroup // Logical groupings of IR codes
	Switches              map[string]OrviboRFCode
	learningIR            bool
	learningIRName        string
	learningIRDescription string
	learningIRDevice      string
	learningIRGroup       string
}

// No config provided? Set up some defaults
func defaultConfig() *OrviboDriverConfig {
	var cg []OrviboIRCodeGroup
	c := []OrviboIRCode{} // Blank IR code
	cg = append(cg, OrviboIRCodeGroup{
		ID:          0,
		Name:        "Main",
		Description: "",
	},
	)

	return &OrviboDriverConfig{
		Initialised: false,
		Codes:       c,
		CodeGroups:  cg,
	}
}

// NewDriver does what it says on the tin: makes a new driver for us to run. This is called through main.go
func NewDriver() (*OrviboDriver, error) {

	// Make a new OrviboDriver. Ampersand means to make a new copy, not reference the parent one (so A = new B instead of A = new B, C = A)
	driver = &OrviboDriver{}
	// Empty map of OrviboDevices
	driver.device = make(map[int]*OrviboDevice)
	// Initialize our driver. Throw back an error if necessary. Remember, := is basically a short way of saying "var blah string = 'abcd'"
	err := driver.Init(info)

	if err != nil {
		log.Fatalf("Failed to initialize Orvibo driver: %s", err)
	}

	// Now we export the driver so the Sphere can find it
	err = driver.Export(driver)

	if err != nil {
		log.Fatalf("Failed to export Orvibo driver: %s", err)
	}

	// NewDriver returns two things, OrviboDriver, and an error if present
	return driver, nil
}

// Start is where the fun and magic happens! The driver is fired up and starts finding sockets
func (d *OrviboDriver) Start(config *OrviboDriverConfig) error {
	log.Printf("Driver Starting with config %v", config)

	d.config = config // Load our config

	if !d.config.Initialised { // No config loaded? Make one
		d.config = defaultConfig()
	}

	d.config.Switches = make(map[string]OrviboRFCode)

	// This tells the API that we're going to expose a UI, and to run GetActions() in configuration.go
	d.Conn.MustExportService(&configService{d}, "$driver/"+info.ID+"/configure", &model.ServiceAnnouncement{
		Schema: "/protocol/configuration",
	})

	// If we've not started the driver
	if started == false {
		// Start a loop that handles everything this driver does (finding sockets, blasting IR etc.)
		// We put it in its own loop to keep the code neat
		theloop(d, config)
	}

	return d.SendEvent("config", config)
}

func theloop(d *OrviboDriver, config *OrviboDriverConfig) error {
	// Run this concurrently to ensure the rest of the driver isn't held up on an infinite loop
	go func() {
		// If started = true, theloop isn't called twice
		started = true
		fmt.Println("Calling theloop")

		// These are our SetIntervals that run. To cancel one, simply send "<- true" to it (e.g. autoDiscover <- true)
		var autoDiscover, resubscribe chan bool

		ready, err := orvibo.Prepare() // You ready? Ask orvibo to start listening on sockets and such.
		if ready == true {             // Yep! Let's do this!
			autoDiscover = setInterval(orvibo.Discover, time.Minute)   // Every minute, try and find new sockets
			resubscribe = setInterval(orvibo.Subscribe, time.Minute*3) // Every 3 minutes, resubscribe.
			orvibo.Discover()                                          // Discover all sockets

			for { // Loop forever
				select { // This lets us do non-blocking channel reads. If we have a message, process it. If not, check for UDP data and loop
				case msg := <-orvibo.Events: // If there is an event waiting
					switch msg.Name {
					case "existingsocketfound": // Found an existing socket. Don't do anything, so just keep going
						fallthrough
					case "socketfound": // Socket has been found go-orvibo has taken care of storing the details in DeviceInfo, so do that.
						fmt.Println("Socket found! MAC address is", msg.DeviceInfo.MACAddress)
						orvibo.Subscribe() // Subscribe to any unsubscribed sockets
						orvibo.Query()     // And query any unqueried sockets
					case "existingallonefound":
						fallthrough
					case "allonefound":
						orvibo.Subscribe()
						orvibo.Query()
					case "subscribed": // We've asked to subscribe to a device and we've had confirmation
						if msg.DeviceInfo.Subscribed == false { // If we've not subscribed before

							fmt.Println("Subscription successful!")
							orvibo.Devices[msg.DeviceInfo.MACAddress].Subscribed = true
							orvibo.Query() // Ask the device for its name
							fmt.Println("Query called")

						}
						orvibo.Query()
					case "queried": // We've asked for a name and we've got the info back
						fmt.Println("Query event called")
						if msg.DeviceInfo.Queried == false {

							d.device[msg.DeviceInfo.ID] = NewOrviboDevice(d, msg.DeviceInfo) // Now we add this to d.device[].Device because we can now control it
							d.device[msg.DeviceInfo.ID].Device.Name = msg.DeviceInfo.Name

							if msg.DeviceInfo.DeviceType == orvibo.SOCKET { // If it's a socket,
								_ = d.Conn.ExportDevice(driver.device[msg.DeviceInfo.ID])                                                           // Let the Sphere know about it
								_ = d.Conn.ExportChannel(driver.device[msg.DeviceInfo.ID], driver.device[msg.DeviceInfo.ID].onOffChannel, "on-off") // Let the Sphere know we've got an on-off channel ready
								d.device[msg.DeviceInfo.ID].Device.State = msg.DeviceInfo.State                                                     // Set the state for internal reference
								d.device[msg.DeviceInfo.ID].onOffChannel.SendState(msg.DeviceInfo.State)                                            // And tell the Sphere what the initial state is. Easy!
								// Now when you go into the Sphere app, there will be a thing ready to add ("Promoted" is true, I think, which makes it show up in the Add Things menu)
							}
							orvibo.Devices[msg.DeviceInfo.MACAddress].Queried = true // We've queried it before

						} else {
							fmt.Println("Already queried")
						}

					case "ircode": // We're in learning mode and an IR code has come back
						if driver.config.learningIR == true {
							ir := OrviboIRCode{
								Name:        driver.config.learningIRName,
								Code:        msg.DeviceInfo.LastIRMessage,
								Description: driver.config.learningIRDescription,
								AllOne:      driver.config.learningIRDevice,
								Group:       driver.config.learningIRGroup,
							}
							driver.saveIR(driver.config, ir)
						}
					case "statechanged": // Something has changed our status (e.g. we've pressed the button on a socket)
						fmt.Println("State changed to:", msg.DeviceInfo.State)
						if msg.DeviceInfo.Queried == true { // If we've queried
							d.device[msg.DeviceInfo.ID].Device.State = msg.DeviceInfo.State          // Save the state
							d.device[msg.DeviceInfo.ID].onOffChannel.SendState(msg.DeviceInfo.State) // And let the Sphere know about it
						}
					case "quit": // We're done. Of course the driver has no quit function (yet?)
						autoDiscover <- true
						resubscribe <- true
					}
				default: // If there are no messages to parse, check for new UDP messages
					orvibo.CheckForMessages()
				}

			}
		} else {
			fmt.Println("Error:", err)

		}

	}()
	return nil
}

// saveIR does what it says on the tin. Takes a hex IR code and stores it in our config
func (d *OrviboDriver) saveIR(config *OrviboDriverConfig, ir OrviboIRCode) error {

	d.config.learningIR = false
	d.config.learningIRName = ""
	d.config.learningIRDevice = ""
	d.config.learningIRDescription = ""

	d.config.Codes = append(d.config.Codes, ir)

	return d.SendEvent("config", d.config)

}

func (d *OrviboDriver) saveRF(config *OrviboDriverConfig, rf OrviboRFCode) error {
	d.config.Switches[rf.ID] = rf
	return d.SendEvent("config", d.config)
}

// Created a new group? Save it. See how stupidly simple saving stuff to the config is? MUCH better than the Ninja Block days!
func (d *OrviboDriver) saveGroups(config *OrviboDriverConfig) error {
	return d.SendEvent("config", d.config)
}

// Again, does what it says on the tin.
func (d *OrviboDriver) deleteIR(config *OrviboDriverConfig, code string) error {
	fmt.Println("========================")
	fmt.Println("Looking for" + code)
	// Go is a stupid language. There is no easy way to delete something from a slice.
	// What I've done here, is loop through all the codes. If the code doesn't equal
	// the code we're looking for, it's saved in the codelist slice. At the end,
	// we replace config.Codes with our new list which doesn't have our code. Easy! ... ish
	var codelist []OrviboIRCode
	for _, ircodes := range d.config.Codes {
		if ircodes.Code != code {
			codelist = append(codelist, ircodes)
		} else {
			fmt.Println("Found", ircodes.Name+".", "Not including in final array..")
		}
	}

	d.config.Codes = codelist
	fmt.Println("Saving options")
	return d.SendEvent("config", d.config)
}

// Stop does nothing. Though if it did, we could pass "quit" to theloop and clean up timers and such. The only way drivers are stopped now, are by force (reboot, Ctrl+C now)
func (d *OrviboDriver) Stop() error {
	return fmt.Errorf("This driver does not support being stopped. YOU HAVE NO POWER HERE.")

}

func stringToBool(i string) bool {
	if i == "true" {
		return true
	}
	return false
}

// Analogous to Javascript's setInterval. Runs a function after a certain duration and keeps running it until "true" is passed to it
func setInterval(what func(), delay time.Duration) chan bool {
	stop := make(chan bool)

	go func() {
		for {
			what()
			select {
			case <-time.After(delay):
			case <-stop:
				return
			}
		}
	}()

	return stop
}
