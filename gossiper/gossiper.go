package gossiper

import (
	"fmt"
	"math/rand"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/carbeer/Peerster/utils"
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
	// Sorted list of received messages
	ReceivedMessages map[string]utils.RumorMessages
	PrivateMessages  map[string][]utils.PrivateMessage
	// Keeps track of wanted messages
	// Tracks the status packets from rumorMongerings owned by this peer
	rumorMongeringChannel map[string]chan utils.StatusPacket
	dataRequestChannel    map[string]chan bool
	// next hop map Origin --> Address
	nextHop map[string]utils.HopInfo
	// Get File by Metahash
	storedFiles  map[string]utils.File
	storedChunks map[string][]byte
	// The next hash to be quested given the current hash
	requestedChunks map[string]utils.ChunkInfo

	// Locks for maps
	receivedMessagesLock      sync.RWMutex
	privateMessagesLock       sync.RWMutex
	rumorMongeringChannelLock sync.RWMutex
	nextHopLock               sync.RWMutex
	storedFilesLock           sync.RWMutex
	storedChunksLock          sync.RWMutex
	requestedChunksLock       sync.RWMutex
	dataRequestChannelLock    sync.RWMutex
}

func NewGossiper(gossipIp, name string, gossipPort, clientPort int, peers []string, simple bool) *Gossiper {
	udpAddr, _ := net.ResolveUDPAddr("udp4", fmt.Sprintf("%s:%d", gossipIp, gossipPort))
	udpConn, _ := net.ListenUDP("udp4", udpAddr)
	clientAddr, _ := net.ResolveUDPAddr("udp4", fmt.Sprintf("%s:%d", utils.GetClientIp(), clientPort))
	clientConn, _ := net.ListenUDP("udp4", clientAddr)
	return &Gossiper{
		Address:                   *udpAddr,
		udpConn:                   *udpConn,
		ClientConn:                *clientConn,
		name:                      name,
		peers:                     peers,
		simple:                    simple,
		idCounter:                 uint32(1),
		ReceivedMessages:          make(map[string]utils.RumorMessages),
		PrivateMessages:           make(map[string][]utils.PrivateMessage),
		rumorMongeringChannel:     make(map[string]chan utils.StatusPacket, 10240),
		nextHop:                   make(map[string]utils.HopInfo),
		storedFiles:               make(map[string]utils.File),
		storedChunks:              make(map[string][]byte),
		requestedChunks:           make(map[string]utils.ChunkInfo),
		dataRequestChannel:        make(map[string]chan bool),
		receivedMessagesLock:      sync.RWMutex{},
		privateMessagesLock:       sync.RWMutex{},
		rumorMongeringChannelLock: sync.RWMutex{},
		nextHopLock:               sync.RWMutex{},
		storedFilesLock:           sync.RWMutex{},
		storedChunksLock:          sync.RWMutex{},
		requestedChunksLock:       sync.RWMutex{},
		dataRequestChannelLock:    sync.RWMutex{},
	}
}

func (g *Gossiper) simpleMessageHandler(msg utils.SimpleMessage) {
	fmt.Printf("SIMPLE MESSAGE origin %s from %s contents %s\n", msg.OriginalName, msg.RelayPeerAddr, msg.Contents)
	fmt.Printf("PEERS %v\n", fmt.Sprint(strings.Join(g.peers, ",")))
	// No need to broadcast one's own message anymore
	if msg.OriginalName == g.name {
		return
	}
	msg.RelayPeerAddr = g.Address.String()
	g.broadcastMessage(utils.GossipPacket{Simple: &msg}, msg.RelayPeerAddr)
}

func (g *Gossiper) rumorMessageHandler(msg utils.RumorMessage, sender string) {
	if msg.Text != "" {
		fmt.Printf("RUMOR origin %s from %s ID %d contents %s\n", msg.Origin, sender, msg.ID, msg.Text)
		fmt.Printf("PEERS %v\n", fmt.Sprint(strings.Join(g.peers, ",")))
	} else {
		fmt.Printf("ROUTE RUMOR origin %s from %s ID %d\n", msg.Origin, sender, msg.ID)
	}
	origin := msg.Origin
	var wg sync.WaitGroup

	// Check whether the message is desired
	if origin != g.name {
		g.updateNextHop(msg, sender)
		if len(g.getReceivedMessages(origin))+1 == int(msg.ID) {
			g.addToKnownMessages(msg)

			wg.Add(1)
			go func() {
				g.startRumorMongering(msg)
				wg.Done()
			}()
		}
	}
	g.sendAcknowledgement(sender)
	wg.Wait()
}

