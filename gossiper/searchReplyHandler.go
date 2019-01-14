package gossiper

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/carbeer/Peerster/utils"
)

// Handles incoming SearchReplies
func (g *Gossiper) searchReplyHandler(msg utils.SearchReply) {
	if msg.Destination == g.Name {
		g.getMatchingSearchRequests(msg)
	} else {
		msg.HopLimit -= 1
		if msg.HopLimit <= 0 {
			return
		}
		gossipMessage := utils.GossipPacket{SearchReply: &msg}
		g.sendToPeer(gossipMessage, g.getNextHop(msg.Destination).Address)
	}
}

// Checks whether the SearchRequest was already processed within the past
func (g *Gossiper) getMatchingSearchRequests(key utils.SearchReply) {
	g.cachedSearchRequestsLock.Lock()
	for _, v := range g.CachedSearchRequests {
		if v.Request.Origin != g.Name {
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

// Updates externalFiles and returns the number of **new** matches
func (g *Gossiper) updateExternalFile(key utils.SearchResult, value string, request utils.SearchRequest) uint32 {
	matches := uint32(0)
	found := false
	g.externalFilesLock.Lock()
	// Entirely new file
	if reflect.ValueOf(g.ExternalFiles[utils.StringHash(key.MetafileHash)]).IsNil() {
		tmp := &utils.ExternalFile{MissingChunksUntilMatch: key.ChunkCount, File: utils.File{Name: key.FileName, MetafileHash: key.MetafileHash}, Holder: make([][]string, key.ChunkCount+1)}
		for ix, _ := range tmp.Holder {
			tmp.Holder[ix] = []string{}
		}
		tmp.Holder[0] = []string{value}
		g.ExternalFiles[utils.StringHash(key.MetafileHash)] = tmp
		found = true
	}
	// corresponds to index within the slice
	for _, val := range key.ChunkMap {
		fmt.Printf("Searching for chunk %d\n", val)
		// already known Chunk but new holder
		if len(g.ExternalFiles[utils.StringHash(key.MetafileHash)].Holder[val]) > 0 && !utils.Contains(g.ExternalFiles[utils.StringHash(key.MetafileHash)].Holder[val], value) {
			g.ExternalFiles[utils.StringHash(key.MetafileHash)].Holder[val] = append(g.ExternalFiles[utils.StringHash(key.MetafileHash)].Holder[val], value)
			found = true
			// new Chunk
		} else {
			g.ExternalFiles[utils.StringHash(key.MetafileHash)].Holder[val] = []string{value}
			tmp := g.ExternalFiles[utils.StringHash(key.MetafileHash)]
			tmp.MissingChunksUntilMatch -= 1
			g.ExternalFiles[utils.StringHash(key.MetafileHash)] = tmp
			found = true

			if g.ExternalFiles[utils.StringHash(key.MetafileHash)].MissingChunksUntilMatch == 0 {
				g.addChronReceivedFiles(g.ExternalFiles[utils.StringHash(key.MetafileHash)])
				matches += 1
			}
		}
	}
	g.externalFilesLock.Unlock()
	if found {
		fmt.Printf("FOUND match %s at %s metafile=%s chunks=%s\n", key.FileName, value, utils.StringHash(key.MetafileHash), strings.Trim(fmt.Sprint(key.ChunkMap), "[]"))
	}
	return matches
}
