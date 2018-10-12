package main

import (
	"errors"
	"flag"
	"fmt"
	"strconv"
	"strings"
)

var uiPort int
var gossipAddr, name string
var peers []string
var simple bool

func main() {
	flag.IntVar(&uiPort, "UIPort", 8080, "Port for the UI client")
	flag.StringVar(&name, "name", "Peer", "name of the gossiper")
	flag.BoolVar(&simple, "simple", false, "run gossiper in simple broadcast mode")
	flag.StringVar(&gossipAddr, "gossipAddr", "127.0.0.1:5000", "ip:port for the gossiper")
	peerList := flag.String("peers", "", "comma seperated list of peers of the form ip:port")
	flag.Parse()

	validateAddress(gossipAddr)
	peers = validatePeerList(*peerList)

	fmt.Println("UIPort has value ", uiPort)
	fmt.Println("gossipAddr has value ", gossipAddr)
	fmt.Println("name has value ", name)
	fmt.Println("peers has value ", peers)
	fmt.Println("simple has value ", simple)
}

func validateAddress(address string) {
	var err error
	tmp := strings.Split(address, ":")

	if len(tmp) == 2 {
		if _, errIp := strconv.Atoi(tmp[0]); errIp == nil {
			if _, errPort := strconv.Atoi(tmp[1]); errPort == nil {
				return
			} else {
				err = errPort
			}
		} else {
			err = errIp
		}
	} else {
		err = errors.New("A valid address must have the format ip:port")
	}
	fmt.Println("An error occured: " + err.Error())
}

func validatePeerList(list string) []string {
	addresses := strings.Split(list, ",")
	for _, a := range addresses {
		validateAddress(a)
	}
	return addresses
}

func receivedClientMessage(msg SimpleMessage) {
	// Broadcast to all clients
	fmt.Printf("SIMPLE MESSAGE origin %v from %v contents %v", msg.OriginalName, msg.RelayPeerAddr, msg.Contents)
}

func receivedPeerMessage(msg SimpleMessage) {
	// Implement
	fmt.Printf("SIMPLE MESSAGE origin %v from %v contents %v", msg.OriginalName, msg.RelayPeerAddr, msg.Contents)

}
