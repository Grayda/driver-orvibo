package main

import (
	"fmt"
	"github.com/Grayda/sphere-orvibo/allone"
	"time"
)

var prepared = false

func main() {

	allone.PrepareSockets()
	for {
		allone.CheckForMessages()
		select {
		case msg := <-allone.Events:
			handleMessage(msg)
		}

	}
}

func handleMessage(msg allone.EventStruct) {
	switch msg.Name {
	case "ready":
		allone.Discover()
	case "socketfound":
		fmt.Println("We have a socket!")
		time.Sleep(time.Millisecond * 100)
		allone.Subscribe()
	case "subscribed":
		fmt.Println("We're subscribed!")
		time.Sleep(time.Millisecond * 100)
		allone.Query()
	case "queried":
		fmt.Println("We've queried. Name is:", msg.SocketInfo.Name)
		allone.SetState(true, msg.SocketInfo.MACAddress)
		fmt.Println("Turned socket ON")
		time.Sleep(time.Second)
		allone.SetState(false, msg.SocketInfo.MACAddress)
		fmt.Println("Turned socket OFF")
	}
}
