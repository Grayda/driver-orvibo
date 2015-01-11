package main

import (
	"fmt"
	"github.com/Grayda/sphere-orvibo/allone"
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
	Socket       allone.Socket
}

func NewOrviboSocket(driver ninja.Driver, id allone.Socket) *OrviboSocket {
	name := id.Name

	socket := &OrviboSocket{
		driver: driver,
		Socket: id,
		info: &model.Device{
			NaturalID:     fmt.Sprintf("socket%d", id.MACAddress),
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

	socket.onOffChannel = channels.NewOnOffChannel(socket)
	return socket
}

func (l *OrviboSocket) GetDeviceInfo() *model.Device {
	return l.info
}

func (l *OrviboSocket) GetDriver() ninja.Driver {
	return l.driver
}

func (l *OrviboSocket) SetOnOff(state bool) error {
	allone.SetState(state, l.Socket.MACAddress)
	return nil
}

func (l *OrviboSocket) ToggleOnOff() error {
	allone.ToggleState(l.Socket.MACAddress)
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

	log.Printf("Pretending we can only set 5 lowercase alphanum. Name now: %s", safe)
	l.Socket.Name = safe
	l.sendEvent("renamed", safe)

	return &safe, nil
}