func (g *Gossiper) statusMessageHandler(status utils.StatusPacket, sender string) {

	fmt.Printf("STATUS from %s%s\n", sender, status.ToString())
	fmt.Printf("PEERS %v\n", fmt.Sprint(strings.Join(g.peers, ",")))

	if g.getRumorMongeringChannel(sender) != nil {
		fmt.Printf("Owner of the channel with %s\n", sender)
		g.sendToRumorMongeringChannel(sender, status)
	} else {
		if g.compareStatus(sender, status) {
			fmt.Printf("Not the owner of the channel with %s (IN SYNC)\n", sender)
		} else {
			fmt.Printf("Not the owner of the channel with %s (NOT IN SYNC)\n", sender)
		}
	}
}

func (g *Gossiper) privateMessageHandler(msg utils.PrivateMessage) {
	if msg.Destination == g.name {
		fmt.Printf("PRIVATE origin %s hop-limit %d contents %s\n", msg.Origin, msg.HopLimit, msg.Text)
		g.appendPrivateMessages(msg.Origin, msg)
	} else {
		var wg sync.WaitGroup
		msg.HopLimit -= 1
		if msg.HopLimit <= 0 {
			// log.Printf("%s: ATTENTION: Dropping a private message for %s\n", g.name, msg.Destination)
			return
		}
		gossipMessage := utils.GossipPacket{Private: &msg}
		wg.Add(1)
		go func() {
			// fmt.Printf("%d: Send the private message\n", time.Now().Second())
			g.sendToPeer(gossipMessage, g.getNextHop(msg.Destination).Address)
			wg.Done()
		}()
		wg.Wait()
	}
}

func (g *Gossiper) dataRequestHandler(msg utils.DataRequest, sender string) {
	if msg.Destination == g.name {
		g.newDataReplyMessage(msg, sender)
	} else {
		msg.HopLimit -= 1
		if msg.HopLimit <= 0 {
			// log.Printf("%s: ATTENTION: Dropping a private message for %s\n", g.name, msg.Destination)
			return
		}
		gossipMessage := utils.GossipPacket{DataRequest: &msg}
		// fmt.Printf("%d: Send the private message\n", time.Now().Second())
		g.sendToPeer(gossipMessage, g.getNextHop(msg.Destination).Address)
	}
}

func (g *Gossiper) dataReplyHandler(msg utils.DataReply) {
	if msg.Destination == g.name {
		g.receiveDataReply(msg)
	} else {
		msg.HopLimit -= 1
		if msg.HopLimit <= 0 {
			// log.Printf("%s: ATTENTION: Dropping a private message for %s\n", g.name, msg.Destination)
			return
		}
		gossipMessage := utils.GossipPacket{DataReply: &msg}
		// fmt.Printf("%d: Send the private message\n", time.Now().Second())
		g.sendToPeer(gossipMessage, g.getNextHop(msg.Destination).Address)
	}
}

func (g *Gossiper) startRumorMongering(msg utils.RumorMessage) {
	gossipPacket := utils.GossipPacket{Rumor: &msg}
	peer := g.pickRandomPeerForMongering("")
	for {
		if peer == "" {
			break
		}
		// If it return true, coinflip decided to continue mongering, else it stops
		if !g.startRumorMongeringConnection(peer, gossipPacket) {
			fmt.Println("Stopping rumormongering")
			break
		}
		peer = g.pickRandomPeerForMongering(peer)
		fmt.Println("FLIPPED COIN sending rumor to", peer)
	}
}

