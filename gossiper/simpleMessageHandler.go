package gossiper

import (
	"fmt"
	"strings"

	"github.com/carbeer/Peerster/utils"
)

func (g *Gossiper) simpleMessageHandler(msg utils.SimpleMessage) {
	fmt.Printf("SIMPLE MESSAGE origin %s from %s contents %s\n", msg.OriginalName, msg.RelayPeerAddr, msg.Contents)
	fmt.Printf("PEERS %v\n", fmt.Sprint(strings.Join(g.peers, ",")))
	// No need to broadcast one's own message anymore
	if msg.OriginalName == g.name {
		return
	}
	msg.RelayPeerAddr = g.Address.String()
	g.broadcastMessage(utils.GossipPacket{Simple: &msg}, msg.RelayPeerAddr)
}
