package gossiper

import (
	"encoding/hex"
	"fmt"
	"time"

	"github.com/carbeer/Peerster/utils"
)

func (g *Gossiper) newRumorMongeringMessage(msg utils.Message) {
	rumorMessage := utils.RumorMessage{Origin: g.name, ID: g.idCounter, Text: msg.Text}
	g.idCounter = g.idCounter + 1
	g.addToKnownMessages(rumorMessage)
	g.startRumorMongering(rumorMessage)
}

func (g *Gossiper) newPrivateMessage(msg utils.Message) {
	privateMessage := utils.PrivateMessage{Origin: g.name, ID: 0, Text: msg.Text, Destination: msg.Destination, HopLimit: utils.GetHopLimitConstant()}
	gossipMessage := utils.GossipPacket{Private: &privateMessage}
	fmt.Printf("SENDING PRIVATE MESSAGE %s TO %s\n", msg.Text, msg.Destination)
	g.appendPrivateMessages(g.name, privateMessage)
	g.sendToPeer(gossipMessage, g.getNextHop(msg.Destination).Address)
}

func (g *Gossiper) sendDataRequest(msg utils.Message) {

	hash, e := hex.DecodeString(msg.Request)
	utils.HandleError(e)
	dataRequest := utils.DataRequest{Origin: g.name, Destination: msg.Destination, HopLimit: utils.GetHopLimitConstant(), HashValue: hash}
	gossipMessage := utils.GossipPacket{DataRequest: &dataRequest}
	if msg.FileName != "" {
		g.setStoredFile(msg.Request, utils.File{FileName: msg.FileName, MetaHash: msg.Request})
		g.addRequestedChunks(msg.Request, utils.ChunkInfo{FileName: msg.FileName})
		fmt.Printf("REQUESTING filename %s from %s hash %s\n", msg.FileName, msg.Destination, msg.Request)
	}

	response := make(chan bool, utils.GetMsgBuffer())
	g.setDataRequestChannel(msg.Request, response)
	for {
		g.sendToPeer(gossipMessage, g.getNextHop(msg.Destination).Address)
		select {
		case <-time.After(5 * time.Second):
			fmt.Printf("TIMEOUT\n")
			if g.getDataRequestChannel(msg.Request) == nil {
				fmt.Printf("Not relevant anymore\n")
				return
			}
			continue
		case <-response:
			fmt.Printf("Received chunk\n")
			g.deleteDataRequestChannel(msg.Request)
			return
		}
	}

}