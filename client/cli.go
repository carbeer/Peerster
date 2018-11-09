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
var dest string
var ip string = "127.0.0.1"

func main() {
	flag.IntVar(&uiPort, "UIPort", 8080, "port for the UI client")
	flag.StringVar(&msg, "msg", "", "message to be sent")
	flag.StringVar(&dest, "dest", "", "destination for the private message")

	flag.Parse()
	log.Println("UIPort has value", uiPort)
	log.Println("msg has value", msg)
	log.Println("dest has value", dest)

	if msg != "" && dest != "" {
		SendPrivateMessage(msg, uiPort, dest)
	} else {
		SendMessage(msg, uiPort)
	}

}

func SendMessage(msg string, uiPort int) {
	message := utils.Message{Text: msg}

	log.Println("Encoding the message")
	packetBytes, e := protobuf.Encode(&message)
	utils.HandleError(e)

	log.Println("Creating a client connection")
	udpConn, e := net.Dial("udp4", fmt.Sprintf("%s:%d", ip, uiPort))
	utils.HandleError(e)

	log.Println("Writing the message")
	_, e = udpConn.Write(packetBytes)
	utils.HandleError(e)
}

func SendPrivateMessage(msg string, uiPort int, dest string) {
	message := utils.Message{Text: msg, Destination: dest}

	log.Println("Encoding the message")
	packetBytes, e := protobuf.Encode(&message)
	utils.HandleError(e)

	log.Println("Creating a client connection")
	udpConn, e := net.Dial("udp4", fmt.Sprintf("%s:%d", ip, uiPort))
	utils.HandleError(e)

	log.Println("Writing the message")
	_, e = udpConn.Write(packetBytes)
	utils.HandleError(e)
}
