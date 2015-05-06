package main

import (
	"fmt"                         // For outputting stuff to the screen
	"github.com/Grayda/go-orvibo" // The magic part that lets us control sockets

	//	"github.com/davecgh/go-spew/spew"     // For neatly outputting stuff
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
	Name        string // A short name for the IR code
	Description string
	Code        string // The IR code itself
}

// OrviboDriverConfig holds config info. I don't think it's extensively used in this driver?
type OrviboDriverConfig struct {
	Initialised    bool           // Has our driver run once before?
	Codes          []OrviboIRCode // Saved IR codes
	learningIR     bool
	learningIRName string
}

// No config provided? Set up some defaults
func defaultConfig() *OrviboDriverConfig {

	c := []OrviboIRCode{} // Blank IR code

	return &OrviboDriverConfig{
		Initialised: false,
		Codes:       c,
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
							saveIR
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

func (d *OrviboDriver) saveIR(config OrviboDriverConfig) error {

	existing := d.config.get(config.ID)

	if existing != nil {
		existing.Pin = tvcfg.Pin
		existing.Name = tvcfg.Name
		existing.IP = tvcfg.IP
		existing.ID = tvcfg.Name + tvcfg.Pin
	} else {
		tvcfg.ID = tvcfg.Name + tvcfg.Pin
		d.config.TVs[tvcfg.ID] = &tvcfg

		go d.createTVDevice(&tvcfg)
	}

	tv := lgtv.TV{}
	tv.Id = tvcfg.ID
	tv.Ip = tvcfg.IP
	tv.Name = tvcfg.Name
	tv.Pin = tvcfg.Pin
	fmt.Print("Save Config - ID:%s IP:%s\n", tv.Id, tvcfg.IP.String())
	tv.PairWithPin()

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