func (g *Gossiper) startRumorMongeringConnection(peer string, gossipPacket utils.GossipPacket) bool {
	// Create a channel that is added to the list of owned rumorMongergings
	g.setRumorMongeringChannel(peer, make(chan utils.StatusPacket, utils.GetMsgBuffer()))
	fmt.Printf("%d: Initiating rumor mongering connection \n", time.Now().Second())
	g.sendToPeer(gossipPacket, peer)

Loop:
	for {
		select {
		case <-time.After(utils.GetRumorMongeringTimeout()):
			fmt.Printf("%d: TIMEOUT\n", time.Now().Second())
			break Loop
		case status := <-g.getRumorMongeringChannel(peer):
			if g.compareStatus(peer, status) {
				// The two peers are in sync
				fmt.Printf("IN SYNC WITH %s\n", peer)
				break Loop
			}
		}
	}
	g.deleteRumorMongeringChannel(peer)
	return utils.FlipCoin()
}

// returns true if status is equal, else false
func (g *Gossiper) compareStatus(peer string, status utils.StatusPacket) bool {
	// Send out additional messages that were requested
	if new, newMsg := g.AdditionalMessages(status); new {
		fmt.Printf("%d: Sending one of the unknown messages: %s %d\n", time.Now().Second(), newMsg.Origin, newMsg.ID)
		g.sendToPeer(utils.GossipPacket{Rumor: &newMsg}, peer)
		return false
	} else if g.HasLessMessagesThan(status) {
		// If no new messages requested, check whether peer has unknown messages and request them
		fmt.Printf("%d: Sending acknowledgement to %s to get unknown message\n", time.Now().Second(), peer)
		g.sendAcknowledgement(peer)
		return false
	}
	return true
}

// Checks whether this gossiper could get additional messages from the peer that sent the status packet
func (g *Gossiper) HasLessMessagesThan(status utils.StatusPacket) bool {
	for i := range status.Want {
		id := status.Want[i].Identifier
		// Check if Origin and IDs all known
		if len(g.getReceivedMessages(id))+1 < int(status.Want[i].NextID) {
			fmt.Println("Have less message for ORIGIN", id)
			return true
		}
	}
	return false
}

// Checks whether this gossiper has additional messages for the peer that sent the status packet
func (g *Gossiper) AdditionalMessages(status utils.StatusPacket) (bool, utils.RumorMessage) {
	g.receivedMessagesLock.RLock()
	defer g.receivedMessagesLock.RUnlock()

Loop:
	for k, v := range g.ReceivedMessages {
		for index := range status.Want {
			if status.Want[index].Identifier == k {
				if int(status.Want[index].NextID) < len(v)+1 {
					return true, v[status.Want[index].NextID-1]
				}
				continue Loop
			}
		}
		return true, g.ReceivedMessages[k][0]
	}
	return false, utils.RumorMessage{}
}

func (g *Gossiper) generateStatusPacket() utils.StatusPacket {
	packet := utils.StatusPacket{}
	g.receivedMessagesLock.RLock()
	for k, v := range g.ReceivedMessages {
		peer := utils.PeerStatus{Identifier: k, NextID: uint32(len(v) + 1)}
		packet.Want = append(packet.Want, peer)
	}
	g.receivedMessagesLock.RUnlock()
	return packet
}

func (g *Gossiper) sendAcknowledgement(adr string) {
	statusPacket := g.generateStatusPacket()
	gossipPacket := utils.GossipPacket{Status: &statusPacket}
	g.sendToPeer(gossipPacket, adr)
}

func (g *Gossiper) pickRandomPeerForMongering(origin string) string {
	peer := ""
	for {
		newRand := rand.New(rand.NewSource(time.Now().UnixNano())).Intn(len(g.peers))
		if g.peers[newRand] != origin && g.getRumorMongeringChannel(peer) == nil {
			peer = g.peers[newRand]
			break
		}
	}
	return peer
}

func (g *Gossiper) addToKnownMessages(msg utils.RumorMessage) {
	g.appendReceivedMessages(msg.Origin, msg)
}

func (g *Gossiper) updateNextHop(msg utils.RumorMessage, sender string) {
	if msg.ID > g.getNextHop(msg.Origin).HighestID {
		g.setNextHop(msg.Origin, utils.HopInfo{Address: sender, HighestID: msg.ID})
		fmt.Printf("DSDV %s %s\n", msg.Origin, sender)
	}
}
