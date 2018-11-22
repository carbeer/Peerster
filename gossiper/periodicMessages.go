package gossiper

import (
	"time"

	"github.com/carbeer/Peerster/utils"
)

func (g *Gossiper) AntiEntropy() {
	var peer string
	for {
		<-time.After(utils.GetAntiEntropyFrequency())
		peer = g.pickRandomPeerForMongering("")
		go g.sendAcknowledgement(peer)
	}
}

func (g *Gossiper) RouteRumor(rtimer string) {
	duration, e := time.ParseDuration(rtimer)
	utils.HandleError(e)

	// Send initial route rumor messages instantaneously, afterwards periodically.
	for {
		rumorMessage := utils.RumorMessage{Origin: g.name, ID: g.idCounter}
		g.idCounter = g.idCounter + 1
		g.addToKnownMessages(rumorMessage)
		go g.startRumorMongering(rumorMessage)
		<-time.After(duration)
	}
}
