package gossiper

import (
	"fmt"
	"strings"
	"time"

	"github.com/carbeer/Peerster/utils"
)

func (g *Gossiper) statusMessageHandler(status utils.StatusPacket, sender string) {

	fmt.Printf("STATUS from %s%s\n", sender, status.ToString())
	fmt.Printf("PEERS %v\n", fmt.Sprint(strings.Join(g.peers, ",")))

	if g.getRumorMongeringChannel(sender) != nil {
		fmt.Printf("Owner of the channel with %s\n", sender)
		g.sendToRumorMongeringChannel(sender, status)
	} else {
		if g.compareStatus(sender, status) {
			fmt.Printf("Not the owner of the channel with %s (IN SYNC)\n", sender)
		} else {
			fmt.Printf("Not the owner of the channel with %s (NOT IN SYNC)\n", sender)
		}
	}
}

// returns true if status is equal, else false
func (g *Gossiper) compareStatus(peer string, status utils.StatusPacket) bool {
	// Send out additional messages that were requested
	if new, newMsg := g.AdditionalMessages(status); new {
		fmt.Printf("%d: Sending one of the unknown messages: %s %d\n", time.Now().Second(), newMsg.Origin, newMsg.ID)
		g.sendToPeer(utils.GossipPacket{Rumor: &newMsg}, peer)
		return false
	} else if g.HasLessMessagesThan(status) {
		// If no new messages requested, check whether peer has unknown messages and request them
		fmt.Printf("%d: Sending acknowledgement to %s to get unknown message\n", time.Now().Second(), peer)
		g.sendAcknowledgement(peer)
		return false
	}
	return true
}

// Checks whether this gossiper could get additional messages from the peer that sent the status packet
func (g *Gossiper) HasLessMessagesThan(status utils.StatusPacket) bool {
	for i := range status.Want {
		id := status.Want[i].Identifier
		// Check if Origin and IDs all known
		if len(g.getReceivedMessages(id))+1 < int(status.Want[i].NextID) {
			fmt.Println("Have less message for ORIGIN", id)
			return true
		}
	}
	return false
}

// Checks whether this gossiper has additional messages for the peer that sent the status packet
func (g *Gossiper) AdditionalMessages(status utils.StatusPacket) (bool, utils.RumorMessage) {
	g.receivedMessagesLock.RLock()
	defer g.receivedMessagesLock.RUnlock()

Loop:
	for k, v := range g.ReceivedMessages {
		for index := range status.Want {
			if status.Want[index].Identifier == k {
				if int(status.Want[index].NextID) < len(v)+1 {
					return true, v[status.Want[index].NextID-1]
				}
				continue Loop
			}
		}
		return true, g.ReceivedMessages[k][0]
	}
	return false, utils.RumorMessage{}
}

func (g *Gossiper) generateStatusPacket() utils.StatusPacket {
	packet := utils.StatusPacket{}
	g.receivedMessagesLock.RLock()
	for k, v := range g.ReceivedMessages {
		peer := utils.PeerStatus{Identifier: k, NextID: uint32(len(v) + 1)}
		packet.Want = append(packet.Want, peer)
	}
	g.receivedMessagesLock.RUnlock()
	return packet
}

func (g *Gossiper) sendAcknowledgement(adr string) {
	// Don't send ack to self
	if adr == g.Address.String() {
		return
	}
	statusPacket := g.generateStatusPacket()
	gossipPacket := utils.GossipPacket{Status: &statusPacket}
	g.sendToPeer(gossipPacket, adr)
}
