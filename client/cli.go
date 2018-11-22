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
var file string
var request string
var keywords string
var budget int

func main() {
	flag.IntVar(&uiPort, "UIPort", 8080, "port for the UI client")
	flag.StringVar(&msg, "msg", "", "message to be sent")
	flag.StringVar(&dest, "dest", "", "destination for the private message")
	flag.StringVar(&file, "file", "", "file to be indexed by the gossiper")
	flag.StringVar(&request, "request", "", "request a chunk or metafile of this hash")
	flag.StringVar(&keywords, "keywords", "", "comma separated list of keywords")
	flag.IntVar(&budget, "budget", 2, "budget for keyword search")
	flag.Parse()

	/*
		log.Println("UIPort has value", uiPort)
		log.Println("msg has value", msg)
		log.Println("dest has value", dest)
		log.Println("file has value", file)
		log.Println("request has value", request)
	*/

	if file != "" {
		if request != "" && dest != "" {
			log.Println("Sending file download request")
			SendMessage(utils.Message{FileName: file, Request: request, Destination: dest}, uiPort)
		} else {
			log.Println("Sending file indexing request")
			SendMessage(utils.Message{FileName: file}, uiPort)
		}
	} else if msg != "" && dest != "" {
		log.Println("Sending private message")
		SendMessage(utils.Message{Text: msg, Destination: dest}, uiPort)
	} else if {
		log.Println("Sending search request")
		SendMessage(utils.Message{Keywords: keywords.split(","), Budget: budget})
	} else {
		log.Println("Sending normal message")
		SendMessage(utils.Message{Text: msg}, uiPort)
	}
}

func SendMessage(message utils.Message, uiPort int) {
	// log.Println("Encoding the message")
	packetBytes, e := protobuf.Encode(&message)
	utils.HandleError(e)

	// log.Println("Creating a client connection")
	udpConn, e := net.Dial("udp4", fmt.Sprintf("%s:%d", utils.GetClientIp(), uiPort))
	utils.HandleError(e)

	// log.Println("Writing the message")
	_, e = udpConn.Write(packetBytes)
	utils.HandleError(e)
}
