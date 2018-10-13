package main

import (
	"errors"
	"flag"
	"log"
	"strconv"
	"strings"

	"github.com/carbeer/Peerster/gossiper"
)

var uiPort int
var gossipIp, name string
var gossipPort int
var peers []string
var simple bool
var quit chan bool

func main() {
	var tmp string
	flag.IntVar(&uiPort, "UIPort", 8080, "Port for the UI client")
	flag.StringVar(&name, "name", "Peer", "name of the gossiper")
	flag.BoolVar(&simple, "simple", false, "run gossiper in simple broadcast mode")
	flag.StringVar(&tmp, "gossipAddr", "127.0.0.1:5000", "ip:port for the gossiper")
	peerList := flag.String("peers", "", "comma seperated list of peers of the form ip:port")
	flag.Parse()

	validateAddress(tmp)
	peers = validatePeerList(*peerList)

	log.Println("UIPort has value", uiPort)
	log.Println("gossipAddr has value", tmp)
	log.Println("name has value", name)
	log.Println("peers has value", peers)
	log.Println("simple has value", simple)

	g := gossiper.NewGossiper(gossipIp, name, gossipPort, uiPort, peers)
	log.Println("Listenting for client messages")
	go g.ListenClientMessages(quit)
	log.Println("Listening for peer messages")
	go g.ListenPeerMessages(quit)
	<-quit

}

func validateAddress(address string) {
	var err error
	tmp := strings.Split(address, ":")

	if len(tmp) == 2 {
		gossipIp = tmp[0]
		if port, errPort := strconv.Atoi(tmp[1]); errPort == nil {
			gossipPort = port
			return
		} else {
			err = errPort
		}
	} else {
		err = errors.New("A valid address must have the format ip:port")
	}
	log.Println("An error occured: " + err.Error())
}

func validatePeerList(list string) []string {
	addresses := strings.Split(list, ",")
	for _, a := range addresses {
		validateAddress(a)
	}
	return addresses
}
