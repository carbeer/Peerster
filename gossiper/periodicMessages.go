package gossiper

import (
	"fmt"
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
	fmt.Printf("Starting route rumors with frequency %fs\n", duration.Seconds())

	// Send initial route rumor messages instantaneously, afterwards periodically.
	for {
		rumorMessage := utils.RumorMessage{Origin: g.name, ID: g.idCounter}
		g.idCounter = g.idCounter + 1
		g.addToKnownMessages(rumorMessage)
		go g.startRumorMongering(rumorMessage)
		fmt.Println(time.Now().Second())
		<-time.After(duration)
		fmt.Println(time.Now().Second())

	}
}
