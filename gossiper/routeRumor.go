package gossiper

import (
	"time"

	"github.com/carbeer/Peerster/utils"
)

func (g *Gossiper) RouteRumor(rtimer string) {
	duration, e := time.ParseDuration(rtimer)
	utils.HandleError(e)
	// Setup route rumor message
	rumorMessage := utils.RumorMessage{Origin: g.name, ID: g.idCounter}
	g.idCounter = g.idCounter + 1
	g.addToKnownMessages(rumorMessage)
	go g.startRumorMongering(rumorMessage)

	// Continuous route rumor messages
	for {
		<-time.After(duration)
		rumorMessage := utils.RumorMessage{Origin: g.name, ID: g.idCounter}
		g.idCounter = g.idCounter + 1
		g.addToKnownMessages(rumorMessage)
		go g.startRumorMongering(rumorMessage)
	}
}
