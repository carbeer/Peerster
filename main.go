package main

import (
	"errors"
	"flag"
	"fmt"
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
var runUI bool

func main() {
	flag.IntVar(&uiPort, "UIPort", 8080, "port for the UI client")
	flag.StringVar(&name, "name", "Peer", "name of the gossiper")
	flag.BoolVar(&simple, "simple", false, "run gossiper in simple broadcast mode")
	tmp := flag.String("gossipAddr", "127.0.0.1:5000", "ip:port for the gossiper")
	peerList := flag.String("peers", "", "comma seperated list of peers of the form ip:port")
	rtimer := strconv.Itoa(*flag.Int("rtimer", 0, "route rumors sending period in seconds, 0 to disable")) + "s"
	flag.BoolVar(&runUI, "runUI", false, "serve UI with this gossiper")
	flag.Parse()

	validateAddress(*tmp)
	elems := strings.Split(*tmp, ":")
	gossipIp = elems[0]
	gossipPort, _ = strconv.Atoi(elems[1])
	peers = validatePeerList(*peerList)

	fmt.Println("UIPort has value", uiPort)
	fmt.Println("gossipAddr has value", *tmp)
	fmt.Println("name has value", name)
	fmt.Println("peers has value", peers)
	fmt.Println("simple has value", simple)
	fmt.Println("rtimer has value", rtimer)
	fmt.Println("runUI has value", runUI)

	g := gossiper.NewGossiper(gossipIp, name, gossipPort, uiPort, peers, simple)

	go g.ListenClientMessages(quit)
	go g.ListenPeerMessages(quit)

	if !simple {
		go g.AntiEntropy()
	}

	if rtimer != "0s" {
		go g.RouteRumor(rtimer)
	}

	if runUI {
		go g.BootstrapUI()
	}

	<-quit

}

func validateAddress(address string) {
	var err error
	tmp := strings.Split(address, ":")

	if len(tmp) == 2 {
		if _, errPort := strconv.Atoi(tmp[1]); errPort == nil {
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
