package gossiper

import (
	"fmt"
	"time"

	"github.com/carbeer/Peerster/utils"
)

func (g *Gossiper) AntiEntropy() {
	fmt.Printf("Starting anti entropy messages with frequency %fs\n", utils.ANTI_ENTROPY_FREQUENCY.Seconds())
	for {
		<-time.After(utils.ANTI_ENTROPY_FREQUENCY)
		peer := g.pickRandomPeerForMongering("")
		go g.sendAcknowledgement(peer)
	}
}

func (g *Gossiper) RouteRumor(rtimer string) {
	duration, e := time.ParseDuration(rtimer)
	utils.HandleError(e)
	fmt.Printf("Starting route rumors with frequency %fs\n", duration.Seconds())

	// Send initial route rumor messages instantaneously, afterwards periodically.
	for {
		rumorMessage := utils.RumorMessage{Origin: g.name, ID: g.idCounter}
		g.idCounter = g.idCounter + 1
		g.appendReceivedMessages(rumorMessage.Origin, rumorMessage)
		go g.startRumorMongering(rumorMessage)
		<-time.After(duration)
	}
}
