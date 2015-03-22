package main

/* This LWRF Wifi link driver has been based on the work by Grayda for his sphere-orvibo driver
https://github.com/grayda/sphere-orvibo */

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"regexp"
	"strings" // For outputting stuff to the screen

	// The magic part that lets us control sockets

	//	"github.com/davecgh/go-spew/spew"     // For neatly outputting stuff
	"log" // Similar thing, I suppose?

	// For neatly outputting stuff
	"github.com/franela/goreq"
	"github.com/ninjasphere/go-ninja/api" // Ninja Sphere API
	"github.com/ninjasphere/go-ninja/support"
)

// package.json is required, otherwise the app just exits and doesn't show any output
var info = ninja.LoadModuleInfo("./package.json")
var serial string

// Are we ready to rock?
var ready = false
var started = false // Stops us from running theloop twice
var device = make(map[string]*LWRFDevice)

// LWRFDriver holds info about our driver, including our configuration
type LWRFDriver struct {
	support.DriverSupport
	config *LWRFDriverConfig
	conn   *ninja.Connection
}

// LWRFDriverConfig holds config info. I don't think it's extensively used in this driver?
type LWRFDriverConfig struct {
	Initialised bool
	email       string
	pin         string
	lwrfDevices map[string]*DeviceStruct
}

func defaultConfig() *LWRFDriverConfig {
	return &LWRFDriverConfig{
		Initialised: false,
		email:       "add email here",
		pin:         "add pin here",
	}
}

//DeviceStruct stores alightwave rf device e.g a light or socket
type DeviceStruct struct {
	RoomDevice string
	PoliteName string
	DevName    string
	RoomName   string
	LwrfType   string
	State      bool
	Queried    bool
}

// EventStruct is what we pass back to our calling code via channels
type EventStruct struct {
	Name       string       // The name of the event (e.g. ready, DeviceLWRFfound)
	DeviceInfo DeviceStruct // And our DeviceLWRF struct so we can look at IP address, MAC etc.
}

var lwrfDevices = make(map[string]*DeviceStruct) // All the DeviceLWRFs we've found

var conn *net.UDPConn // Our UDP connection, read and write
var msg []byte

//Devices is a map of the lwrf devices in the system
var Devices = make(map[string]*DeviceStruct)

// NewDriver does what it says on the tin: makes a new driver for us to run.
func NewDriver() (*LWRFDriver, error) {

	// Make a new LWRFDriver. Ampersand means to make a new copy, not reference the parent one (so A = new B instead of A = new B, C = A)
	driver := &LWRFDriver{}

	fmt.Println("Trying driver init")
	// Initialize our driver. Throw back an error if necessary. Remember, := is basically a short way of saying "var blah string = 'abcd'"
	err := driver.Init(info)
	if err != nil {
		log.Fatalf("Failed to initialize LWRF driver: %s", err)
	}

	// Now we export the driver so the Sphere can find it (?)
	err = driver.Export(driver)
	if err != nil {
		log.Fatalf("Failed to export LWRF driver: %s", err)
	}

	// NewDriver returns two things, LWRFDriver, and an error if present
	return driver, nil
}

// Start is where the fun and magic happens! The driver is fired up and starts finding sockets
func (d *LWRFDriver) Start(config *LWRFDriverConfig) error {

	d.config = config
	if !d.config.Initialised {
		d.config = defaultConfig()
	}
	Prepare()
	if started == false {
		broadcastMessage("100,!F*p.")
		LwrfGetDevices(d)
	}

	started = true

	return d.SendEvent("config", config)
}

//Stop - drop the anchors!
func (d *LWRFDriver) Stop() error {
	return fmt.Errorf("This driver does not support being stopped. YOU HAVE NO POWER HERE.")

}

