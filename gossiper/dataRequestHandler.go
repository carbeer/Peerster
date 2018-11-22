package gossiper

import (
	"encoding/hex"
	"fmt"
	"time"

	"github.com/carbeer/Peerster/utils"
)

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

func (g *Gossiper) dataRequestHandler(msg utils.DataRequest, sender string) {
	if msg.Destination == g.name {
		g.newDataReplyMessage(msg, sender)
	} else {
		msg.HopLimit -= 1
		if msg.HopLimit <= 0 {
			// log.Printf("%s: ATTENTION: Dropping a private message for %s\n", g.name, msg.Destination)
			return
		}
		gossipMessage := utils.GossipPacket{DataRequest: &msg}
		// fmt.Printf("%d: Send the private message\n", time.Now().Second())
		g.sendToPeer(gossipMessage, g.getNextHop(msg.Destination).Address)
	}
}
