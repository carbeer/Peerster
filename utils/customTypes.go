package utils

import (
	"fmt"
	"strings"
	"time"
)

type SimpleMessage struct {
	OriginalName  string
	RelayPeerAddr string
	Contents      string
}

type RumorMessage struct {
	Origin string
	ID     uint32
	Text   string
}

type PrivateMessage struct {
	Origin      string
	ID          uint32
	Text        string
	Destination string
	HopLimit    uint32
}

type DataRequest struct {
	Origin      string
	Destination string
	HopLimit    uint32
	HashValue   []byte
}

type DataReply struct {
	Origin      string
	Destination string
	HopLimit    uint32
	HashValue   []byte
	Data        []byte
}

type Message struct {
	Text        string   `json:"text"`
	Destination string   `json:"destination"`
	FileName    string   `json:"filename"`
	Request     string   `json:"request"`
	Keywords    []string `json:"keywords"`
	Budget      int64    `json:"budget"`
	Peer        string   `json:"peer"`
}

type TxPublish struct {
	File     File
	HopLimit uint32
}

type BlockPublish struct {
	Block    Block
	HopLimit uint32
}

type Block struct {
	PrevHash     [32]byte
	Nonce        [32]byte
	Transactions []TxPublish
}

type BlockWrapper struct {
	Block
	Counter int
}

type PeerStatus struct {
	// Equals the Origin in RumorMessage
	Identifier string
	NextID     uint32
}

type StatusPacket struct {
	Want []PeerStatus
}

type SearchRequest struct {
	Origin   string
	Budget   uint64
	Keywords []string
}

func (s *SearchRequest) GetIdentifier() string {
	return fmt.Sprintf("%s:%s", s.Origin, s.Keywords)
}

func (s *SearchRequest) GetLocalIdentifier() string {
	return fmt.Sprintf("%d:%s", s.Budget, s.Keywords)
}

func (s *SearchRequest) GetKeywordIdentifier() string {
	return strings.Join(s.Keywords, ",")
}

type SearchReply struct {
	Origin      string
	Destination string
	HopLimit    uint32
	Results     []*SearchResult
}

type SearchResult struct {
	FileName     string
	MetafileHash []byte
	ChunkMap     []uint64
	ChunkCount   uint64
}

func (sp *StatusPacket) ToString() string {
	s := ""
	for status := range sp.Want {
		s = fmt.Sprintf("%s peer %s nextID %v", s, sp.Want[status].Identifier, sp.Want[status].NextID)
	}
	return s
}

type File struct {
	Name         string
	Size         int64
	MetafileHash []byte
}

type FileSkeleton struct {
	Name         string `json:"fileName"`
	MetafileHash string `json:"metaHash"`
}

type CachedRequest struct {
	Timestamp time.Time
	Request   SearchRequest
}

type ExternalFile struct {
	File
	// Chunk holder by chunk id
	Holder                  [][]string
	MissingChunksUntilMatch uint64
}

type ChunkInfo struct {
	// Starts from 1
	ChunkNr int
	// Either NextHash (normal chunks) OR MetaHash (last chunk)
	MetaHash string
	NextHash string
	FileName string
}

type HopInfo struct {
	Address   string
	HighestID uint32
}

type GossipPacket struct {
	Simple        *SimpleMessage
	Rumor         *RumorMessage
	Status        *StatusPacket
	Private       *PrivateMessage
	DataRequest   *DataRequest
	DataReply     *DataReply
	SearchRequest *SearchRequest
	SearchReply   *SearchReply
	TxPublish     *TxPublish
	BlockPublish  *BlockPublish
}

type StoredMessage struct {
	Message   interface{}
	Timestamp time.Time
}
