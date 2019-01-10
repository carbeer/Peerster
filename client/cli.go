package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/carbeer/Peerster/utils"
	"github.com/dedis/protobuf"
)

var uiPort int
var msg string
var dest string
var file string
var request string
var keywords string
var budget int64
var encrypted bool
var privateFile bool
var replications int
var state string

func main() {
	flag.IntVar(&uiPort, "UIPort", 8080, "port for the UI client")
	flag.StringVar(&msg, "msg", "", "message to be sent")
	flag.StringVar(&dest, "dest", "", "destination for the private message")
	flag.StringVar(&file, "file", "", "file to be indexed by the gossiper")
	flag.StringVar(&request, "request", "", "request a chunk or metafile of this hash")
	flag.StringVar(&keywords, "keywords", "", "comma separated list of keywords")
	flag.Int64Var(&budget, "budget", -1, "budget for keyword search")
	flag.BoolVar(&encrypted, "encrypt", false, "encrypt private message with the public key of the destination")
	flag.BoolVar(&privateFile, "private", false, "index file privately")
	flag.IntVar(&replications, "replications", 1, "number of replications for private file uploads")
	flag.StringVar(&state, "state", "", "upload state")
	flag.Parse()

	if file != "" {
		if request != "" {
			if encrypted {
				log.Println("Sending private download request")
				SendMessage(utils.Message{FileName: file, Request: request, Encrypted: true}, uiPort)
			} else if dest != "" {
				log.Println("Sending file download request")
				SendMessage(utils.Message{FileName: file, Request: request, Destination: dest}, uiPort)
			} else {
				log.Println("Sending file download request without dest")
				SendMessage(utils.Message{FileName: file, Request: request}, uiPort)
			}
		} else {
			if privateFile {
				log.Println("Sending private file indexing request")
				if replications == 0 {
					replications = -1
				}
				SendMessage(utils.Message{FileName: file, Replications: replications}, uiPort)
			} else {
				log.Println("Sending file indexing request")
				SendMessage(utils.Message{FileName: file}, uiPort)
			}
		}
	} else if msg != "" && dest != "" {
		log.Println("Sending private message")
		SendMessage(utils.Message{Text: msg, Destination: dest, Encrypted: encrypted}, uiPort)
	} else if keywords != "" {
		log.Println("Sending search request")
		SendMessage(utils.Message{Keywords: strings.Split(keywords, ","), Budget: budget}, uiPort)
	} else if msg != "" {
		log.Println("Sending normal message")
		SendMessage(utils.Message{Text: msg}, uiPort)
	} else if state != "" {
		log.Println("Uploading state")
		SendMessage(utils.Message{FileName: state, Encrypted: true}, uiPort)
	}
}

func SendMessage(message utils.Message, uiPort int) {
	packetBytes, e := protobuf.Encode(&message)
	utils.HandleError(e)

	udpConn, e := net.Dial("udp4", fmt.Sprintf("%s:%d", utils.CLIENT_IP, uiPort))
	utils.HandleError(e)

	_, e = udpConn.Write(packetBytes)
	utils.HandleError(e)
}
