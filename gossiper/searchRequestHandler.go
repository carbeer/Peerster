package gossiper

import (
	"fmt"
	"math/rand"
	"reflect"
	"strings"
	"time"

	"github.com/carbeer/Peerster/utils"
)

// Creates SearchRequest as per client request
func (g *Gossiper) newSearchRequest(msg utils.Message) {
	var searchRequest utils.SearchRequest
	var timeout <-chan time.Time
	fulfilled := uint32(0)
	defaultBudget := false

	if msg.Budget < 0 {
		defaultBudget = true
		searchRequest = utils.SearchRequest{Origin: g.Name, Budget: utils.DEFAULT_BUDGET, Keywords: msg.Keywords}
	} else {
		searchRequest = utils.SearchRequest{Origin: g.Name, Budget: uint64(msg.Budget), Keywords: msg.Keywords}
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
			// Update number of successful search requests
			fulfilled += new
			if fulfilled >= utils.MIN_THRESHOLD {
				fmt.Printf("SEARCH FINISHED\n")
				break Loop
			}
		}
	}
	g.deleteSearchRequestChannel(searchRequest)
}

// Handles incoming SearchRequests
// Resend is used to indicate that the current process already searched for available files
func (g *Gossiper) searchRequestHandler(msg utils.SearchRequest, resend bool) {

	if g.isCachedSearchRequest(msg) {
		return
	}

	msg.Budget--

	if !resend {
		// Get local matches
		g.searchReplyHandler(g.searchForOwnedFiles(msg))
	}

	// Initiate Reply
	if msg.Budget <= 0 {
		return
	}

	perPeerBudget := int(msg.Budget) / len(g.Peers)
	remainder := int(msg.Budget) % len(g.Peers)

	gossipMessageLargeB := utils.GossipPacket{SearchRequest: &utils.SearchRequest{Origin: msg.Origin, Budget: uint64(perPeerBudget + remainder), Keywords: msg.Keywords}}
	gossipMessageSmallB := utils.GossipPacket{SearchRequest: &utils.SearchRequest{Origin: msg.Origin, Budget: uint64(perPeerBudget), Keywords: msg.Keywords}}

	// Randomize which peers will get the larger budget
	for ix := range rand.Perm(len(g.Peers)) {
		if remainder > 0 {
			remainder--
			g.sendToPeer(gossipMessageLargeB, g.Peers[ix])
		} else if perPeerBudget > 0 {
			g.sendToPeer(gossipMessageSmallB, g.Peers[ix])
		} else {
			break
		}
	}
}

// Retrieve matching local files as SearchReply
func (g *Gossiper) searchForOwnedFiles(msg utils.SearchRequest) utils.SearchReply {
	results := []*utils.SearchResult{}

	g.storedFilesLock.RLock()
	defer g.storedFilesLock.RUnlock()
	for _, v := range g.StoredFiles {
		for _, name := range msg.Keywords {
			if strings.Contains(v.Name, name) {
				chunkMap, chunkCount := g.getAvailableChunks(v)
				results = append(results, &utils.SearchResult{FileName: v.Name, MetafileHash: v.MetafileHash, ChunkMap: chunkMap, ChunkCount: chunkCount})
				fmt.Println("Found the following files: ")
				fmt.Printf("%+v", &utils.SearchResult{FileName: v.Name, MetafileHash: v.MetafileHash, ChunkMap: chunkMap, ChunkCount: chunkCount})
				break // Already returning the file, no matter how many additional keyword matches we have
			}
		}
	}
	return utils.SearchReply{Origin: g.Name, Destination: msg.Origin, HopLimit: utils.HOPLIMIT_CONSTANT, Results: results}
}

// Returns all available chunks for a specific File
func (g *Gossiper) getAvailableChunks(file utils.File) ([]uint64, uint64) {
	var chunkMap []uint64
	metaFile := g.getStoredChunk(utils.StringHash(file.MetafileHash))
	i := 0
	for ; utils.GetHashAtIndex(metaFile, i) != nil; i++ {
		if !reflect.ValueOf(g.getStoredChunk(utils.StringHash(utils.GetHashAtIndex(metaFile, i)))).IsNil() {
			chunkMap = append(chunkMap, uint64(i+1))
		}
	}
	return chunkMap, uint64(i)
}

// Get files that were found for the specified keywords
func (g *Gossiper) getAvailableFileResults(keywords []string) []utils.FileSkeleton {
	results := []utils.FileSkeleton{}
	g.chronReceivedFilesLock.Lock()

	if keywords == nil {
		for _, v := range g.ChronReceivedFiles {
			results = append(results, utils.FileSkeleton{Name: v.File.Name, MetafileHash: utils.StringHash(v.File.MetafileHash)})
		}
	} else {
		for _, v := range g.ChronReceivedFiles {
			if utils.Contains(keywords, v.Name) {
				results = append(results, utils.FileSkeleton{Name: v.File.Name, MetafileHash: utils.StringHash(v.File.MetafileHash)})
			}
		}
	}

	g.chronReceivedFilesLock.Unlock()
	return results
}
