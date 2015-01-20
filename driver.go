package main

import (
	"fmt"  // For outputting stuff to the screen
	"log"  // Similar thing, I suppose?
	"time" // Used as part of "setInterval" and for pausing code to allow for data to come back

	"github.com/Grayda/sphere-orvibo/allone" // The magic part that lets us control sockets
	"github.com/ninjasphere/go-ninja/api"    // Ninja Sphere API
	"github.com/ninjasphere/go-ninja/support"
)

// package.json is required, otherwise the app just exits and doesn't show any output
var info = ninja.LoadModuleInfo("./package.json")

// Are we ready to rock?
var ready = false

// OrviboDriver holds info about our driver, including our configuration
type OrviboDriver struct {
	support.DriverSupport
	config *OrviboDriverConfig
}

// OrviboDriverConfig holds config info. I don't think it's extensively used in this driver.
type OrviboDriverConfig struct {
	Initialised    bool
	NumberOfLights int
}

// No config provided? Set up some defaults
func defaultConfig() *OrviboDriverConfig {
	return &OrviboDriverConfig{
		Initialised:    false,
		NumberOfLights: 0,
	}
}

// NewDriver does what it says on the tin: makes a new driver for us to run.
func NewDriver() (*OrviboDriver, error) {

	// Copy (?) our OrviboDriver into this variable instead of making a new copy
	driver := &OrviboDriver{}

	// Initialize our driver. Throw back an error if necessary. Remember, := is basically a short way of saying "var blah string = 'abcd'"
	err := driver.Init(info)
	if err != nil {
		log.Fatalf("Failed to initialize Orvibo driver: %s", err)
	}

	// Now we export the driver so the Sphere can find it (?)
	err = driver.Export(driver)
	if err != nil {
		log.Fatalf("Failed to export Orvibo driver: %s", err)
		allone.Close()
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

	// These are our SetIntervals that run. To cancel one, simply send "<- true" to it (e.g. autoDiscover <- true)
	var autoDiscover, resubscribe chan bool

	var device *OrviboSocket

	d.SendEvent("config", config)

	for {

		allone.CheckForMessages()

		select {
		case msg := <-allone.Events:
			fmt.Println("Event:", msg.Name, "For socket:", msg.SocketInfo.MACAddress)
			switch msg.Name {
			case "ready":

				log.Printf("Ready to go!")

				autoDiscover = setInterval(allone.Discover, time.Minute)
				resubscribe = setInterval(allone.Subscribe, time.Minute*3)

			case "socketfound":

				fmt.Println("We have a socket!")
				// firstSubscribe = setInterval(allone.Subscribe, time.Second)
				allone.Subscribe()

			case "subscribed":
				fmt.Println("We're subscribed!")

				allone.Query()
				continue
			case "queried":

				fmt.Println("We've queried. Name is:", msg.SocketInfo.Name)
				device = NewOrviboSocket(d, msg.SocketInfo)
				fmt.Println("Here")
				device.Socket.Name = msg.Name
				device.Socket.State = msg.SocketInfo.State
				err := d.Conn.ExportDevice(device)
				err = d.Conn.ExportChannel(device, device.onOffChannel, "on-off")

				if err != nil {
					log.Fatalf("Failed to export Orvibo socket on off channel %s: %s", msg.SocketInfo.MACAddress, err)
					allone.Close()
				}
				device.onOffChannel.SendState(msg.SocketInfo.State)
				allone.Subscribe()
			case "statechanged":
				fmt.Println("State changed to:", msg.SocketInfo.State)
				device.Socket.State = msg.SocketInfo.State
				device.onOffChannel.SendState(msg.SocketInfo.State)
			case "quit":
				fmt.Println("Quitting")
				autoDiscover <- true
				resubscribe <- true
			}

		}
		allone.CheckForMessages()
		continue
	}

	return d.SendEvent("config", config)
}

func (d *OrviboSocket) Stop() error {
	allone.Close()
	return fmt.Errorf("This driver does not support being stopped. YOU HAVE NO POWER HERE.")

}

type In struct {
	Name string
}

type Out struct {
	Age  int
	Name string
}

func (d *OrviboDriver) Blarg(in *In) (*Out, error) {
	log.Printf("GOT INCOMING! %s", in.Name)
	return &Out{
		Name: in.Name,
		Age:  30,
	}, nil
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
