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

func (g *Gossiper) setDataRequestChannel(key string, value chan bool) {
	g.dataRequestChannelLock.Lock()
	g.dataRequestChannel[key] = value
	g.dataRequestChannelLock.Unlock()
}

func (g *Gossiper) deleteDataRequestChannel(key string) {
	g.dataRequestChannelLock.Lock()
	delete(g.dataRequestChannel, key)
	g.dataRequestChannelLock.Unlock()
}

func (g *Gossiper) sendToDataRequestChannel(key string, value bool) {
	g.dataRequestChannelLock.Lock()
	g.dataRequestChannel[key] <- value
	g.dataRequestChannelLock.Unlock()
}

func (g *Gossiper) getNextHop(key string) utils.HopInfo {
	g.nextHopLock.RLock()
	val := g.nextHop[key]
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
	g.externalFilesLock.RLock()
	val := g.externalFiles[key].Holder[chunkNr][0]
	g.externalFilesLock.RUnlock()
	return val
}


func (g *Gossiper) getChronReceivedFiles(msg utils.SearchRequest) []*utils.ExternalFile {
	g.chronReceivedFilesLock.RLock()
	val := g.chronReceivedFiles[msg.GetLocalIdentifier()]
	g.chronReceivedFilesLock.RUnlock()
	return val
}

func (g *Gossiper) addChronReceivedFiles(key utils.SearchRequest, value *utils.ExternalFile) {
	g.chronReceivedFilesLock.Lock()
	if reflect.ValueOf(g.chronReceivedFiles[key.GetLocalIdentifier()]).IsNil() {
		g.chronReceivedFiles[key.GetLocalIdentifier()] = []*utils.ExternalFile{value}
	} else {
		g.chronReceivedFiles[key.GetLocalIdentifier()] = append(g.chronReceivedFiles[key.GetLocalIdentifier()], value)
	}
	g.chronReceivedFilesLock.Unlock()
}
