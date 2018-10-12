package client

import (
	"flag"
	"fmt"
)

var uiPort int
var msg string

func main() {
	flag.IntVar(&uiPort, "UIPort", 8080, "Port for the UI client")
	flag.StringVar(&msg, "msg", "", "Message to be sent")
	flag.Parse()

	fmt.Println("UIPort has value", uiPort)
	fmt.Println("msg has value", msg)
}

func sendMessage() {
	// Send message via UDP
}