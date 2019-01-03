package gossiper

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strings"
	"sync"

	"github.com/carbeer/Peerster/utils"
	"github.com/dedis/protobuf"
)

type Gossiper struct {
	// Address ip:port on which the Gossiper instance runs
	Address net.UDPAddr
	Name    string
	// Addresses of all peers
	Peers     []string
	Simple    bool
	IdCounter uint32
	// Sorted list of received messages
	ReceivedMessages map[string][]utils.RumorMessage
	PrivateMessages  map[string][]utils.PrivateMessage
	// By metahash
	ExternalFiles        map[string]*utils.ExternalFile
	CachedSearchRequests map[string]utils.CachedRequest
	ChronRumorMessages   []utils.StoredMessage
	ChronPrivateMessages map[string][]utils.StoredMessage
	ChronReceivedFiles   []*utils.ExternalFile
	// next hop map Origin --> Address
	NextHop map[string]utils.HopInfo
	// Get utils.File by Metahash
	StoredFiles  map[string]utils.File
	StoredChunks map[string][]byte
	BlockHistory map[utils.Hash]utils.BlockWrapper
	// Get block by hash of the previour block
	DetachedBlocks      map[utils.Hash]utils.Block
	LastBlock           utils.BlockWrapper
	PendingTransactions []utils.TxPublish
	ReceivedBlock       bool

	// UDP connection for peers
	udpConn net.UDPConn
	// Client connecteion (UI)
	clientConn net.UDPConn
	// Identifier of the Gossiper
	miner chan bool
	// Tracks the status packets from rumorMongerings owned by this peer
	rumorMongeringChannel map[string]chan utils.StatusPacket
	dataRequestChannel    map[string]chan bool
	destinationSpecified  map[string]bool
	searchRequestChannel  map[string]chan uint32
	// The next hash to be requested given the current hash
	requestedChunks map[string]utils.ChunkInfo
	privateKey      rsa.PrivateKey

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

		udpConn:               *udpConn,
		clientConn:            *clientConn,
		destinationSpecified:  make(map[string]bool),
		rumorMongeringChannel: make(map[string]chan utils.StatusPacket, 10240),
		dataRequestChannel:    make(map[string]chan bool),
		searchRequestChannel:  make(map[string]chan uint32),
		miner:                 make(chan bool, 1024),

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

	if name == "" {
		g.privateKey = *utils.GenerateKeyPair()
		g.Name = utils.PublicKeyAsString(&g.privateKey)
	} else {
		log.Println("No RSA keys generated. Asymmetric encryption is therefore not supported.")
	}
	return g
}

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

func (g *Gossiper) GetAllPeers() string {
	return strings.Join(g.Peers, "\n")
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

func (g *Gossiper) SaveState() {
	obj, e := json.MarshalIndent(g, "", "\t")
	utils.HandleError(e)
	_ = os.Mkdir(utils.STATE_FOLDER, os.ModePerm)
	e = ioutil.WriteFile(fmt.Sprint(utils.STATE_PATH, g.Name, ".json"), obj, 0644)
	utils.HandleError(e)
}

func LoadState(id string) *Gossiper {
	g := Gossiper{}
	f, e := ioutil.ReadFile(fmt.Sprint(utils.STATE_PATH, id, ".json"))
	utils.HandleError(e)
	json.Unmarshal(f, &g)
	log.Printf("Got this: %+v\n", &g)
	return &g
}
