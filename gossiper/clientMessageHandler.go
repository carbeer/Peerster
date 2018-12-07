package gossiper

import (
	"crypto"
	"encoding/hex"
	"fmt"
	"math"
	"os"
	"strings"
	"time"

	"github.com/carbeer/Peerster/utils"
)

func (g *Gossiper) newRumorMongeringMessage(msg utils.Message) {
	rumorMessage := utils.RumorMessage{Origin: g.name, ID: g.idCounter, Text: msg.Text}
	fmt.Printf("New rumor mongering message %+v\n", rumorMessage)
	g.idCounter = g.idCounter + 1
	g.appendReceivedMessages(rumorMessage.Origin, rumorMessage)
	g.startRumorMongering(rumorMessage)
}

func (g *Gossiper) newPrivateMessage(msg utils.Message) {
	privateMessage := utils.PrivateMessage{Origin: g.name, ID: 0, Text: msg.Text, Destination: msg.Destination, HopLimit: utils.HOPLIMIT_CONSTANT}
	gossipMessage := utils.GossipPacket{Private: &privateMessage}
	fmt.Printf("SENDING PRIVATE MESSAGE %s TO %s\n", msg.Text, msg.Destination)
	g.appendPrivateMessages(g.name, privateMessage)
	g.sendToPeer(gossipMessage, g.getNextHop(msg.Destination).Address)
}

func (g *Gossiper) newSearchRequest(msg utils.Message) {
	var searchRequest utils.SearchRequest
	var timeout <-chan time.Time
	fulfilled := uint32(0)
	defaultBudget := false

	if msg.Budget < 0 {
		defaultBudget = true
		searchRequest = utils.SearchRequest{Origin: g.name, Budget: utils.DEFAULT_BUDGET, Keywords: msg.Keywords}
	} else {
		searchRequest = utils.SearchRequest{Origin: g.name, Budget: uint64(msg.Budget), Keywords: msg.Keywords}
	}
	// Create a channel that is added to the list of searchRequests
	g.setSearchRequestChannel(searchRequest, make(chan uint32, utils.MSG_BUFFER))
	g.searchRequestHandler(searchRequest, false)
	fmt.Printf("SEARCHING for keywords %s with budget %d\n", strings.Join(searchRequest.Keywords, ","), searchRequest.Budget)

	if defaultBudget {
		timeout = time.After(utils.SEARCH_REQUEST_TIMEOUT)
	}
Loop:
	for {
		select {
		case <-timeout:
			fmt.Printf("%d: TIMEOUT, DOUBLING BUDGET\n", time.Now().Second())
			if searchRequest.Budget == utils.MAX_BUDGET {
				fmt.Printf("Reached maximum budget. Stop searching.\n")
				break Loop
			}
			searchRequest.Budget = utils.MinUint64(searchRequest.Budget*2, utils.MAX_BUDGET)
			fmt.Printf("SEARCHING for keywords %s with budget %d\n", strings.Join(searchRequest.Keywords, ","), searchRequest.Budget)
			g.searchRequestHandler(searchRequest, true)
			timeout = time.After(utils.SEARCH_REQUEST_TIMEOUT)
		case new := <-g.getSearchRequestChannel(searchRequest):
			fulfilled += new
			// Process message
			if fulfilled >= utils.MIN_THRESHOLD {
				fmt.Printf("SEARCH FINISHED\n")
				break Loop
			}
		}
	}
	g.deleteSearchRequestChannel(searchRequest)
}

func (g *Gossiper) fulfilledQuery(msg utils.SearchRequest) bool {
	// 2 files were found and all the chunks were found at at least one peer

	if 0 == 0 {
		return true
	} else if msg.Budget >= utils.MAX_BUDGET {
		return true
	}
	return false
}

func (g *Gossiper) newDataReplyMessage(msg utils.DataRequest, sender string) {
	dataReplyMessage := utils.DataReply{Origin: g.name, Destination: msg.Origin, HopLimit: utils.HOPLIMIT_CONSTANT, HashValue: msg.HashValue, Data: g.getStoredChunk(hex.EncodeToString(msg.HashValue))}
	gossipMessage := utils.GossipPacket{DataReply: &dataReplyMessage}

	nextHop := g.getNextHop(dataReplyMessage.Destination)
	if nextHop.HighestID != 0 {
		fmt.Printf("Sending data reply %s to %s via %s\n", hex.EncodeToString(dataReplyMessage.HashValue), dataReplyMessage.Destination, nextHop.Address)
		g.sendToPeer(gossipMessage, nextHop.Address)
	} else {
		fmt.Printf("No next hop to %s. Sending data response for %s back to where it came from: %s\n", dataReplyMessage.Destination, hex.EncodeToString(dataReplyMessage.HashValue), sender)
		g.sendToPeer(gossipMessage, sender)
	}
}

func (g *Gossiper) indexFile(msg utils.Message) {
	fmt.Printf("REQUESTING INDEXING filename %s\n", msg.FileName)
	var hashFunc = crypto.SHA256.New()
	var chunkHashed []byte
	file, e := os.Open(fmt.Sprintf(".%s%s%s%s", string(os.PathSeparator), utils.SHARED_FOLDER, string(os.PathSeparator), msg.FileName))
	utils.HandleError(e)

	fileInfo, e := file.Stat()
	utils.HandleError(e)

	fileSize := fileInfo.Size()
	noChunks := int(math.Ceil(float64(fileSize) / float64(utils.CHUNK_SIZE)))

	for i := 0; i < noChunks; i++ {
		hashFunc.Reset()
		chunk := utils.GetNextDataChunk(file, utils.CHUNK_SIZE)
		hashFunc.Write(chunk)
		temp := hashFunc.Sum(nil)
		g.addStoredChunk(hex.EncodeToString(temp), chunk)
		chunkHashed = append(chunkHashed, temp...)
	}
	file.Close()

	if len(chunkHashed) > utils.CHUNK_SIZE {
		fmt.Printf("CAN'T INDEX FILE %s: METAFILE LARGER THAN ALLOWED\n", msg.FileName)
		return
	}

	hashFunc.Reset()
	hashFunc.Write(chunkHashed)
	metaHash := hex.EncodeToString(hashFunc.Sum(nil))
	g.addStoredChunk(metaHash, chunkHashed)
	fmt.Printf("Indexed File with Metahash %s\n", metaHash)
	g.setStoredFile(utils.StringHash(hashFunc.Sum(nil)), utils.File{Name: msg.FileName, MetafileHash: utils.ByteMetaHash(metaHash), Size: fileSize})
	g.txPublishHandler(utils.TxPublish{File: utils.File{Name: msg.FileName, MetafileHash: utils.ByteMetaHash(metaHash), Size: fileSize}, HopLimit: utils.TX_PUBLISH_HOP_LIMIT + 1}, "")
}
