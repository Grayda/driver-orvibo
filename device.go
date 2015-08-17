package main

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/Grayda/go-orvibo"
	"github.com/ninjasphere/go-ninja/api"
	"github.com/ninjasphere/go-ninja/channels"
	"github.com/ninjasphere/go-ninja/model"
)

// OrviboDevice holds info about our socket.
type OrviboDevice struct {
	driver       ninja.Driver
	info         *model.Device
	sendEvent    func(event string, payload interface{}) error // For pasing info back to the API. Use this to send configs and such
	onOffChannel *channels.OnOffChannel                        // There are other channels, but
	Device       *orvibo.Device
}

// NewOrviboDevice is called when When go-orvibo finds a new Orvibo device. The results are then appended to an array
// Please read over go-orvibo to learn more about orvibo.Device (which is go-orvibo's internal list of devices)
func NewOrviboDevice(driver ninja.Driver, id *orvibo.Device) *OrviboDevice {
	// I know what you're thinking, coz I'm thinking too. Not every OrviboDevice is a socket. go-orvibo takes care of this check for us
	name := id.Name

	device := &OrviboDevice{
		driver: driver,
		Device: id,
		info: &model.Device{
			NaturalID:     fmt.Sprintf("socket%s", id.MACAddress),
			NaturalIDType: "socket",
			Name:          &name,
			Signatures: &map[string]string{ // This stuff appears in our JSON when browsing the "things" part of REST, plus also in MQTT
				"ninja:manufacturer": "Orvibo",
				"ninja:productName":  "OrviboDevice",
				"ninja:productType":  "Socket", // I think thingType (and maybe productType) is stored in a redis database
				"ninja:thingType":    "socket",
			},
		},
	}

	// Make a new on / off channel. Channels are what info can be passed through. For example, onoff is for things like
	// lightswitches (and appear in the Sphere app as a power button with power usage graph), while BrightnessChannel is for things like lights or TV brightness etc.
	device.onOffChannel = channels.NewOnOffChannel(device)
	return device
}

// GetDeviceInfo just returns some info back to the API (I think?) about what type of thing it is etc., to show the proper icon in the Sphere app
func (d *OrviboDevice) GetDeviceInfo() *model.Device {
	return d.info
}

// GetDriver does a similar thing as GetDeviceInfo. Because it's got a capital letter at the start, it's exported by Go, so other packages can access it.
func (d *OrviboDevice) GetDriver() ninja.Driver {
	return d.driver
}

// SetOnOff does what it says on the tin: Turns our socket on or off. This function is called when you tap an icon on the LED matrix or turn it on / off in the Sphere app
func (d *OrviboDevice) SetOnOff(state bool) error {
	fmt.Println("Setting state to", state)
	orvibo.SetState(d.Device.MACAddress, state)
	d.onOffChannel.SendState(state)
	return nil
}

// ToggleOnOff does a imilar thing to SetOnOff, but is state independent. If it's on, turn it off, and the other way around
func (d *OrviboDevice) ToggleOnOff() error {
	fmt.Println("Toggling state")
	// Asks go-orvibo to do this for us
	orvibo.ToggleState(d.Device.MACAddress)
	// Tells the Ninja Sphere what our current state is
	d.onOffChannel.SendState(d.Device.State)
	return nil
}

// SetEventHandler is something I have no idea of. Looks like it's just handing a sendEvent off to the OrviboDevice. Necessary?
func (d *OrviboDevice) SetEventHandler(sendEvent func(event string, payload interface{}) error) {
	d.sendEvent = sendEvent
}

// Regex that finds a-z (lowercase) and 0-9. Used when creating a safe name for our socket
var reg, _ = regexp.Compile("[^a-z0-9]")

// SetName might not be used by us, but the Sphere might do this for us (though I doubt it?). If we rename a device in the Sphere app, make it into a name that won't cause crashes
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
