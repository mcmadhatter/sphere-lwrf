package main

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/ninjasphere/go-ninja/api"
	"github.com/ninjasphere/go-ninja/channels"
	"github.com/ninjasphere/go-ninja/model"
)

// LWRFDevice holds info about our socket.
type LWRFDevice struct {
	driver       ninja.Driver
	info         *model.Device
	sendEvent    func(event string, payload interface{}) error
	onOffChannel *channels.OnOffChannel
	Device       DeviceStruct
}

func NewLWRFDevice(driver ninja.Driver, id DeviceStruct) *LWRFDevice {
	name := id.RoomDevice

	device := &LWRFDevice{
		driver: driver,
		Device: id,
		info: &model.Device{
			NaturalID:     fmt.Sprintf("socket%s", name),
			NaturalIDType: "socket",
			Name:          &name,
			Signatures: &map[string]string{
				"ninja:manufacturer": "LWRF",
				"ninja:productName":  "LWRFDevice",
				"ninja:productType":  "Socket",
				"ninja:thingType":    "socket",
			},
		},
	}

	device.onOffChannel = channels.NewOnOffChannel(device)
	return device
}

func (d *LWRFDevice) GetDeviceInfo() *model.Device {
	return d.info
}

func (d *LWRFDevice) GetDriver() ninja.Driver {
	return d.driver
}

func (d *LWRFDevice) SetOnOff(state bool) error {

	SetState(&d.Device, state)
	d.onOffChannel.SendState(state)
	return nil
}

func (d *LWRFDevice) ToggleOnOff() error {

	ToggleState(&d.Device)
	d.onOffChannel.SendState(d.Device.State)
	return nil
}

func (d *LWRFDevice) SetEventHandler(sendEvent func(event string, payload interface{}) error) {
	d.sendEvent = sendEvent
}

var reg, _ = regexp.Compile("[^a-z0-9]")

// Exported by service/device schema
func (d *LWRFDevice) SetName(name *string) (*string, error) {

	fmt.Println("Setting device name to %s", *name)

	safe := reg.ReplaceAllString(strings.ToLower(*name), "")
	if len(safe) > 16 {
		safe = safe[0:16]
	}

	fmt.Println("We can only set 5 lowercase alphanum. Name now: %s", safe)
	d.Device.PoliteName = safe
	d.sendEvent("renamed", safe)

	return &safe, nil
}
