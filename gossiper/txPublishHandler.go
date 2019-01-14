package gossiper

import (
	"fmt"

	"github.com/carbeer/Peerster/utils"
)

// Handles TxPublish messages
func (g *Gossiper) txPublishHandler(msg utils.TxPublish, sender string) {
	if !g.inBlockHistory(msg) && !g.inPendingTransactions(msg) {
		g.addPendingTransaction(msg)
		msg.HopLimit--
		if msg.HopLimit > 0 {
			g.broadcastMessage(utils.GossipPacket{TxPublish: &msg}, sender)
		}
	} else {
		fmt.Printf("Discarding transaction %+v\n", msg)
	}
}
