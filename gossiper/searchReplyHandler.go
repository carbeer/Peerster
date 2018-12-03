package gossiper

import (
	"encoding/hex"
	"fmt"
	"log"
	"reflect"
	"strings"
	"time"

	"github.com/carbeer/Peerster/utils"
)

func (g *Gossiper) searchReplyHandler(msg utils.SearchReply) {
	if msg.Destination == g.name {
		fmt.Println("Received a search reply")
		g.getMatchingSearchRequests(msg)
	} else {
		msg.HopLimit -= 1
		if msg.HopLimit <= 0 {
			log.Printf("%s: ATTENTION: Dropping a search reply message for %s\n", g.name, msg.Destination)
			return
		}
		gossipMessage := utils.GossipPacket{SearchReply: &msg}

		fmt.Printf("%d: Send the search reply message\n", time.Now().Second())
		g.sendToPeer(gossipMessage, g.getNextHop(msg.Destination).Address)
	}
}

func (g *Gossiper) getMatchingSearchRequests(key utils.SearchReply) {

	g.cachedSearchRequestsLock.Lock()
	for _, v := range g.CachedSearchRequests {
		if v.Request.Origin != g.name {
			continue
		}
		for _, result := range key.Results {
			if result.ChunkCount == 0 || reflect.ValueOf(result.ChunkMap).IsNil() {
				continue
			}

			found := false
			for _, keyword := range v.Request.Keywords {
				if strings.Contains(result.FileName, keyword) {
					fmt.Printf("Cached Request %+v matched result %s\n", v, result.FileName)
					found = true
					break // Request satisfies at least one request keyword
				}
			}
			if !found {
				break // Files were returned that didn't match any pattern of a request
			}

			matches := g.updateExternalFile(*result, key.Origin, v.Request)
			g.searchRequestChannelLock.Lock()
			if !reflect.ValueOf(g.searchRequestChannel[v.Request.GetKeywordIdentifier()]).IsNil() {
				g.searchRequestChannel[v.Request.GetKeywordIdentifier()] <- matches
			}
			g.searchRequestChannelLock.Unlock()
		}
	}
	g.cachedSearchRequestsLock.Unlock()
}

// Get matching searchrequests
// update external files --> set reference on chronological search requests

// Updates externalFiles and returns the number of **new** matches
func (g *Gossiper) updateExternalFile(key utils.SearchResult, value string, request utils.SearchRequest) uint32 {
	matches := uint32(0)
	found := false
	g.externalFilesLock.Lock()
	// Entirely new file
	if reflect.ValueOf(g.externalFiles[key.FileName]).IsNil() {
		g.externalFiles[key.FileName] = &utils.ExternalFile{MissingChunksUntilMatch: key.ChunkCount, File: utils.File{FileName: key.FileName, MetaHash: hex.EncodeToString(key.MetafileHash)}, Holder: make([][]string, key.ChunkCount)}
		found = true
	}
	// val-1 corresponds to index within the slice
	for _, val := range key.ChunkMap {
		fmt.Printf("Searching for chunk %d\n", val-1)
		// already known Chunk but new holder
		if g.externalFiles[key.FileName].Holder[val-1] != nil && !utils.Contains(g.externalFiles[key.FileName].Holder[val-1], value) {
			g.externalFiles[key.FileName].Holder[val-1] = append(g.externalFiles[key.FileName].Holder[val-1], value)
			found = true
			// new Chunk
		} else {
			g.externalFiles[key.FileName].Holder[val-1] = []string{value}
			tmp := g.externalFiles[key.FileName]
			tmp.MissingChunksUntilMatch -= 1
			g.externalFiles[key.FileName] = tmp
			found = true

			if g.externalFiles[key.FileName].MissingChunksUntilMatch == 0 {
				fmt.Printf("Found all chunks for file %s :-)\n", key.FileName)
				g.addChronReceivedFiles(request, g.externalFiles[key.FileName])
				matches += 1
			}
		}
	}
	g.externalFilesLock.Unlock()
	if found {
		fmt.Printf("FOUND match %s at node %s metafile=%s chunks=%s\n", key.FileName, value, utils.StringHash(key.MetafileHash), strings.Trim(fmt.Sprint(key.ChunkMap), "[]"))
	} else {
		fmt.Printf("No match found for result %+v\n", key)
	}
	return matches
}