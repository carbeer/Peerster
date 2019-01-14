package gossiper

import (
	"errors"
	"fmt"
	"time"

	"github.com/carbeer/Peerster/utils"
)

// Download a file as per client request
func (g *Gossiper) DownloadFile(msg utils.Message) error {
	var err error
	g.fileDownloadChannel[msg.Request] = make(chan bool, 1024)
	if msg.Destination != "" {
		g.sendDataRequest(msg, msg.Destination)
	} else {
		g.sendDataRequest(msg, g.getChunkHolder(msg.Request, 0))
	}

	timeout := time.After(utils.PRIV_FILE_REQ_TIMEOUT)

	select {
	case <-timeout:
		err = errors.New("Timeout, couldn't download file")
		break
	case <-g.fileDownloadChannel[msg.Request]:
		break
	}
	delete(g.fileDownloadChannel, msg.Request)
	return err
}

// Send out new DataRequest message
func (g *Gossiper) sendDataRequest(msg utils.Message, destination string) {
	var dataRequest utils.DataRequest
	hash := utils.ByteMetaHash(msg.Request)
	dataRequest = utils.DataRequest{Origin: g.Name, Destination: destination, HopLimit: utils.HOPLIMIT_CONSTANT, HashValue: hash}

	gossipMessage := utils.GossipPacket{DataRequest: &dataRequest}
	if msg.FileName != "" {
		g.setStoredFile(msg.Request, utils.File{Name: msg.FileName, MetafileHash: utils.ByteMetaHash(msg.Request)})
		g.addRequestedChunks(msg.Request, utils.ChunkInfo{FileName: msg.FileName})
		fmt.Printf("REQUESTING fileName %s from %s hash %s\n", msg.FileName, destination, msg.Request)
	}

	response := make(chan bool, utils.MSG_BUFFER)
	val := msg.Destination != ""
	g.setDataRequestChannel(msg.Request, response, val)

	// Repeat request if the peer doesn't respond
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

// Handles DataRequest messages
func (g *Gossiper) dataRequestHandler(msg utils.DataRequest, sender string) {
	if msg.Destination == g.Name {
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
