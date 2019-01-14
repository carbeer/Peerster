package gossiper

import (
	"crypto/rsa"
	"fmt"
	"log"
	"net"
	"runtime/debug"
	"strings"
	"sync"

	"github.com/carbeer/Peerster/utils"
	"github.com/dedis/protobuf"
)

type Gossiper struct {
	Address              net.UDPAddr                       // Address ip:port on which the Gossiper instance runs
	Name                 string                            // Identifier of the Gossiper
	Peers                []string                          // Addresses of all peers
	Simple               bool                              // Flag for Simple mode (use only simple messages)
	IdCounter            uint32                            // Counter for messageIDs
	ReceivedMessages     map[string][]utils.RumorMessage   // Sorted list of received messages by Name
	PrivateMessages      map[string][]utils.PrivateMessage // Private messages by Name
	ExternalFiles        map[string]*utils.ExternalFile    // By metahash
	CachedSearchRequests map[string]utils.CachedRequest    // By Name
	ChronRumorMessages   []utils.StoredMessage             // Chronologically stored rumor messages
	ChronPrivateMessages map[string][]utils.StoredMessage  // Chronologically stored private messages by Name
	ChronReceivedFiles   []*utils.ExternalFile             // Chronologically stored external files
	NextHop              map[string]utils.HopInfo          // next hop map returning Address by Name
	StoredFiles          map[string]utils.File             // utils.File by Metahash
	StoredChunks         map[string][]byte                 // Stored chunk data by chunk hash
	BlockHistory         map[utils.Hash]utils.BlockWrapper // Block by hash of the previous block
	DetachedBlocks       map[utils.Hash]utils.Block        // Detached block (not in main fork) by previous block hash
	LastBlock            utils.BlockWrapper                // Currently last block in the main fork of the chain
	PendingTransactions  []utils.TxPublish                 // Transactions to be mined in a block
	ReceivedBlock        bool                              // Bool val indicating whether gossiper already received a block
	PrivFiles            map[string]utils.PrivateFile      // PrivateFile by metafilehash
	Replications         map[string]*utils.Replica         `json:"-"` // Map of chunk hashes to Replica pointers (of PrivateFiles)

	udpConn               net.UDPConn                               // UDP connection for peers
	clientConn            net.UDPConn                               // Client connecteion (UI)
	miner                 chan bool                                 // Channel to notify the continously running miner
	rumorMongeringChannel map[string]chan utils.StatusPacket        // Tracks the status packets from rumorMongerings owned by this peer
	dataRequestChannel    map[string]chan bool                      // Channel by chunk hash, just to notify the process
	destinationSpecified  map[string]bool                           // By Name, returns whether file is downloaded from a specific node or from the entire network
	searchRequestChannel  map[string]chan uint32                    // Channel by SearchRequest keyword identifier, returns number of successful search requests
	fileDownloadChannel   map[string]chan bool                      // Channel by metahash, just to notify the process
	requestedChunks       map[string]utils.ChunkInfo                // The next hash to be requested given the current hash
	privateKey            rsa.PrivateKey                            // private RSA key
	fileExchangeChannel   map[string]chan utils.FileExchangeRequest // Channel by metafilehash, sends the FileExchangeRequest to the running process
	challengeChannel      map[string]chan utils.Challenge           // Channel by metafilehash, sends Challenge responses to running process

	// Locks for maps
	receivedMessagesLock      sync.RWMutex
	privateMessagesLock       sync.RWMutex
	rumorMongeringChannelLock sync.RWMutex
	nextHopLock               sync.RWMutex
	storedFilesLock           sync.RWMutex
	storedChunksLock          sync.RWMutex
	requestedChunksLock       sync.RWMutex
	dataRequestChannelLock    sync.RWMutex
	chronRumorMessagesLock    sync.RWMutex
	chronPrivateMessagesLock  sync.RWMutex
	cachedSearchRequestsLock  sync.RWMutex
	chronReceivedFilesLock    sync.RWMutex
	searchRequestChannelLock  sync.RWMutex
	externalFilesLock         sync.RWMutex
	chainLock                 sync.RWMutex
	challengeChannelLock      sync.RWMutex
	fileExchangeChannelLock   sync.RWMutex
	privFileLock              sync.RWMutex
}

