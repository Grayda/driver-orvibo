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

// OrviboSocket holds info about our socket.
type OrviboSocket struct {
	driver       ninja.Driver
	info         *model.Device
	sendEvent    func(event string, payload interface{}) error
	onOffChannel *channels.OnOffChannel
	Device       orvibo.Device
}

func NewOrviboDevice(driver ninja.Driver, id orvibo.Device) *OrviboSocket {
	name := id.Name

	device := &OrviboSocket{
		driver: driver,
		Device: id,
		info: &model.Device{
			NaturalID:     fmt.Sprintf("socket%s", id.MACAddress),
			NaturalIDType: "socket",
			Name:          &name,
			Signatures: &map[string]string{
				"ninja:manufacturer": "Orvibo",
				"ninja:productName":  "OrviboSocket",
				"ninja:productType":  "Socket",
				"ninja:thingType":    "socket",
			},
		},
	}

	device.onOffChannel = channels.NewOnOffChannel(device)
	return device
}

func (l *OrviboSocket) GetDeviceInfo() *model.Device {
	return l.info
}

func (l *OrviboSocket) GetDriver() ninja.Driver {
	return l.driver
}

func (l *OrviboSocket) SetOnOff(state bool) error {
	fmt.Println("Setting state to", state)
	orvibo.SetState(l.Device.MACAddress, state)
	l.onOffChannel.SendState(state)
	return nil
}

func (l *OrviboSocket) ToggleOnOff() error {
	fmt.Println("Toggling state")
	orvibo.ToggleState(l.Device.MACAddress)
	l.onOffChannel.SendState(l.Device.State)
	return nil
}

func (l *OrviboSocket) SetEventHandler(sendEvent func(event string, payload interface{}) error) {
	l.sendEvent = sendEvent
}

var reg, _ = regexp.Compile("[^a-z0-9]")

// Exported by service/device schema
func (l *OrviboSocket) SetName(name *string) (*string, error) {

	log.Printf("Setting device name to %s", *name)

	safe := reg.ReplaceAllString(strings.ToLower(*name), "")
	if len(safe) > 16 {
		safe = safe[0:16]
	}

	log.Printf("We can only set 5 lowercase alphanum. Name now: %s", safe)
	l.Device.Name = safe
	l.sendEvent("renamed", safe)

	return &safe, nil
}
