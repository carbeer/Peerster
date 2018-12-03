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

func (g *Gossiper) newSearchRequest(msg utils.Message) {
	var searchRequest utils.SearchRequest
	var timeout <-chan time.Time
	fulfilled := uint32(0)
	defaultBudget := false

	if msg.Budget < 0 {
		defaultBudget = true
		searchRequest = utils.SearchRequest{Origin: g.name, Budget: utils.GetDefaultBudget(), Keywords: msg.Keywords}
	} else {
		searchRequest = utils.SearchRequest{Origin: g.name, Budget: uint64(msg.Budget), Keywords: msg.Keywords}
	}
	// Create a channel that is added to the list of searchRequests
	g.setSearchRequestChannel(searchRequest, make(chan uint32, utils.GetMsgBuffer()))
	g.searchRequestHandler(searchRequest, false)
	fmt.Printf("SEARCHING for keywords %s with budget %d\n", strings.Join(searchRequest.Keywords, ","), searchRequest.Budget)

	if defaultBudget {
		timeout = time.After(utils.GetSearchRequestTimeout())
	}
Loop:
	for {
		select {
		case <-timeout:
			fmt.Printf("%d: TIMEOUT, DOUBLING BUDGET\n", time.Now().Second())
			if searchRequest.Budget == utils.GetMaximumBudget() {
				fmt.Printf("Reached maximum budget. Stop searching.\n")
				break Loop
			}
			searchRequest.Budget = utils.MinUint64(searchRequest.Budget*2, utils.GetMaximumBudget())
			fmt.Printf("SEARCHING for keywords %s with budget %d\n", strings.Join(searchRequest.Keywords, ","), searchRequest.Budget)
			g.searchRequestHandler(searchRequest, true)
			timeout = time.After(utils.GetSearchRequestTimeout())
		case new := <-g.getSearchRequestChannel(searchRequest):
			fulfilled += new
			// Process message
			if fulfilled >= utils.GetMinimumThreshold() {
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
	} else if msg.Budget >= utils.GetMaximumBudget() {
		return true
	}
	return false
}

func (g *Gossiper) newDataReplyMessage(msg utils.DataRequest, sender string) {
	dataReplyMessage := utils.DataReply{Origin: g.name, Destination: msg.Origin, HopLimit: utils.GetHopLimitConstant(), HashValue: msg.HashValue, Data: g.getStoredChunk(hex.EncodeToString(msg.HashValue))}
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
	file, e := os.Open(fmt.Sprintf(".%s%s%s%s", string(os.PathSeparator), utils.GetSharedFolder(), string(os.PathSeparator), msg.FileName))
	utils.HandleError(e)

	fileInfo, e := file.Stat()
	utils.HandleError(e)

	fileSize := fileInfo.Size()
	noChunks := int(math.Ceil(float64(fileSize) / float64(utils.GetChunkSize())))

	for i := 0; i < noChunks; i++ {
		hashFunc.Reset()
		chunk := utils.GetNextDataChunk(file, utils.GetChunkSize())
		hashFunc.Write(chunk)
		temp := hashFunc.Sum(nil)
		g.addStoredChunk(hex.EncodeToString(temp), chunk)
		chunkHashed = append(chunkHashed, temp...)
	}
	file.Close()

	if len(chunkHashed) > utils.GetChunkSize() {
		fmt.Printf("CAN'T INDEX FILE %s: METAFILE LARGER THAN ALLOWED\n", msg.FileName)
		return
	}

	hashFunc.Reset()
	hashFunc.Write(chunkHashed)
	metaHash := hex.EncodeToString(hashFunc.Sum(nil))
	g.addStoredChunk(metaHash, chunkHashed)
	fmt.Printf("Indexed File with Metahash %s\n", metaHash)
	g.setStoredFile(hex.EncodeToString(hashFunc.Sum(nil)), utils.File{FileName: msg.FileName, MetaHash: metaHash, FileSize: fileSize})
}