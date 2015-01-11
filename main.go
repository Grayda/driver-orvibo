package main

import (
	"fmt"

	"github.com/Grayda/sphere-orvibo/allone"
	"os"
	"os/signal"
	"time"
)

// For those playing along at home who have as little idea about driver development on the Sphere as I did:
// This driver is split into two sections. The first is "driver.go" which is all about the driver side of things --
// It starts the driver off, which starts searching for the sockets, looks for messages and such.
// When a socket is found, it makes a new device (in device.go) which handles the toggling of the sockets
// and setting of the name.

func main() {
	fmt.Println("Yo")
	// In case it crash loops =))
	time.Sleep(time.Second * 1)
	// Prepare here, because if Start is called again (which it was, during testing?), we'd get "address in use"
	allone.PrepareSockets()
	// Now that we're prepared, we can create our driver
	_, err := NewDriver()

	if err != nil {
		fmt.Println("Failed to create driver: %s", err)
		return
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)

	// Block until a signal is received.
	s := <-c
	fmt.Println("Got signal:", s)

}
