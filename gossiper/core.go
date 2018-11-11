package gossiper

import (
	"fmt"
	"net"
	"sync"

	"github.com/carbeer/Peerster/utils"
	"github.com/dedis/protobuf"
)

func (g *Gossiper) sendToPeer(gossipPacket utils.GossipPacket, targetIpPort string) {
	if targetIpPort == "" {
		fmt.Println("No target address given. WE DON'T SEEM TO KNOW THAT PEER YET.")
		return
	}
	if gossipPacket.Rumor != nil {
		if gossipPacket.Rumor.Text != "" {
			fmt.Printf("MONGERING with %s\n", targetIpPort)
		} else {
			fmt.Printf("Route mongering with %s\n", targetIpPort)
		}
	}
	byteStream, e := protobuf.Encode(&gossipPacket)
	utils.HandleError(e)
	adr, e := net.ResolveUDPAddr("udp4", targetIpPort)
	utils.HandleError(e)
	_, e = g.udpConn.WriteToUDP(byteStream, adr)
	utils.HandleError(e)
}

func (g *Gossiper) broadcastMessage(packet utils.GossipPacket, receivedFrom string) {
	var wg sync.WaitGroup
	// Broadcast to everyone except originPeer
	for _, p := range g.peers {
		if p != receivedFrom {
			wg.Add(1)
			go func(p string) {
				// fmt.Printf("%d: Broadcasting\n", time.Now().Second())
				g.sendToPeer(packet, p)
				wg.Done()
			}(p)
		}
	}
	wg.Wait()
}

func (g *Gossiper) addPeerToListIfApplicable(adr string) {
	for i := range g.peers {
		if g.peers[i] == adr {
			return
		}
	}
	g.peers = append(g.peers, adr)
}
