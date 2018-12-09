package gossiper

import (
	"fmt"
	"time"

	"github.com/carbeer/Peerster/utils"
)

func (g *Gossiper) sendDataRequest(msg utils.Message, destination string) {
	var dataRequest utils.DataRequest
	hash := utils.ByteMetaHash(msg.Request)
	dataRequest = utils.DataRequest{Origin: g.name, Destination: destination, HopLimit: utils.HOPLIMIT_CONSTANT, HashValue: hash}

	gossipMessage := utils.GossipPacket{DataRequest: &dataRequest}
	if msg.FileName != "" {
		g.setStoredFile(msg.Request, utils.File{Name: msg.FileName, MetafileHash: utils.ByteMetaHash(msg.Request)})
		g.addRequestedChunks(msg.Request, utils.ChunkInfo{FileName: msg.FileName})
		fmt.Printf("REQUESTING filename %s from %s hash %s\n", msg.FileName, destination, msg.Request)
	}

	response := make(chan bool, utils.MSG_BUFFER)
	g.setDataRequestChannel(msg.Request, response, (msg.Destination != ""))

	for {
		g.sendToPeer(gossipMessage, g.getNextHop(destination).Address)
		select {
		case <-time.After(utils.DATA_REQUEST_TIMEOUT):
			continue
		case <-response:
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
			return
		}
		gossipMessage := utils.GossipPacket{DataRequest: &msg}
		g.sendToPeer(gossipMessage, g.getNextHop(msg.Destination).Address)
	}
}
