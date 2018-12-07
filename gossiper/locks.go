package gossiper

import (
	"fmt"
	"reflect"
	"time"

	"github.com/carbeer/Peerster/utils"
)

func (g *Gossiper) getReceivedMessages(key string) []utils.RumorMessage {
	g.receivedMessagesLock.RLock()
	val := g.ReceivedMessages[key]
	g.receivedMessagesLock.RUnlock()
	return val
}

func (g *Gossiper) appendReceivedMessages(key string, value utils.RumorMessage) {
	g.receivedMessagesLock.Lock()
	g.ReceivedMessages[key] = append(g.ReceivedMessages[key], value)
	g.addChronRumorMessage(&value)
	g.receivedMessagesLock.Unlock()
}

func (g *Gossiper) getPrivateMessages(key string) []utils.PrivateMessage {
	g.privateMessagesLock.RLock()
	val := g.PrivateMessages[key]
	g.privateMessagesLock.RUnlock()
	return val
}

func (g *Gossiper) appendPrivateMessages(key string, value utils.PrivateMessage) {
	g.privateMessagesLock.Lock()
	g.PrivateMessages[key] = append(g.PrivateMessages[key], value)
	g.addChronPrivateMessage(&value)
	g.privateMessagesLock.Unlock()
}

func (g *Gossiper) getRumorMongeringChannel(key string) chan utils.StatusPacket {
	g.rumorMongeringChannelLock.RLock()
	val := g.rumorMongeringChannel[key]
	g.rumorMongeringChannelLock.RUnlock()
	return val
}

func (g *Gossiper) setRumorMongeringChannel(key string, value chan utils.StatusPacket) {
	g.rumorMongeringChannelLock.Lock()
	g.rumorMongeringChannel[key] = value
	g.rumorMongeringChannelLock.Unlock()
}

func (g *Gossiper) deleteRumorMongeringChannel(key string) {
	g.rumorMongeringChannelLock.Lock()
	delete(g.rumorMongeringChannel, key)
	g.rumorMongeringChannelLock.Unlock()
}

func (g *Gossiper) sendToRumorMongeringChannel(key string, value utils.StatusPacket) {
	g.rumorMongeringChannelLock.Lock()
	g.rumorMongeringChannel[key] <- value
	g.rumorMongeringChannelLock.Unlock()
}

func (g *Gossiper) getDataRequestChannel(key string) chan bool {
	g.dataRequestChannelLock.RLock()
	val := g.dataRequestChannel[key]
	g.dataRequestChannelLock.RUnlock()
	return val
}

func (g *Gossiper) getDesintationSpecified(key string) bool {
	g.dataRequestChannelLock.RLock()
	val := g.destinationSpecified[key]
	g.dataRequestChannelLock.RUnlock()
	return val
}

func (g *Gossiper) setDataRequestChannel(key string, value chan bool, addValue bool) {
	g.dataRequestChannelLock.Lock()
	g.dataRequestChannel[key] = value
	g.destinationSpecified[key] = addValue
	g.dataRequestChannelLock.Unlock()
}

func (g *Gossiper) deleteDataRequestChannel(key string) {
	g.dataRequestChannelLock.Lock()
	delete(g.dataRequestChannel, key)
	delete(g.destinationSpecified, key)
	g.dataRequestChannelLock.Unlock()
}

func (g *Gossiper) sendToDataRequestChannel(key string, value bool) {
	g.dataRequestChannelLock.Lock()
	g.dataRequestChannel[key] <- value
	g.dataRequestChannelLock.Unlock()
}

func (g *Gossiper) getNextHop(key string) utils.HopInfo {
	var val utils.HopInfo
	g.nextHopLock.RLock()
	if key == g.name {
		val = utils.HopInfo{Address: g.Address.String()}
	} else {
		val = g.nextHop[key]
	}
	g.nextHopLock.RUnlock()
	return val
}

func (g *Gossiper) setNextHop(key string, value utils.HopInfo) {
	g.nextHopLock.Lock()
	g.nextHop[key] = value
	g.nextHopLock.Unlock()
}

func (g *Gossiper) getStoredFile(key string) utils.File {
	g.storedFilesLock.RLock()
	val := g.storedFiles[key]
	g.storedFilesLock.RUnlock()
	return val
}