//LwrfGetDevices -  Get all of your lightwaverf devices from the lightwaverfhost.co.uk
func LwrfGetDevices(d *LWRFDriver) {
	itemsPerRoom := 10
	item := url.Values{}
	item.Set("email", d.config.email)
	item.Add("pin", d.config.pin)

	res, err := goreq.Request{
		Uri:         "https://lightwaverfhost.co.uk/manager/",
		QueryString: item,
	}.Do()

	if (err != nil) || (res.StatusCode != 200) {

	} else {

		// Time to parse that stuff
		s, _ := res.Body.ToString()

		r, _ := regexp.Compile("var ([a-zA-Z]*) = (\\[.*?\\]);")

		str := r.FindAllString(s, -1)
		//parse the r findall results	to a JSON file

		r, _ = regexp.Compile("\"([a-zA-Z0-9 \\/><?]*)\"")
		rooms := r.FindAllString(str[2], -1)
		roomStatus := r.FindAllString(str[3], -1)
		lwrfDevices := r.FindAllString(str[0], -1)
		lwrfDeviceStatus := r.FindAllString(str[1], -1)

		for i := 0; i < len(rooms); i++ {
			if roomStatus[i] == "\"A\"" {
				fmt.Printf("Found Room %d: %s", i+1, rooms[i])

				startIndex := itemsPerRoom * i
				for j := startIndex; (j < (startIndex + itemsPerRoom)) && (j < len(lwrfDeviceStatus)); j++ {
					lwrftype := ""
					var FindOK bool

					FindOK = false
					if lwrfDevices[j] == "All Off" {
						/* not currently supported */
						FindOK = false
					} else if strings.ToUpper(lwrfDeviceStatus[j]) == "\"D\"" {
						lwrftype = "Light"

						FindOK = true
					} else if strings.ToUpper(lwrfDeviceStatus[j]) == "\"O\"" {
						lwrftype = "Switch"

						FindOK = true
					} else {
						/* do nothing */
						FindOK = false
					}

					if FindOK == true {
						var Rd string
						var Pl string
						fmt.Printf("Found R%dD%d - it is a %s called %s", i+1, (j%itemsPerRoom)+1, lwrftype, lwrfDevices[j])
						Rd = fmt.Sprintf("R%dD%d", i+1, (j%itemsPerRoom)+1)
						Pl = fmt.Sprintf("LWRF %s %s", rooms[i], lwrfDevices[j])
						Devices[Rd] = &DeviceStruct{Rd, Pl, lwrfDevices[j], rooms[i], lwrftype, false, false}
						DeviceLWRF := DeviceStruct{Rd, Pl, lwrfDevices[j], rooms[i], lwrftype, false, false}

						device[DeviceLWRF.RoomDevice] = NewLWRFDevice(d, DeviceLWRF)

						d.config.Initialised = true
						_ = d.Conn.ExportDevice(device[DeviceLWRF.RoomDevice])
						_ = d.Conn.ExportChannel(device[DeviceLWRF.RoomDevice], device[DeviceLWRF.RoomDevice].onOffChannel, "on-off")
						device[DeviceLWRF.RoomDevice].Device.PoliteName = DeviceLWRF.PoliteName
						device[DeviceLWRF.RoomDevice].Device.State = DeviceLWRF.State
						Devices[DeviceLWRF.RoomDevice].Queried = true
						device[DeviceLWRF.RoomDevice].onOffChannel.SendState(DeviceLWRF.State)

					}
				}
			}
		}
	}
}

func getLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}

	ifaces, _ := net.Interfaces()
	// handle err
	for _, i := range ifaces {
		addrs, _ := i.Addrs()
		// handle err
		for _, addr := range addrs {
			switch v := addr.(type) {
			case *net.IPAddr:
				return v.IP.String(), nil
			}

		}
	}

	return "", errors.New("Unable to find IP address. Ensure you're connected to a network")
}

func broadcastMessage(msg string) {
	fmt.Println("Broadcasting message:", msg, "to", net.IPv4bcast.String()+":9760")
	udpAddr, err := net.ResolveUDPAddr("udp4", net.IPv4bcast.String()+":9760")
	if err != nil {
		fmt.Println("ERROR!:", err)
		os.Exit(1)
	}
	//	buf, _ := hex.DecodeString(msg)
	buf := []byte(msg)
	// If we'e got an error
	if err != nil {
		fmt.Println("ERROR!:", err)
		os.Exit(1)
	}

	_, _ = conn.WriteToUDP(buf, udpAddr)
	return
}

// ToggleState finds out if the DeviceLWRF is on or off, then toggles it
func ToggleState(lwrfid *DeviceStruct) {
	var bcastStr string
	if lwrfid.State {
		bcastStr = fmt.Sprintf("351,!%sF0|Ninja|Off", lwrfid.RoomDevice)
		lwrfid.State = !lwrfid.State

	} else {
		bcastStr = fmt.Sprintf("351,!%sF1|Ninja|On", lwrfid.RoomDevice)
		lwrfid.State = !lwrfid.State
	}

	broadcastMessage(bcastStr)
}

// SetState sets the state of a DeviceLWRF, given DeviceStruct
func SetState(lwrfid *DeviceStruct, state bool) {
	var bcastStr string
	if state {

		bcastStr = fmt.Sprintf("351,!%sF1|Ninja|On", lwrfid.RoomDevice)
		lwrfid.State = state

	} else {
		bcastStr = fmt.Sprintf("351,!%sF0|Ninja|Off", lwrfid.RoomDevice)
		lwrfid.State = state
	}

	broadcastMessage(bcastStr)
}

//Prepare the ports for packet tx and rx
func Prepare() (bool, error) {

	_, err := getLocalIP() // Get our local IP. Not actually used in this func, but is more of a failsafe
	if err != nil {        // Error? Return false
		return false, err
	}

	udpAddr, resolveErr := net.ResolveUDPAddr("udp4", ":9761") // Get our address ready for listening
	if resolveErr != nil {
		return false, resolveErr
	}

	var listenErr error
	conn, listenErr = net.ListenUDP("udp", udpAddr) // Now we listen on the address we just resolved
	if listenErr != nil {
		return false, listenErr
	}

	return true, nil
}
