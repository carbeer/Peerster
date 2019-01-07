package gossiper

import (
	"fmt"

	"github.com/carbeer/Peerster/utils"
)

func (g *Gossiper) newEncryptedPrivateMessage(msg utils.Message) {
	// Store message unencrypted locally
	privateMessage := utils.PrivateMessage{Origin: g.Name, ID: 0, EncryptedText: msg.Text, Destination: msg.Destination, HopLimit: utils.HOPLIMIT_CONSTANT}
	fmt.Printf("SENDING PRIVATE ENCRYPTED MESSAGE %s TO %s\n", msg.Text, msg.Destination)
	g.appendPrivateMessages(g.Name, privateMessage)
	// Encrypt prior to sending
	privateMessage.EncryptedText = RSAEncryptText(msg.Destination, msg.Text)
	gossipMessage := utils.GossipPacket{Private: &privateMessage}
	g.sendToPeer(gossipMessage, g.getNextHop(msg.Destination).Address)
}

func (g *Gossiper) newPrivateMessage(msg utils.Message) {
	privateMessage := utils.PrivateMessage{Origin: g.Name, ID: 0, Text: msg.Text, Destination: msg.Destination, HopLimit: utils.HOPLIMIT_CONSTANT}
	gossipMessage := utils.GossipPacket{Private: &privateMessage}
	fmt.Printf("SENDING PRIVATE MESSAGE %s TO %s\n", msg.Text, msg.Destination)
	g.appendPrivateMessages(g.Name, privateMessage)
	g.sendToPeer(gossipMessage, g.getNextHop(msg.Destination).Address)
}

func (g *Gossiper) privateMessageHandler(msg utils.PrivateMessage) {
	if msg.Destination == g.Name {
		if msg.EncryptedText != "" {
			msg.EncryptedText = g.RSADecryptText(msg.EncryptedText)
			fmt.Printf("PRIVATE ENCRYPTED origin %s hop-limit %d contents %s\n", msg.Origin, msg.HopLimit, msg.EncryptedText)
		}
		if msg.Text != "" {
			fmt.Printf("PRIVATE origin %s hop-limit %d contents %s\n", msg.Origin, msg.HopLimit, msg.Text)
		}
		g.appendPrivateMessages(msg.Origin, msg)
	} else {
		msg.HopLimit -= 1
		if msg.HopLimit <= 0 {
			return
		}
		gossipMessage := utils.GossipPacket{Private: &msg}
		g.sendToPeer(gossipMessage, g.getNextHop(msg.Destination).Address)
	}
}