func (g *Gossiper) setStoredFile(key string, value utils.File) {
	g.storedFilesLock.Lock()
	g.storedFiles[key] = value
	g.storedFilesLock.Unlock()
}

func (g *Gossiper) getStoredChunk(key string) []byte {
	g.storedChunksLock.RLock()
	val := g.storedChunks[key]
	g.storedChunksLock.RUnlock()
	return val
}

func (g *Gossiper) addStoredChunk(key string, value []byte) {
	g.storedChunksLock.Lock()
	g.storedChunks[key] = value
	g.storedChunksLock.Unlock()
}

func (g *Gossiper) getRequestedChunks(key string) utils.ChunkInfo {
	g.requestedChunksLock.RLock()
	val := g.requestedChunks[key]
	g.requestedChunksLock.RUnlock()
	return val
}

func (g *Gossiper) addRequestedChunks(key string, value utils.ChunkInfo) {
	g.requestedChunksLock.Lock()
	g.requestedChunks[key] = value
	g.requestedChunksLock.Unlock()
}

func (g *Gossiper) popRequestedChunks(key string) utils.ChunkInfo {
	g.requestedChunksLock.Lock()
	val := g.requestedChunks[key]
	delete(g.requestedChunks, key)
	g.requestedChunksLock.Unlock()
	return val
}

func (g *Gossiper) addChronRumorMessage(value interface{}) {
	g.chronRumorMessagesLock.Lock()
	msg, _ := value.(*utils.RumorMessage)
	g.chronRumorMessages = append(g.chronRumorMessages, utils.StoredMessage{Message: msg, Timestamp: time.Now()})
	g.chronRumorMessagesLock.Unlock()
}

func (g *Gossiper) getAllRumorMessages() string {
	allMsg := ""
	g.chronRumorMessagesLock.RLock()
	for _, stored := range g.chronRumorMessages {
		msg, _ := stored.Message.(*utils.RumorMessage)
		if msg.Origin == g.name {
			allMsg = fmt.Sprintf("%sYOU: %s\n", allMsg, msg.Text)
		} else {
			allMsg = fmt.Sprintf("%s%s: %s\n", allMsg, msg.Origin, msg.Text)
		}
	}
	g.chronRumorMessagesLock.RUnlock()
	return allMsg
}

func (g *Gossiper) addChronPrivateMessage(value interface{}) {
	g.chronPrivateMessagesLock.Lock()
	msg, _ := value.(*utils.PrivateMessage)
	if msg.Origin != g.name {
		g.chronPrivateMessages[msg.Origin] = append(g.chronPrivateMessages[msg.Origin], utils.StoredMessage{Message: msg, Timestamp: time.Now()})
	} else {
		g.chronPrivateMessages[msg.Origin] = append(g.chronPrivateMessages[msg.Destination], utils.StoredMessage{Message: msg, Timestamp: time.Now()})
	}
	g.chronPrivateMessagesLock.Unlock()
}

func (g *Gossiper) getAllPrivateMessages(dest string) string {
	allMsg := ""
	g.chronPrivateMessagesLock.RLock()
	for _, stored := range g.chronPrivateMessages[dest] {
		msg, _ := stored.Message.(*utils.PrivateMessage)
		if msg.Destination == dest {
			allMsg = fmt.Sprintf("%sYOU: %s\n", allMsg, msg.Text)
		} else {
			allMsg = fmt.Sprintf("%s%s: %s\n", allMsg, msg.Origin, msg.Text)
		}
	}
	g.chronPrivateMessagesLock.RUnlock()
	return allMsg
}

// Cached requests: same keywords and origin within 0.5s
func (g *Gossiper) isCachedSearchRequest(msg utils.SearchRequest) bool {
	g.cachedSearchRequestsLock.RLock()
	cached := g.CachedSearchRequests[msg.GetIdentifier()]
	g.cachedSearchRequestsLock.RUnlock()

	if cached.Timestamp.IsZero() || cached.Timestamp.Add(utils.GetCachingDurationMS()).Before(time.Now()) {
		g.putSearchRequest(msg)
		return false
	}
	return true
}

