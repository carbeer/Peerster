package gossiper

import (
	"fmt"
	"time"

	"github.com/carbeer/Peerster/utils"
)

// Send anti entropy messages periodically
func (g *Gossiper) AntiEntropy() {
	fmt.Printf("Starting anti entropy messages with frequency %fs\n", utils.ANTI_ENTROPY_FREQUENCY.Seconds())
	for {
		<-time.After(utils.ANTI_ENTROPY_FREQUENCY)
		peer := g.pickRandomPeerForMongering("")
		go g.sendAcknowledgement(peer)
	}
}

// Sends route rumor messages. First one instantly, subsequent ones periodically
func (g *Gossiper) RouteRumor(rtimer string) {
	duration, e := time.ParseDuration(rtimer)
	utils.HandleError(e)
	fmt.Printf("Starting route rumors with frequency %fs\n", duration.Seconds())
	for {
		rumorMessage := utils.RumorMessage{Origin: g.Name, ID: g.IdCounter}
		g.IdCounter = g.IdCounter + 1
		g.appendReceivedMessages(rumorMessage.Origin, rumorMessage)
		go g.startRumorMongering(rumorMessage)
		<-time.After(duration)
	}
}
