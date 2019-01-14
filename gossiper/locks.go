package gossiper

import (
	"errors"
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

func (g *Gossiper) setChallengeChannel(key string, value chan utils.Challenge) {
	g.challengeChannelLock.Lock()
	g.challengeChannel[key] = value
	g.challengeChannelLock.Unlock()
}

func (g *Gossiper) getChallengeChannel(key string) chan utils.Challenge {
	g.challengeChannelLock.RLock()
	val := g.challengeChannel[key]
	g.challengeChannelLock.RUnlock()
	return val
}

func (g *Gossiper) sendToChallangeChannel(key string, value utils.Challenge) {
	g.challengeChannelLock.Lock()
	g.challengeChannel[key] <- value
	g.challengeChannelLock.Unlock()
}

func (g *Gossiper) deleteChallengeChannel(key string) {
	g.challengeChannelLock.Lock()
	delete(g.challengeChannel, key)
	g.challengeChannelLock.Unlock()
}

func (g *Gossiper) setFileExchangeChannel(key string, value chan utils.FileExchangeRequest) {
	g.fileExchangeChannelLock.Lock()
	g.fileExchangeChannel[key] = value
	g.fileExchangeChannelLock.Unlock()
}

func (g *Gossiper) getFileExchangeChannel(key string) chan utils.FileExchangeRequest {
	g.fileExchangeChannelLock.RLock()
	val := g.fileExchangeChannel[key]
	g.fileExchangeChannelLock.RUnlock()
	return val
}

func (g *Gossiper) sendToFileExchangeChannel(key string, value utils.FileExchangeRequest) {
	g.fileExchangeChannelLock.Lock()
	g.fileExchangeChannel[key] <- value
	g.fileExchangeChannelLock.Unlock()
}

func (g *Gossiper) deleteFileExchangeChannel(key string) {
	g.fileExchangeChannelLock.Lock()
	delete(g.fileExchangeChannel, key)
	g.fileExchangeChannelLock.Unlock()
}

func (g *Gossiper) addPrivateFile(msg utils.PrivateFile) {
	g.privFileLock.Lock()
	g.PrivFiles[utils.StringHash(msg.MetafileHash)] = msg
	for i, _ := range msg.Replications {
		el := &msg.Replications[i]
		g.Replications[el.Metafilehash] = el
	}
	g.privFileLock.Unlock()
}

func (g *Gossiper) GetReplica(mfh string) utils.Replica {
	g.privFileLock.RLock()
	val := *g.Replications[mfh]
	g.privFileLock.RUnlock()
	return val
}

// If nodeID is empty, then no exchange has been arranged so far
func (g *Gossiper) scanOpenFileExchanges(peer string) utils.Replica {
	g.privFileLock.Lock()
	defer g.privFileLock.Unlock()
Loop:
	for k, f := range g.PrivFiles {
		var candidate *utils.Replica = nil
		for i, _ := range f.Replications {
			// Node is already serving a replica of this file --> Not interesting
			if g.PrivFiles[k].Replications[i].NodeID == peer {
				continue Loop
			}
			if g.PrivFiles[k].Replications[i].NodeID == "" {
				candidate = &g.PrivFiles[k].Replications[i]
			}
		}
		if candidate != nil {
			return *candidate
		}
	}
	return utils.Replica{}
}

// Returns true if the peer is already storing a replication of that file
func (g *Gossiper) checkReplicationAtPeer(mfh string, peer string) bool {
	g.privFileLock.RLock()
	defer g.privFileLock.RUnlock()
	for _, f := range g.PrivFiles {
		for _, r := range f.Replications {
			if r.Metafilehash == mfh {
				return g.holdsFile(f, peer)
			}
		}
	}
	return false
}

func (g *Gossiper) holdsFile(file utils.PrivateFile, peer string) bool {
	for _, r := range file.Replications {
		if r.NodeID == peer {
			return true
		}
	}
	return false
}

// Assigning the Replica to a peer
// Throws an error if the Replica is already assigned to another peer
func (g *Gossiper) AssignReplica(mfh string, nodeID string, exchangeMFH string) error {
	g.privFileLock.Lock()
	defer g.privFileLock.Unlock()
	if g.Replications[mfh] == nil {
		return errors.New("Replica is not existing")
	}
	if g.Replications[mfh].NodeID != "" || g.Replications[mfh].ExchangeMFH != "" {
		return errors.New("Replication is already assigned")
	}
	fmt.Println("Assigning new parameters for: ", mfh, nodeID, exchangeMFH)
	g.Replications[mfh].NodeID = nodeID
	g.Replications[mfh].ExchangeMFH = exchangeMFH
	return nil
}

// Removing references to peer, thereby allowing the Replica to be reassigned
func (g *Gossiper) freeReplica(mfh string) {
	g.privFileLock.Lock()
	g.Replications[mfh].NodeID = ""
	g.Replications[mfh].ExchangeMFH = ""
	g.privFileLock.Unlock()
}

func (g *Gossiper) getDistributedReplica(f utils.PrivateFile) utils.Replica {
	g.privFileLock.RLock()
	defer g.privFileLock.RUnlock()
	for _, r := range f.Replications {
		if r.NodeID != "" {
			return r
		}
	}
	return utils.Replica{}
}

func (g *Gossiper) countDistributedReplicas(f utils.PrivateFile) int {
	g.privFileLock.RLock()
	defer g.privFileLock.RUnlock()
	count := 0
	for _, r := range f.Replications {
		if r.NodeID != "" {
			count++
		}
	}
	return count
}

func (g *Gossiper) getPrivateFile(mfh string) utils.PrivateFile {
	g.privFileLock.RLock()
	defer g.privFileLock.RUnlock()
	return g.PrivFiles[mfh]
}

func (g *Gossiper) getDataRequestChannel(key string) chan bool {
	g.dataRequestChannelLock.RLock()
	val := g.dataRequestChannel[key]
	g.dataRequestChannelLock.RUnlock()
	return val
}

func (g *Gossiper) getDestinationSpecified(key string) bool {
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
	if key == g.Name {
		val = utils.HopInfo{Address: g.Address.String()}
	} else {
		val = g.NextHop[key]
	}
	g.nextHopLock.RUnlock()
	return val
}

func (g *Gossiper) setNextHop(key string, value utils.HopInfo) {
	g.nextHopLock.Lock()
	g.NextHop[key] = value
	g.nextHopLock.Unlock()
}

func (g *Gossiper) getStoredFile(key string) utils.File {
	g.storedFilesLock.RLock()
	val := g.StoredFiles[key]
	g.storedFilesLock.RUnlock()
	return val
}

func (g *Gossiper) setStoredFile(key string, value utils.File) {
	g.storedFilesLock.Lock()
	g.StoredFiles[key] = value
	g.storedFilesLock.Unlock()
}

func (g *Gossiper) getStoredChunk(key string) []byte {
	g.storedChunksLock.RLock()
	val := g.StoredChunks[key]
	g.storedChunksLock.RUnlock()
	return val
}

func (g *Gossiper) removeStoredChunk(key string) {
	g.storedChunksLock.Lock()
	delete(g.StoredChunks, key)
	g.storedChunksLock.Unlock()
}

func (g *Gossiper) addStoredChunk(key string, value []byte) {
	g.storedChunksLock.Lock()
	g.StoredChunks[key] = value
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
	// Filter route rumor messages
	if msg.Text != "" {
		g.ChronRumorMessages = append(g.ChronRumorMessages, utils.StoredMessage{Message: msg, Timestamp: time.Now()})
	}
	g.chronRumorMessagesLock.Unlock()
}

func (g *Gossiper) getAllRumorMessages() string {
	allMsg := ""
	g.chronRumorMessagesLock.RLock()
	for _, stored := range g.ChronRumorMessages {
		msg, _ := stored.Message.(*utils.RumorMessage)
		if msg.Origin == g.Name {
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
	if msg.Origin != g.Name {
		g.ChronPrivateMessages[msg.Origin] = append(g.ChronPrivateMessages[msg.Origin], utils.StoredMessage{Message: msg, Timestamp: time.Now()})
	} else {
		g.ChronPrivateMessages[msg.Destination] = append(g.ChronPrivateMessages[msg.Destination], utils.StoredMessage{Message: msg, Timestamp: time.Now()})
	}
	g.chronPrivateMessagesLock.Unlock()
}

func (g *Gossiper) getAllPrivateMessages(dest string) string {
	allMsg := ""
	g.chronPrivateMessagesLock.RLock()
	for _, stored := range g.ChronPrivateMessages[dest] {
		msg, _ := stored.Message.(*utils.PrivateMessage)

		if msg.Destination == dest {
			if msg.Text != "" {
				allMsg = fmt.Sprintf("%sYOU: %s\n", allMsg, msg.Text)
			}
			if msg.EncryptedText != "" {
				allMsg = fmt.Sprintf("%sYOU (encrypted): %s\n", allMsg, msg.EncryptedText)
			}
		} else {
			if msg.Text != "" {
				allMsg = fmt.Sprintf("%s%s: %s\n", allMsg, msg.Origin, msg.Text)
			}
			if msg.EncryptedText != "" {
				allMsg = fmt.Sprintf("%s%s (encrypted): %s\n", allMsg, msg.Origin, msg.EncryptedText)
			}
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

	if cached.Timestamp.IsZero() || cached.Timestamp.Add(utils.CACHING_DURATION).Before(time.Now()) {
		g.putSearchRequest(msg)
		return false
	}
	return true
}

func (g *Gossiper) GetAllOrigins() string {
	g.nextHopLock.RLock()
	allOrigins := ""
	for k, _ := range g.NextHop {
		allOrigins = allOrigins + k + "\n"
	}
	g.nextHopLock.RUnlock()
	return allOrigins
}

func (g *Gossiper) getAllPrivateFiles() map[string]string {
	g.privFileLock.RLock()
	// metafilehash to string
	allPrivateFiles := make(map[string]string)

	for k, v := range g.PrivFiles {
		allPrivateFiles[k] = v.Name
	}
	g.privFileLock.RUnlock()
	return allPrivateFiles
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
	if !reflect.ValueOf(g.ExternalFiles[key]).IsNil() && len(g.ExternalFiles[key].Holder[chunkNr]) > 0 {
		val = g.ExternalFiles[key].Holder[chunkNr][0]
	}
	g.externalFilesLock.RUnlock()
	return val
}

func (g *Gossiper) getChronReceivedFiles() []*utils.ExternalFile {
	g.chronReceivedFilesLock.RLock()
	val := g.ChronReceivedFiles
	g.chronReceivedFilesLock.RUnlock()
	return val
}

func (g *Gossiper) addChronReceivedFiles(value *utils.ExternalFile) {
	g.chronReceivedFilesLock.Lock()

	for _, v := range g.ChronReceivedFiles {
		if utils.StringHash(v.MetafileHash) == utils.StringHash(value.MetafileHash) {
			// Already have this file, nothing to append
			g.chronReceivedFilesLock.Unlock()
			return
		}
	}

	if reflect.ValueOf(g.ChronReceivedFiles).IsNil() {
		g.ChronReceivedFiles = []*utils.ExternalFile{value}
	} else {
		g.ChronReceivedFiles = append(g.ChronReceivedFiles, value)
	}
	g.chronReceivedFilesLock.Unlock()
}

func (g *Gossiper) inBlockHistory(key utils.TxPublish) bool {
	g.chainLock.Lock()
	defer g.chainLock.Unlock()
	currBlock := g.LastBlock
	for currBlock.Counter != 0 {
		for _, v := range currBlock.Transactions {
			if v.File.Name == key.File.Name {
				return true
			}
		}
		currBlock = g.BlockHistory[currBlock.PrevHash] // next block
	}
	return false
}

func (g *Gossiper) inPendingTransactions(key utils.TxPublish) bool {
	g.chainLock.Lock()
	defer g.chainLock.Unlock()
	for _, v := range g.PendingTransactions {
		if v.File.Name == key.File.Name {
			return true
		}
	}
	return false
}

func (g *Gossiper) addPendingTransaction(key utils.TxPublish) {
	g.chainLock.Lock()
	g.PendingTransactions = append(g.PendingTransactions, key)
	// notify miner
	g.miner <- true
	g.chainLock.Unlock()
}

func (g *Gossiper) getPendingTransactions() []utils.TxPublish {
	g.chainLock.RLock()
	val := g.PendingTransactions
	g.chainLock.RUnlock()
	return val
}

func (g *Gossiper) getLastBlock() utils.BlockWrapper {
	g.chainLock.RLock()
	val := g.LastBlock
	g.chainLock.RUnlock()
	return val
}