// creates or updates searchRequest
func (g *Gossiper) putSearchRequest(msg utils.SearchRequest) {

	g.cachedSearchRequestsLock.Lock()
	if g.CachedSearchRequests[msg.GetIdentifier()].Timestamp.IsZero() {
		g.CachedSearchRequests[msg.GetIdentifier()] = utils.CachedRequest{Timestamp: time.Now(), Request: msg}
	} else {
		tmp := g.CachedSearchRequests[msg.GetIdentifier()]
		tmp.Timestamp = time.Now()
		g.CachedSearchRequests[msg.GetIdentifier()] = tmp
	}
	g.cachedSearchRequestsLock.Unlock()
}

func (g *Gossiper) setSearchRequestChannel(key utils.SearchRequest, value chan uint32) {
	g.searchRequestChannelLock.Lock()
	g.searchRequestChannel[key.GetKeywordIdentifier()] = value
	g.searchRequestChannelLock.Unlock()
}

func (g *Gossiper) getSearchRequestChannel(key utils.SearchRequest) chan uint32 {
	g.searchRequestChannelLock.RLock()
	val := g.searchRequestChannel[key.GetKeywordIdentifier()]
	g.searchRequestChannelLock.RUnlock()
	return val
}

func (g *Gossiper) deleteSearchRequestChannel(key utils.SearchRequest) {
	g.searchRequestChannelLock.Lock()
	delete(g.searchRequestChannel, key.GetKeywordIdentifier())
	g.searchRequestChannelLock.Unlock()
}

func (g *Gossiper) sendToSearchRequestChannel(key utils.SearchRequest, value uint32) {
	g.searchRequestChannelLock.Lock()
	g.searchRequestChannel[key.GetKeywordIdentifier()] <- value
	g.searchRequestChannelLock.Unlock()
}

func (g *Gossiper) getChunkHolder(key string, chunkNr int) string {
	var val string
	g.externalFilesLock.RLock()
	if !reflect.ValueOf(g.externalFiles[key]).IsNil() && len(g.externalFiles[key].Holder[chunkNr]) > 0 {
		val = g.externalFiles[key].Holder[chunkNr][0]
	}
	g.externalFilesLock.RUnlock()
	return val
}

func (g *Gossiper) getChronReceivedFiles() []*utils.ExternalFile {
	g.chronReceivedFilesLock.RLock()
	val := g.chronReceivedFiles
	g.chronReceivedFilesLock.RUnlock()
	return val
}

func (g *Gossiper) addChronReceivedFiles(value *utils.ExternalFile) {
	g.chronReceivedFilesLock.Lock()

	for _, v := range g.chronReceivedFiles {
		if utils.StringHash(v.MetafileHash) == utils.StringHash(value.MetafileHash) {
			// Already have this file, nothing to append
			g.chronReceivedFilesLock.Unlock()
			return
		}
	}

	if reflect.ValueOf(g.chronReceivedFiles).IsNil() {
		g.chronReceivedFiles = []*utils.ExternalFile{value}
	} else {
		g.chronReceivedFiles = append(g.chronReceivedFiles, value)
	}
	g.chronReceivedFilesLock.Unlock()
}

func (g *Gossiper) inBlockHistory(key utils.TxPublish) bool {
	g.chainLock.Lock()
	defer g.chainLock.Unlock()
	currBlock := g.lastBlock
	for currBlock.Counter != 0 {
		for _, v := range currBlock.Transactions {
			if v.File.Name == key.File.Name {
				return true
			}
		}
		currBlock = g.blockHistory[currBlock.PrevHash] // next block
	}
	return false
}

func (g *Gossiper) inPendingTransactions(key utils.TxPublish) bool {
	g.chainLock.Lock()
	defer g.chainLock.Unlock()
	for _, v := range g.pendingTransactions {
		if v.File.Name == key.File.Name {
			return true
		}
	}
	return false
}

func (g *Gossiper) addPendingTransaction(key utils.TxPublish) {
	g.chainLock.Lock()
	g.pendingTransactions = append(g.pendingTransactions, key)
	// notify miner
	g.miner <- true
	g.chainLock.Unlock()
}

func (g *Gossiper) getPendingTransactions() []utils.TxPublish {
	g.chainLock.RLock()
	val := g.pendingTransactions
	g.chainLock.RUnlock()
	return val
}

func (g *Gossiper) getLastBlock() utils.BlockWrapper {
	g.chainLock.RLock()
	val := g.lastBlock
	g.chainLock.RUnlock()
	return val
}
