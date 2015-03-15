package main

import (
	"fmt"
	"github.com/Grayda/go-orvibo"
	"github.com/ninjasphere/go-ninja/api"
	"github.com/ninjasphere/go-ninja/channels"
	"github.com/ninjasphere/go-ninja/model"
	"log"
	"regexp"
	"strings"
)

// OrviboDevice holds info about our socket.
type OrviboDevice struct {
	driver       ninja.Driver
	info         *model.Device
	sendEvent    func(event string, payload interface{}) error
	onOffChannel *channels.OnOffChannel
	Device       orvibo.Device
}

func NewOrviboDevice(driver ninja.Driver, id orvibo.Device) *OrviboDevice {
	name := id.Name

	device := &OrviboDevice{
		driver: driver,
		Device: id,
		info: &model.Device{
			NaturalID:     fmt.Sprintf("socket%s", id.MACAddress),
			NaturalIDType: "socket",
			Name:          &name,
			Signatures: &map[string]string{
				"ninja:manufacturer": "Orvibo",
				"ninja:productName":  "OrviboDevice",
				"ninja:productType":  "Socket",
				"ninja:thingType":    "socket",
			},
		},
	}

	device.onOffChannel = channels.NewOnOffChannel(device)
	return device
}

func (d *OrviboDevice) GetDeviceInfo() *model.Device {
	return d.info
}

func (d *OrviboDevice) GetDriver() ninja.Driver {
	return d.driver
}

func (d *OrviboDevice) SetOnOff(state bool) error {
	fmt.Println("Setting state to", state)
	orvibo.SetState(d.Device.MACAddress, state)
	d.onOffChannel.SendState(state)
	return nil
}

func (d *OrviboDevice) ToggleOnOff() error {
	fmt.Println("Toggling state")
	orvibo.ToggleState(d.Device.MACAddress)
	d.onOffChannel.SendState(d.Device.State)
	return nil
}

func (d *OrviboDevice) SetEventHandler(sendEvent func(event string, payload interface{}) error) {
	d.sendEvent = sendEvent
}

var reg, _ = regexp.Compile("[^a-z0-9]")

// Exported by service/device schema
func (d *OrviboDevice) SetName(name *string) (*string, error) {

	log.Printf("Setting device name to %s", *name)

	safe := reg.ReplaceAllString(strings.ToLower(*name), "")
	if len(safe) > 16 {
		safe = safe[0:16]
	}

	log.Printf("We can only set 5 lowercase alphanum. Name now: %s", safe)
	d.Device.Name = safe
	d.sendEvent("renamed", safe)

	return &safe, nil
}
