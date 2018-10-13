package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/carbeer/Peerster/utils"
	"github.com/dedis/protobuf"
)

var uiPort int
var msg string
var ip string = "127.0.0.1"

func main() {
	flag.IntVar(&uiPort, "UIPort", 8080, "Port for the UI client")
	flag.StringVar(&msg, "msg", "", "Message to be sent")
	flag.Parse()

	log.Println("UIPort has value", uiPort)
	log.Println("msg has value", msg)
	sendMessage(msg, uiPort)
}

func sendMessage(msg string, uiPort int) {
	message := utils.Message{Text: msg}

	log.Println("Encoding the message")
	packetBytes, e := protobuf.Encode(&message)
	if e != nil {
		log.Fatal(e)
	}

	log.Println("Creating a client connection")
	udpConn, e := net.Dial("udp4", fmt.Sprintf("%s:%d", ip, uiPort))
	if e != nil {
		log.Fatal(e)
	}

	log.Println("Writing the message")
	_, e = udpConn.Write(packetBytes)
	if e != nil {
		log.Fatal(e)
	}
}
