package gossiper

import (
	"log"

	"github.com/carbeer/Peerster/utils"
)

func (g *Gossiper) RouteRumor(rtimer string) {
	log.Println("Initializing RouteRumors")
	// Setup route rumor message
	rumorMessage := utils.RumorMessage{Origin: g.name, ID: g.idCounter}
	g.idCounter = g.idCounter + 1
	g.addToKnownMessages(rumorMessage)
	go g.startRumorMongering(rumorMessage)

	// Continuous route rumor messages
	for {
		timeout := make(chan bool)
		go utils.TimeoutCounter(timeout, rtimer)
		<-timeout
		rumorMessage := utils.RumorMessage{Origin: g.name, ID: g.idCounter}
		g.idCounter = g.idCounter + 1
		log.Println("Sending route rumor")
		g.addToKnownMessages(rumorMessage)
		go g.startRumorMongering(rumorMessage)
	}
}
