package main

import (
	"fmt"
	"log"
	"time"

	"github.com/Grayda/sphere-orvibo/allone"
	"github.com/ninjasphere/go-ninja/api"

	"github.com/ninjasphere/go-ninja/support"
)

var info = ninja.LoadModuleInfo("./package.json")
var ready = false

/*model.Module{
ID:          "com.ninjablocks.OrviboSocket",
Name:        "Fake Driver",
Version:     "1.0.2",
Description: "Just used to test go-ninja",
Author:      "Elliot Shepherd <elliot@ninjablocks.com>",
License:     "MIT",
}*/

type OrviboDriver struct {
	support.DriverSupport
	config *OrviboDriverConfig
}

type OrviboDriverConfig struct {
	Initialised    bool
	NumberOfLights int
}

func defaultConfig() *OrviboDriverConfig {
	return &OrviboDriverConfig{
		Initialised:    false,
		NumberOfLights: 0,
	}
}

func NewDriver() (*OrviboDriver, error) {

	driver := &OrviboDriver{}

	err := driver.Init(info)
	if err != nil {
		log.Fatalf("Failed to initialize Orvibo driver: %s", err)
	}

	err = driver.Export(driver)
	if err != nil {
		log.Fatalf("Failed to export Orvibo driver: %s", err)
		allone.Close()
	}

	return driver, nil
}

func (d *OrviboDriver) Start(config *OrviboDriverConfig) error {
	log.Printf("Driver Starting with config %v", config)

	d.config = config
	if !d.config.Initialised {
		d.config = defaultConfig()
	}
	var firstDiscover, firstSubscribe, firstQuery, autoDiscover, resubscribe chan bool
	var device *OrviboSocket

	d.SendEvent("config", config)
	go func() {
		allone.CheckForMessages()
		for {
			allone.CheckForMessages()
			select {
			case msg := <-allone.Events:
				fmt.Println("!!!T Type:", msg.Name)
				switch msg.Name {
				case "ready":

					log.Printf("Ready to go!")
					firstDiscover = setInterval(allone.Discover, time.Second)
					autoDiscover = setInterval(allone.Discover, time.Minute)
					resubscribe = setInterval(allone.Subscribe, time.Minute*3)

				case "socketfound":
					firstDiscover <- true // Stop our setInterval
					fmt.Println("We have a socket!")

					if msg.SocketInfo.Subscribed == false {
						firstSubscribe = setInterval(allone.Subscribe, time.Second)
					}
					allone.CheckForMessages()
					time.Sleep(time.Second)

				case "subscribed":
					go func() { firstSubscribe <- true }()
					fmt.Println("We're subscribed!")
					time.Sleep(time.Millisecond * 100)
					firstQuery = setInterval(allone.Query, time.Second)
					allone.CheckForMessages()
				case "queried":
					firstQuery <- true
					fmt.Println("We've queried. Name is:", msg.SocketInfo.Name)
					device = NewOrviboSocket(d, msg.SocketInfo)
					device.Socket.Name = msg.Name
					err := d.Conn.ExportDevice(device)
					err = d.Conn.ExportChannel(device, device.onOffChannel, "on-off")
					if err != nil {
						log.Fatalf("Failed to export Orvibo socket on off channel %s: %s", msg.SocketInfo.MACAddress, err)
						allone.Close()
					}
					allone.Discover()
					allone.Subscribe()
					allone.CheckForMessages()
				case "statechanged":
					fmt.Println("State changed to:", msg.SocketInfo.State)
					allone.CheckForMessages()
				}
			}
			allone.CheckForMessages()

		}
	}()

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
