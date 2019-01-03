package main

import (
	"crypto/rsa"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/carbeer/Peerster/gossiper"
)

var uiPort int
var gossipIp, name string
var gossipPort int
var peers []string
var simple bool
var quit chan bool
var rtimer int
var runUI bool
var warmStart bool
var privKey *rsa.PrivateKey

func main() {
	flag.IntVar(&uiPort, "UIPort", 8080, "port for the UI client")
	flag.StringVar(&name, "name", "", "name/ public key of the gossiper (required for warmStartup). Leave empty to generate a new RSA keypair")
	flag.IntVar(&rtimer, "rtimer", 0, "route rumors sending period in seconds, 0 to disable (default 0)")
	tmp := flag.String("gossipAddr", "127.0.0.1:5000", "ip:port for the gossiper")
	peerList := flag.String("peers", "", "comma seperated list of peers of the form ip:port")
	flag.BoolVar(&simple, "simple", false, "run gossiper in simple broadcast mode")
	flag.BoolVar(&runUI, "runUI", false, "serve UI with this gossiper")
	flag.BoolVar(&warmStart, "warmStart", false, "load a previous state from the disk for a warm start up")
	flag.Parse()

	elems := strings.Split(*tmp, ":")
	gossipIp = elems[0]
	gossipPort, _ = strconv.Atoi(elems[1])
	peers = strings.Split(*peerList, ",")
	fmt.Println("peers has value", peers)
	fmt.Println("UIPort has value", uiPort)
	fmt.Println("gossipAddr has value", *tmp)
	fmt.Println("name has value", name)
	fmt.Println("simple has value", simple)
	fmt.Println("rtimer has value", rtimer)
	fmt.Println("runUI has value", runUI)
	fmt.Println("warmStart has value", warmStart)

	var g *gossiper.Gossiper
	if !warmStart {
		g = gossiper.NewGossiper(gossipIp, name, gossipPort, uiPort, peers, simple)
	} else {
		if name == "" {
			log.Println("Please provide a name or public key for startup")
			os.Exit(1)
		}
		g = gossiper.RestoreGossiper(gossipIp, name, gossipPort, uiPort, peers)
	}

	go g.ListenClientMessages()
	go g.ListenPeerMessages()

	if !simple {
		go g.AntiEntropy()
	}

	if rtimer != 0 {
		go g.RouteRumor(strconv.Itoa(rtimer) + "s")
	}

	if runUI {
		go g.BootstrapUI()
	}
	go g.MineBlocks()

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("Received a kill signal. Saving node state.")
		g.SaveState()
		log.Println("Saved state of node", name)
		os.Exit(1)
	}()
	// Block forever
	<-quit
}
