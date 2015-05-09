package main

import (
	"fmt"                                 // For outputting stuff to the screen
	"github.com/Grayda/go-orvibo"         // The magic part that lets us control sockets
	"github.com/ninjasphere/go-ninja/api" // Ninja Sphere API
	"github.com/ninjasphere/go-ninja/model"
	"github.com/ninjasphere/go-ninja/support"
	"log"  // Similar thing, I suppose?
	"time" // Used as part of "setInterval" and for pausing code to allow for data to come back
)

// package.json is required, otherwise the app just exits and doesn't show any output
var info = ninja.LoadModuleInfo("./package.json")
var serial string
var driver *OrviboDriver // So we can access this in our configuration.go file

// Are we ready to rock?
var ready = false
var started = false // Stops us from running theloop twice

// OrviboDriver holds info about our driver, including our configuration
type OrviboDriver struct {
	support.DriverSupport
	config *OrviboDriverConfig
	conn   *ninja.Connection
	device map[int]*OrviboDevice
}

// OrviboIRCode is a struct that holds info about saved IR codes. Used with config
type OrviboIRCode struct {
	ID          int    // The index of our code
	Name        string // A short name for the IR code
	Description string
	Code        string // The IR code itself
	AllOne      string // Which AllOne to blast through (MACAddress)
	Group       string
}

type OrviboIRCodeGroup struct {
	ID          int    // The index of our code
	Name        string // A short name for the IR code
	Description string
}

// OrviboDriverConfig holds config info. I don't think it's extensively used in this driver?
type OrviboDriverConfig struct {
	Initialised           bool                // Has our driver run once before?
	Codes                 []OrviboIRCode      // Saved IR codes
	CodeGroups            []OrviboIRCodeGroup // Logical groupings of IR codes
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

// NewDriver does what it says on the tin: makes a new driver for us to run.
func NewDriver() (*OrviboDriver, error) {

	// Make a new OrviboDriver. Ampersand means to make a new copy, not reference the parent one (so A = new B instead of A = new B, C = A)
	driver = &OrviboDriver{}
	driver.device = make(map[int]*OrviboDevice)
	// Initialize our driver. Throw back an error if necessary. Remember, := is basically a short way of saying "var blah string = 'abcd'"
	err := driver.Init(info)

	if err != nil {
		log.Fatalf("Failed to initialize Orvibo driver: %s", err)
	}

	// Now we export the driver so the Sphere can find it (?)
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

	d.config = config

	if !d.config.Initialised {
		d.config = defaultConfig()
	}
	d.config = config

	d.Conn.MustExportService(&configService{d}, "$driver/"+info.ID+"/configure", &model.ServiceAnnouncement{
		Schema: "/protocol/configuration",
	})

	if started == false {
		theloop(d, config)
	}

	return d.SendEvent("config", config)
}

func theloop(d *OrviboDriver, config *OrviboDriverConfig) error {
	go func() {
		started = true
		fmt.Println("Calling theloop")
		// These are our SetIntervals that run. To cancel one, simply send "<- true" to it (e.g. autoDiscover <- true)
		var autoDiscover, resubscribe chan bool

		ready, err := orvibo.Prepare() // You ready?
		if ready == true {             // Yep! Let's do this!
			// Because we'll never reach the end of the for loop (in theory),
			// we run SendEvent here.

			autoDiscover = setInterval(orvibo.Discover, time.Minute)
			resubscribe = setInterval(orvibo.Subscribe, time.Minute*3)
			orvibo.Discover() // Discover all sockets

			for { // Loop forever
				select { // This lets us do non-blocking channel reads. If we have a message, process it. If not, check for UDP data and loop
				case msg := <-orvibo.Events:
					switch msg.Name {
					case "existingsocketfound":
						fallthrough
					case "socketfound":
						fmt.Println("Socket found! MAC address is", msg.DeviceInfo.MACAddress)
						orvibo.Subscribe() // Subscribe to any unsubscribed sockets
						orvibo.Query()     // And query any unqueried sockets
					case "existingallonefound":
						fallthrough
					case "allonefound":
						orvibo.Subscribe()
						orvibo.Query()
					case "subscribed":
						if msg.DeviceInfo.Subscribed == false {

							fmt.Println("Subscription successful!")
							orvibo.Devices[msg.DeviceInfo.MACAddress].Subscribed = true
							orvibo.Query()
							fmt.Println("Query called")

						}
						orvibo.Query()
					case "queried":
						fmt.Println("Query event called")
						if msg.DeviceInfo.Queried == false {
							fmt.Println("Not queried. Name is", msg.DeviceInfo.Name)
							d.device[msg.DeviceInfo.ID] = NewOrviboDevice(d, msg.DeviceInfo)
							d.device[msg.DeviceInfo.ID].Device.Name = msg.DeviceInfo.Name

							if msg.DeviceInfo.DeviceType == orvibo.SOCKET {
								_ = d.Conn.ExportDevice(driver.device[msg.DeviceInfo.ID])
								_ = d.Conn.ExportChannel(driver.device[msg.DeviceInfo.ID], driver.device[msg.DeviceInfo.ID].onOffChannel, "on-off")
								d.device[msg.DeviceInfo.ID].Device.State = msg.DeviceInfo.State
								d.device[msg.DeviceInfo.ID].onOffChannel.SendState(msg.DeviceInfo.State)
							}
							orvibo.Devices[msg.DeviceInfo.MACAddress].Queried = true

						} else {
							fmt.Println("Already queried")
						}

					case "ircode":
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
					case "statechanged":
						fmt.Println("State changed to:", msg.DeviceInfo.State)
						if msg.DeviceInfo.Queried == true {
							d.device[msg.DeviceInfo.ID].Device.State = msg.DeviceInfo.State
							d.device[msg.DeviceInfo.ID].onOffChannel.SendState(msg.DeviceInfo.State)
						}
					case "quit":
						autoDiscover <- true
						resubscribe <- true
					}
				default:
					orvibo.CheckForMessages()
				}

			}
		} else {
			fmt.Println("Error:", err)

		}

	}()
	return nil
}

func (d *OrviboDriver) saveIR(config *OrviboDriverConfig, ir OrviboIRCode) error {

	d.config.learningIR = false
	d.config.learningIRName = ""
	d.config.learningIRDevice = ""
	d.config.learningIRDescription = ""

	d.config.Codes = append(d.config.Codes, ir)

	return d.SendEvent("config", d.config)

}

func (d *OrviboDriver) saveGroups(config *OrviboDriverConfig) error {
	return d.SendEvent("config", d.config)
}

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

func (d *OrviboDriver) Stop() error {
	return fmt.Errorf("This driver does not support being stopped. YOU HAVE NO POWER HERE.")

}

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
