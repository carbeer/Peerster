package gossiper

import (
	"fmt"
	"math/rand"
	"reflect"
	"strings"

	"github.com/carbeer/Peerster/utils"
)

// resend is used to indicate that the current process already searched for available files
func (g *Gossiper) searchRequestHandler(msg utils.SearchRequest, resend bool) {

	if g.isCachedSearchRequest(msg) {
		return
	}

	msg.Budget--

	if !resend {
		g.searchReplyHandler(g.searchForOwnedFiles(msg))
	}

	// Initiate Reply
	if msg.Budget <= 0 {
		return
	}

	perPeerBudget := int(msg.Budget) / len(g.peers)
	remainder := int(msg.Budget) % len(g.peers)

	gossipMessageLargeB := utils.GossipPacket{SearchRequest: &utils.SearchRequest{Origin: msg.Origin, Budget: uint64(perPeerBudget + remainder), Keywords: msg.Keywords}}
	gossipMessageSmallB := utils.GossipPacket{SearchRequest: &utils.SearchRequest{Origin: msg.Origin, Budget: uint64(perPeerBudget), Keywords: msg.Keywords}}

	// Randomize which peers will get the larger budget
	for ix := range rand.Perm(len(g.peers)) {
		if remainder > 0 {
			remainder--
			g.sendToPeer(gossipMessageLargeB, g.peers[ix])
		} else if perPeerBudget > 0 {
			g.sendToPeer(gossipMessageSmallB, g.peers[ix])
		} else {
			break
		}
	}
}

func (g *Gossiper) searchForOwnedFiles(msg utils.SearchRequest) utils.SearchReply {
	results := []*utils.SearchResult{}

	for _, v := range g.storedFiles {
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
	return utils.SearchReply{Origin: g.name, Destination: msg.Origin, HopLimit: utils.HOPLIMIT_CONSTANT, Results: results}
}

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

func (g *Gossiper) getAvailableFileResults(keywords []string) []utils.FileSkeleton {
	results := []utils.FileSkeleton{}
	g.chronReceivedFilesLock.Lock()

	if keywords == nil {
		for _, v := range g.chronReceivedFiles {
			results = append(results, utils.FileSkeleton{Name: v.File.Name, MetafileHash: utils.StringHash(v.File.MetafileHash)})
		}
	} else {
		for _, v := range g.chronReceivedFiles {
			if utils.Contains(keywords, v.Name) {
				results = append(results, utils.FileSkeleton{Name: v.File.Name, MetafileHash: utils.StringHash(v.File.MetafileHash)})
			}
		}
	}

	g.chronReceivedFilesLock.Unlock()
	return results
}
