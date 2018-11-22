package gossiper

import (
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/carbeer/Peerster/utils"
)

func (g *Gossiper) rumorMessageHandler(msg utils.RumorMessage, sender string) {
	if msg.Text != "" {
		fmt.Printf("RUMOR origin %s from %s ID %d contents %s\n", msg.Origin, sender, msg.ID, msg.Text)
		fmt.Printf("PEERS %v\n", fmt.Sprint(strings.Join(g.peers, ",")))
	} else {
		fmt.Printf("ROUTE RUMOR origin %s from %s ID %d\n", msg.Origin, sender, msg.ID)
	}
	origin := msg.Origin
	var wg sync.WaitGroup

	// Check whether the message is desired
	if origin != g.name {
		g.updateNextHop(msg, sender)
		if len(g.getReceivedMessages(origin))+1 == int(msg.ID) {
			g.addToKnownMessages(msg)

			wg.Add(1)
			go func() {
				g.startRumorMongering(msg)
				wg.Done()
			}()
		}
	}
	g.sendAcknowledgement(sender)
	wg.Wait()
}

func (g *Gossiper) addToKnownMessages(msg utils.RumorMessage) {
	g.appendReceivedMessages(msg.Origin, msg)
}

func (g *Gossiper) updateNextHop(msg utils.RumorMessage, sender string) {
	if msg.ID > g.getNextHop(msg.Origin).HighestID {
		g.setNextHop(msg.Origin, utils.HopInfo{Address: sender, HighestID: msg.ID})
		fmt.Printf("DSDV %s %s\n", msg.Origin, sender)
	}
}

func (g *Gossiper) pickRandomPeerForMongering(origin string) string {
	peer := ""
	for {
		newRand := rand.New(rand.NewSource(time.Now().UnixNano())).Intn(len(g.peers))
		if g.peers[newRand] != origin && g.getRumorMongeringChannel(peer) == nil {
			peer = g.peers[newRand]
			break
		}
	}
	return peer
}

func (g *Gossiper) startRumorMongering(msg utils.RumorMessage) {
	gossipPacket := utils.GossipPacket{Rumor: &msg}
	peer := g.pickRandomPeerForMongering("")
	for {
		if peer == "" {
			break
		}
		// If it return true, coinflip decided to continue mongering, else it stops
		if !g.startRumorMongeringConnection(peer, gossipPacket) {
			fmt.Println("Stopping rumormongering")
			break
		}
		peer = g.pickRandomPeerForMongering(peer)
		fmt.Println("FLIPPED COIN sending rumor to", peer)
	}
}

func (g *Gossiper) startRumorMongeringConnection(peer string, gossipPacket utils.GossipPacket) bool {
	// Create a channel that is added to the list of owned rumorMongergings
	g.setRumorMongeringChannel(peer, make(chan utils.StatusPacket, utils.GetMsgBuffer()))
	fmt.Printf("%d: Initiating rumor mongering connection \n", time.Now().Second())
	g.sendToPeer(gossipPacket, peer)

Loop:
	for {
		select {
		case <-time.After(utils.GetRumorMongeringTimeout()):
			fmt.Printf("%d: TIMEOUT\n", time.Now().Second())
			break Loop
		case status := <-g.getRumorMongeringChannel(peer):
			if g.compareStatus(peer, status) {
				// The two peers are in sync
				fmt.Printf("IN SYNC WITH %s\n", peer)
				break Loop
			}
		}
	}
	g.deleteRumorMongeringChannel(peer)
	return utils.FlipCoin()
}
