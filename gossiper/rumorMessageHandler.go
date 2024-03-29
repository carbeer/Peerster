package gossiper

import (
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/carbeer/Peerster/utils"
)

// Send out new RumorMessage as per client request
func (g *Gossiper) newRumorMongeringMessage(msg utils.Message) {
	rumorMessage := utils.RumorMessage{Origin: g.Name, ID: g.IdCounter, Text: msg.Text}
	fmt.Printf("New rumor mongering message %+v\n", rumorMessage)
	g.IdCounter = g.IdCounter + 1
	g.appendReceivedMessages(rumorMessage.Origin, rumorMessage)
	g.startRumorMongering(rumorMessage)
}

// Handles incoming RumorMessages
func (g *Gossiper) rumorMessageHandler(msg utils.RumorMessage, sender string) {
	// Check if its only a Route Rumor
	if msg.Text != "" {
		fmt.Printf("RUMOR origin %s from %s ID %d contents %s\n", msg.Origin, sender, msg.ID, msg.Text)
		fmt.Printf("PEERS %v\n", fmt.Sprint(strings.Join(g.Peers, ",")))
	} else {
		fmt.Printf("ROUTE RUMOR origin %s from %s ID %d\n", msg.Origin, sender, msg.ID)
	}
	origin := msg.Origin
	var wg sync.WaitGroup

	if origin != g.Name {
		g.updateNextHop(msg.Origin, msg.ID, sender)
		// Check whether the message is desired
		if len(g.getReceivedMessages(origin))+1 == int(msg.ID) {
			g.appendReceivedMessages(msg.Origin, msg)

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

// Pick random neighbor for mongering that is not `origin` and not already in a rumor mongering with the gossiper
func (g *Gossiper) pickRandomPeerForMongering(origin string) string {
	peer := ""
	for {
		newRand := rand.New(rand.NewSource(time.Now().UnixNano())).Intn(len(g.Peers))
		if g.Peers[newRand] != origin && g.getRumorMongeringChannel(peer) == nil {
			peer = g.Peers[newRand]
			break
		}
	}
	return peer
}

// Kicks of a rumor mongering connection and keeps track of closes bilateral connections
func (g *Gossiper) startRumorMongering(msg utils.RumorMessage) {
	gossipPacket := utils.GossipPacket{Rumor: &msg}
	peer := g.pickRandomPeerForMongering("")
	for {
		if peer == "" {
			break
		}
		// If it returns true, coinflip decided to continue mongering, else it stops
		if !g.startRumorMongeringConnection(peer, gossipPacket) {
			fmt.Println("Stopping rumormongering")
			break
		}
		peer = g.pickRandomPeerForMongering(peer)
		fmt.Println("FLIPPED COIN sending rumor to", peer)
	}
}

// Represents a bilateral rumor mongering connnection. Once its sended, the method returns a coin flip value
func (g *Gossiper) startRumorMongeringConnection(peer string, gossipPacket utils.GossipPacket) bool {
	// Create a channel that is added to the list of owned rumorMongergings
	g.setRumorMongeringChannel(peer, make(chan utils.StatusPacket, utils.MSG_BUFFER))
	fmt.Printf("%d: Initiating rumor mongering connection \n", time.Now().Second())
	g.sendToPeer(gossipPacket, peer)

Loop:
	for {
		select {
		case <-time.After(utils.RUNOR_TIMEOUT):
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
