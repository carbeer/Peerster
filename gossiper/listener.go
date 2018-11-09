package gossiper

import (
	"fmt"
	"strings"
	"sync"

	"github.com/carbeer/Peerster/utils"
	"github.com/dedis/protobuf"
)

func (g *Gossiper) ListenClientMessages(quit chan bool) {
	// Listen infinitely long
	for {
		buffer := make([]byte, 4096)
		n, _, e := g.ClientConn.ReadFromUDP(buffer)
		utils.HandleError(e)
		msg := &utils.Message{}
		// TODO: Larger Messages!
		e = protobuf.Decode(buffer[:n], msg)
		utils.HandleError(e)
		fmt.Println("CLIENT MESSAGE", msg.Text)
		fmt.Printf("PEERS %v\n", fmt.Sprint(strings.Join(g.peers, ",")))

		go g.ClientMessageHandler(*msg)
		// broadcast message to all peers
	}
	quit <- true
}

func (g *Gossiper) ListenPeerMessages(quit chan bool) {
	for {
		buffer := make([]byte, 4096)
		n, peerAddr, e := g.udpConn.ReadFromUDP(buffer)
		// fmt.Println("Peeraddr", peerAddr.String())
		// fmt.Println(fmt.Sprintf("%s:%d", peerAddr.IP, peerAddr.Port))
		utils.HandleError(e)

		msg := &utils.GossipPacket{}
		e = protobuf.Decode(buffer[:n], msg)
		utils.HandleError(e)
		go g.peerMessageHandler(*msg, peerAddr.String())
	}
	quit <- true
}

func (g *Gossiper) ClientMessageHandler(msg utils.Message) {
	var wg sync.WaitGroup
	var gossipPacket utils.GossipPacket

	// Just broadcast using a simple message
	if g.simple {
		simpleMessage := utils.SimpleMessage{OriginalName: g.name, RelayPeerAddr: g.Address.String(), Contents: msg.Text}
		gossipPacket = utils.GossipPacket{Simple: &simpleMessage}
		for _, p := range g.peers {
			wg.Add(1)
			go func(p string) {
				g.sendToPeer(gossipPacket, p)
				wg.Done()
			}(p)
		}

	} else {
		if msg.Destination == "" {
			// Make a rumor message out of it
			g.newRumorMongeringMessage(msg.Text)
		} else if msg.Text != "" {
			g.newPrivateMessage(msg)
		}
	}
	wg.Wait()
}

// TODO: Handle two non-nil and three nil cases!
func (g *Gossiper) peerMessageHandler(msg utils.GossipPacket, sender string) {
	g.addPeerToListIfApplicable(sender)
	if msg.Simple != nil {
		g.simpleMessageHandler(*msg.Simple)
	} else if msg.Rumor != nil {
		g.rumorMessageHandler(*msg.Rumor, sender)
	} else if msg.Status != nil {
		g.statusMessageHandler(*msg.Status, sender)
	}
}