func NewGossiper(gossipIp, name string, gossipPort, clientPort int, peers []string, simple bool) *Gossiper {
	udpAddr, _ := net.ResolveUDPAddr("udp4", fmt.Sprintf("%s:%d", gossipIp, gossipPort))
	udpConn, _ := net.ListenUDP("udp4", udpAddr)
	clientAddr, _ := net.ResolveUDPAddr("udp4", fmt.Sprintf("%s:%d", utils.CLIENT_IP, clientPort))
	clientConn, _ := net.ListenUDP("udp4", clientAddr)
	g := &Gossiper{
		Address:              *udpAddr,
		Name:                 name,
		Peers:                peers,
		Simple:               simple,
		ReceivedMessages:     make(map[string][]utils.RumorMessage),
		PrivateMessages:      make(map[string][]utils.PrivateMessage),
		StoredFiles:          make(map[string]utils.File),
		StoredChunks:         make(map[string][]byte),
		requestedChunks:      make(map[string]utils.ChunkInfo),
		NextHop:              make(map[string]utils.HopInfo),
		ChronReceivedFiles:   []*utils.ExternalFile{},
		ChronRumorMessages:   []utils.StoredMessage{},
		ChronPrivateMessages: make(map[string][]utils.StoredMessage),
		CachedSearchRequests: make(map[string]utils.CachedRequest),
		ExternalFiles:        make(map[string]*utils.ExternalFile),
		BlockHistory:         make(map[utils.Hash]utils.BlockWrapper),
		LastBlock:            utils.BlockWrapper{},
		DetachedBlocks:       make(map[utils.Hash]utils.Block),
		ReceivedBlock:        false,
		IdCounter:            uint32(1),
		PrivFiles:            make(map[string]utils.PrivateFile),

		udpConn:               *udpConn,
		clientConn:            *clientConn,
		destinationSpecified:  make(map[string]bool),
		rumorMongeringChannel: make(map[string]chan utils.StatusPacket, 10240),
		dataRequestChannel:    make(map[string]chan bool),
		searchRequestChannel:  make(map[string]chan uint32),
		miner:                 make(chan bool, 1024),
		fileExchangeChannel:   make(map[string]chan utils.FileExchangeRequest, 10240),
		challengeChannel:      make(map[string]chan utils.Challenge, 10240),
		Replications:          make(map[string]*utils.Replica),
		fileDownloadChannel:   make(map[string]chan bool, 1024),

		receivedMessagesLock:      sync.RWMutex{},
		privateMessagesLock:       sync.RWMutex{},
		rumorMongeringChannelLock: sync.RWMutex{},
		nextHopLock:               sync.RWMutex{},
		storedFilesLock:           sync.RWMutex{},
		storedChunksLock:          sync.RWMutex{},
		requestedChunksLock:       sync.RWMutex{},
		dataRequestChannelLock:    sync.RWMutex{},
		chronRumorMessagesLock:    sync.RWMutex{},
		chronPrivateMessagesLock:  sync.RWMutex{},
		cachedSearchRequestsLock:  sync.RWMutex{},
		chronReceivedFilesLock:    sync.RWMutex{},
		searchRequestChannelLock:  sync.RWMutex{},
		externalFilesLock:         sync.RWMutex{},
		chainLock:                 sync.RWMutex{},
		challengeChannelLock:      sync.RWMutex{},
		fileExchangeChannelLock:   sync.RWMutex{},
		privFileLock:              sync.RWMutex{},
	}

	if name == "" {
		g.privateKey = *utils.GenerateKeyPair()
		g.Name = utils.PublicKeyAsString(&g.privateKey)
	} else {
		log.Println("No RSA keys generated. Asymmetric encryption is therefore not supported.")
	}
	return g
}

