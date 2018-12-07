package gossiper

import (
	"fmt"
	"net"
	"sync"

	"github.com/carbeer/Peerster/utils"
	"github.com/dedis/protobuf"
)

type Gossiper struct {
	// Address ip:port on which the Gossiper instance runs
	Address net.UDPAddr
	// UDP connection for peers
	udpConn net.UDPConn
	// Client connecteion (UI)
	ClientConn net.UDPConn
	// Identifier of the Gossiper
	name string
	// Addresses of all peers
	peers     []string
	simple    bool
	idCounter uint32
	miner     chan bool
	// Sorted list of received messages
	ReceivedMessages map[string][]utils.RumorMessage
	PrivateMessages  map[string][]utils.PrivateMessage
	// By metahash
	externalFiles        map[string]*utils.ExternalFile
	CachedSearchRequests map[string]utils.CachedRequest
	chronRumorMessages   []utils.StoredMessage
	chronPrivateMessages map[string][]utils.StoredMessage
	chronReceivedFiles   []*utils.ExternalFile

	// Tracks the status packets from rumorMongerings owned by this peer
	rumorMongeringChannel map[string]chan utils.StatusPacket
	dataRequestChannel    map[string]chan bool
	destinationSpecified  map[string]bool
	searchRequestChannel  map[string]chan uint32
	// next hop map Origin --> Address
	nextHop map[string]utils.HopInfo
	// Get utils.File by Metahash
	storedFiles  map[string]utils.File
	storedChunks map[string][]byte
	// The next hash to be requested given the current hash
	requestedChunks     map[string]utils.ChunkInfo
	blockHistory        map[[32]byte]utils.BlockWrapper
	lastBlock           utils.BlockWrapper
	pendingTransactions []utils.TxPublish

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
}

func NewGossiper(gossipIp, name string, gossipPort, clientPort int, peers []string, simple bool) *Gossiper {
	udpAddr, _ := net.ResolveUDPAddr("udp4", fmt.Sprintf("%s:%d", gossipIp, gossipPort))
	udpConn, _ := net.ListenUDP("udp4", udpAddr)
	clientAddr, _ := net.ResolveUDPAddr("udp4", fmt.Sprintf("%s:%d", utils.CLIENT_IP, clientPort))
	clientConn, _ := net.ListenUDP("udp4", clientAddr)
	return &Gossiper{
		Address:                   *udpAddr,
		udpConn:                   *udpConn,
		ClientConn:                *clientConn,
		name:                      name,
		peers:                     peers,
		simple:                    simple,
		idCounter:                 uint32(1),
		ReceivedMessages:          make(map[string][]utils.RumorMessage),
		PrivateMessages:           make(map[string][]utils.PrivateMessage),
		rumorMongeringChannel:     make(map[string]chan utils.StatusPacket, 10240),
		nextHop:                   make(map[string]utils.HopInfo),
		storedFiles:               make(map[string]utils.File),
		storedChunks:              make(map[string][]byte),
		requestedChunks:           make(map[string]utils.ChunkInfo),
		dataRequestChannel:        make(map[string]chan bool),
		destinationSpecified:      make(map[string]bool),
		searchRequestChannel:      make(map[string]chan uint32),
		chronRumorMessages:        []utils.StoredMessage{},
		chronPrivateMessages:      make(map[string][]utils.StoredMessage),
		CachedSearchRequests:      make(map[string]utils.CachedRequest),
		chronReceivedFiles:        []*utils.ExternalFile{},
		externalFiles:             make(map[string]*utils.ExternalFile),
		blockHistory:              make(map[[32]byte]utils.BlockWrapper),
		lastBlock:                 utils.BlockWrapper{},
		miner:                     make(chan bool, 1024),
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
	}
}

func (g *Gossiper) sendToPeer(gossipPacket utils.GossipPacket, targetIpPort string) {
	if targetIpPort == "" {
		fmt.Printf("No target address given. Not sending this message: %+v.\n", gossipPacket)
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

func (g *Gossiper) broadcastMessage(packet utils.GossipPacket, receivedFrom string) {
	var wg sync.WaitGroup
	// Broadcast to everyone except originPeer
	for _, p := range g.peers {
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

func (g *Gossiper) addPeerToListIfApplicable(adr string) {
	if adr == "" {
		return
	}
	for i := range g.peers {
		if g.peers[i] == adr {
			return
		}
	}
	g.peers = append(g.peers, adr)
}

func (g *Gossiper) GetAllPeers() string {
	allPeers := ""
	for _, peer := range g.peers {
		allPeers = allPeers + peer + "\n"
	}
	return allPeers
}

func (g *Gossiper) GetAllOrigins() string {
	allOrigins := ""
	g.nextHopLock.RLock()
	for k, _ := range g.nextHop {
		allOrigins = allOrigins + k + "\n"
	}
	g.nextHopLock.RUnlock()
	return allOrigins
}
