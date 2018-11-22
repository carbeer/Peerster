package gossiper

import "github.com/carbeer/Peerster/utils"

func (g *Gossiper) getReceivedMessages(key string) []utils.RumorMessage {
	g.receivedMessagesLock.RLock()
	val := g.ReceivedMessages[key]
	g.receivedMessagesLock.RUnlock()
	return val
}

func (g *Gossiper) appendReceivedMessages(key string, value utils.RumorMessage) {
	g.receivedMessagesLock.Lock()
	g.ReceivedMessages[key] = append(g.ReceivedMessages[key], value)
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