// Restore gossiper from stored state
// NOT WORKING PROPERLY
func RestoreGossiper(gossipIp, name string, gossipPort, clientPort int, peers []string) *Gossiper {
	g := LoadState(name)
	udpAddr, e := net.ResolveUDPAddr("udp4", fmt.Sprintf("%s:%d", gossipIp, gossipPort))
	utils.HandleError(e)
	udpConn, e := net.ListenUDP("udp4", udpAddr)
	utils.HandleError(e)
	clientAddr, e := net.ResolveUDPAddr("udp4", fmt.Sprintf("%s:%d", utils.CLIENT_IP, clientPort))
	utils.HandleError(e)
	clientConn, e := net.ListenUDP("udp4", clientAddr)
	utils.HandleError(e)
	g.Address = *udpAddr
	g.udpConn = *udpConn
	g.clientConn = *clientConn

	g.destinationSpecified = make(map[string]bool)
	g.rumorMongeringChannel = make(map[string]chan utils.StatusPacket, 10240)
	g.dataRequestChannel = make(map[string]chan bool)
	g.searchRequestChannel = make(map[string]chan uint32)
	g.miner = make(chan bool, 1024)
	g.fileExchangeChannel = make(map[string]chan utils.FileExchangeRequest, 10240)
	g.challengeChannel = make(map[string]chan utils.Challenge, 10240)

	g.receivedMessagesLock = sync.RWMutex{}
	g.privateMessagesLock = sync.RWMutex{}
	g.rumorMongeringChannelLock = sync.RWMutex{}
	g.nextHopLock = sync.RWMutex{}
	g.storedFilesLock = sync.RWMutex{}
	g.storedChunksLock = sync.RWMutex{}
	g.requestedChunksLock = sync.RWMutex{}
	g.dataRequestChannelLock = sync.RWMutex{}
	g.chronRumorMessagesLock = sync.RWMutex{}
	g.chronPrivateMessagesLock = sync.RWMutex{}
	g.cachedSearchRequestsLock = sync.RWMutex{}
	g.chronReceivedFilesLock = sync.RWMutex{}
	g.searchRequestChannelLock = sync.RWMutex{}
	g.externalFilesLock = sync.RWMutex{}
	g.chainLock = sync.RWMutex{}

	// Join previously known peers with peers provided at startup
	for _, p := range peers {
		g.addPeerToListIfApplicable(p)
	}

	privateKey, e := utils.LoadKey(name)
	if e != nil || privateKey.Validate() != nil {
		log.Println("No RSA keys available. Asymmetric encryption is therefore not supported.")
	} else {
		g.privateKey = *privateKey
	}
	return g
}

// Send a packet to the neighbor with address targetIpPort
func (g *Gossiper) sendToPeer(gossipPacket utils.GossipPacket, targetIpPort string) {
	if targetIpPort == "" {
		fmt.Printf("No target address given. Not sending this message: %+v.\n", gossipPacket)
		debug.PrintStack()
		return
	}
	if gossipPacket.Rumor != nil {
		if gossipPacket.Rumor.Text != "" {
			fmt.Printf("MONGERING with %s\n", targetIpPort)
		} else {
			fmt.Printf("Route mongering with %s\n", targetIpPort)
		}
	}
	byteStream, e := protobuf.Encode(&gossipPacket)
	utils.HandleError(e)
	adr, e := net.ResolveUDPAddr("udp4", targetIpPort)
	utils.HandleError(e)
	_, e = g.udpConn.WriteToUDP(byteStream, adr)
	utils.HandleError(e)
}

// Broadcast packet to neighbors (except the one the packet was received from)
func (g *Gossiper) broadcastMessage(packet utils.GossipPacket, receivedFrom string) {
	var wg sync.WaitGroup
	for _, p := range g.Peers {
		if p != receivedFrom {
			wg.Add(1)
			go func(p string) {
				g.sendToPeer(packet, p)
				wg.Done()
			}(p)
		}
	}
	wg.Wait()
}

// Add peer to known neighbors
func (g *Gossiper) addPeerToListIfApplicable(adr string) {
	if adr == "" {
		return
	}
	if len(g.Peers) == 0 || (len(g.Peers) == 1 && g.Peers[0] == "") {
		g.Peers = []string{adr}
	} else {
		for i := range g.Peers {
			if g.Peers[i] == adr {
				return
			}
		}
		g.Peers = append(g.Peers, adr)
	}
}

// Returns all peers
func (g *Gossiper) GetAllPeers() string {
	return strings.Join(g.Peers, "\n")
}

// Updates the routing table
func (g *Gossiper) updateNextHop(origin string, id uint32, sender string) {
	if id > g.getNextHop(origin).HighestID || g.getNextHop(origin).HighestID == 0 {
		g.setNextHop(origin, utils.HopInfo{Address: sender, HighestID: id})
		fmt.Printf("DSDV %s %s\n", origin, sender)
	}
}
