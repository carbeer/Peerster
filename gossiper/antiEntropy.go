package gossiper

import (
	"time"

	"github.com/carbeer/Peerster/utils"
)

func (g *Gossiper) AntiEntropy() {
	var peer string
	for {
		<-time.After(5 * utils.GetAntiEntropyFrequency())
		peer = g.pickRandomPeerForMongering("")
		go g.sendAcknowledgement(peer)
	}
}
