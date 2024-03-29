package utils

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type GossipPacket struct {
	Simple              *SimpleMessage
	Rumor               *RumorMessage
	Status              *StatusPacket
	Private             *PrivateMessage
	DataRequest         *DataRequest
	DataReply           *DataReply
	SearchRequest       *SearchRequest
	SearchReply         *SearchReply
	TxPublish           *TxPublish
	BlockPublish        *BlockPublish
	FileExchangeRequest *FileExchangeRequest
	Challenge           *Challenge
}

type Message struct {
	Text         string   `json:"text"`
	Destination  string   `json:"destination"`
	FileName     string   `json:"filename"`
	Request      string   `json:"request"`
	Keywords     []string `json:"keywords"`
	Budget       int64    `json:"budget"`
	Peer         string   `json:"peer"`
	Encrypted    bool     `json:"encrypted"`
	Replications int      `json:"replications"`
}

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
	Origin        string
	ID            uint32
	Text          string
	EncryptedText string
	Destination   string
	HopLimit      uint32
	Signature     []byte `json:"-"` // Needs to be excluded due to the signing algorithm
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

type TxPublish struct {
	File     File
	HopLimit uint32
}

type BlockPublish struct {
	Block    Block
	HopLimit uint32
}

type Block struct {
	PrevHash     Hash
	Nonce        Hash
	Transactions []TxPublish
}

type BlockWrapper struct {
	Block
	Counter int
}

type PeerStatus struct {
	Identifier string // Equals the Origin in RumorMessage
	NextID     uint32
}

type StatusPacket struct {
	Want []PeerStatus
}

func (sp *StatusPacket) ToString() string {
	s := ""
	for status := range sp.Want {
		s = fmt.Sprintf("%s peer %s nextID %v", s, sp.Want[status].Identifier, sp.Want[status].NextID)
	}
	return s
}

type SearchRequest struct {
	Origin   string
	Budget   uint64
	Keywords []string
}

func (s *SearchRequest) GetIdentifier() string {
	return fmt.Sprintf("%s:%s", s.Origin, s.Keywords)
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

type File struct {
	Name         string
	Size         int64
	MetafileHash []byte
}

type PrivateFile struct {
	File
	Replications []Replica
}

type Replica struct {
	NodeID        string // Name of the remote peer storing the file
	EncryptionKey []byte // AES Encryption key
	ExchangeMFH   string // Exchange Metafilehash
	Metafilehash  string // Metafilehash of the encrypted file
}

type FileExchangeRequest struct {
	Origin               string
	Destination          string // Empty for OFFER, i.e. the initial broadcast
	Status               string // OFFER, ACCEPT, FIX
	HopLimit             uint32
	MetaFileHash         string // hex-representation of the Metafilehash of the Replica
	ExchangeMetaFileHash string // Empty for OFFER, i.e. the initial broadcast
}

type Challenge struct {
	Origin       string
	Destination  string
	MetaFileHash string
	ChunkHash    string // Targeted chunk
	Postpend     []byte // Data to be used for PoR
	Solution     []byte // Empty for request
	HopLimit     uint32
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

type StoredMessage struct {
	Message   interface{}
	Timestamp time.Time
}

type Hash [32]byte

func (h Hash) MarshalText() (text []byte, err error) {
	return []byte(StringHash(text)), nil
}

func (h *Hash) UnmarshalText(text []byte) error {
	var s string
	if e := json.Unmarshal(text, &s); e != nil {
		return e
	}
	*h = FixedByteHash(s)
	return nil
}
