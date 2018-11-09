package gossiper

import (
	"github.com/carbeer/Peerster/utils"
)

func (g *Gossiper) AntiEntropy() {
	var peer string
	for {
		timeout := make(chan bool)
		go utils.TimeoutCounter(timeout, "1s")
		<-timeout
		peer = g.pickRandomPeerForMongering("")
		go g.sendAcknowledgement(peer)
	}
}
