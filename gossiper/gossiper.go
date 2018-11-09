package gossiper

import (
	"fmt"
	"math/rand"
	"net"
	"strings"
	"sync"
	"time"

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
	// Sorted list of received messages
	ReceivedMessages map[string]utils.RumorMessages
	PrivateMessages  map[string][]utils.PrivateMessage
	// Keeps track of wanted messages
	WantedMessages map[string]uint32
	// Tracks the status packets from rumorMongerings owned by this peer
	rumorMongeringChannel map[string]chan utils.StatusPacket
	// next hop map Origin --> Address
	nextHop map[string]string

	// Locks for maps
	receivedMessagesLock      sync.RWMutex
	privateMessagesLock       sync.RWMutex
	wantedMessagesLock        sync.RWMutex
	rumorMongeringChannelLock sync.RWMutex
	nextHopLock               sync.RWMutex
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
		WantedMessages:            make(map[string]uint32),
		rumorMongeringChannel:     make(map[string]chan utils.StatusPacket, 2048),
		nextHop:                   make(map[string]string),
		receivedMessagesLock:      sync.RWMutex{},
		privateMessagesLock:       sync.RWMutex{},
		wantedMessagesLock:        sync.RWMutex{},
		rumorMongeringChannelLock: sync.RWMutex{},
		nextHopLock:               sync.RWMutex{},
	}
}

func (g *Gossiper) addPeerToListIfApplicable(adr string) {
	for i := range g.peers {
		if g.peers[i] == adr {
			return
		}
	}
	g.peers = append(g.peers, adr)
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
	}
	origin := msg.Origin
	var wg sync.WaitGroup
	// Check whether the message is desired
	if origin != g.name && (g.getWantedMessages(origin) == msg.ID || (g.getWantedMessages(origin) == 0 && msg.ID == 1)) {
		g.addToKnownMessages(msg)

		// If it's a desired message, update the next hop
		g.updateNextHop(origin, sender)

		wg.Add(1)
		go func() {
			g.startRumorMongering(msg)
			wg.Done()
		}()
	}
	wg.Add(1)
	go func() {
		g.sendAcknowledgement(sender)
		wg.Done()
	}()
	wg.Wait()
}

func (g *Gossiper) statusMessageHandler(msg utils.StatusPacket, sender string) {
	var wg sync.WaitGroup

	fmt.Printf("STATUS from %s%s\n", sender, msg.ToString())
	fmt.Printf("PEERS %v\n", fmt.Sprint(strings.Join(g.peers, ",")))

	if g.getRumorMongeringChannel(sender) != nil {
		// fmt.Printf("Owener of the channel with %s", sender)
		g.sendToRumorMongeringChannel(sender, msg)
	} else {
		// fmt.Printf("Not the owener of the channel with %s", sender)
		if g.HasLessMessagesThan(msg) {
			wg.Add(1)
			go func() {
				g.sendAcknowledgement(sender)
				wg.Done()
			}()
		} else if msgAv, msg := g.AdditionalMessages(msg); msgAv {
			// fmt.Printf("Sending requested message %s %d to %s\n", msg.Origin, msg.ID, sender)
			wg.Add(1)
			go func() {
				g.startRumorMongeringConnection(sender, utils.GossipPacket{Rumor: &msg})
				wg.Done()
			}()
		} else {
			// This case will only happen to the initiator of the rumor mongering
			fmt.Printf("IN SYNC WITH %s\n", sender)
		}
	}
	wg.Wait()
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
			g.sendToPeer(gossipMessage, g.getNextHop(msg.Destination))
			wg.Done()
		}()
		wg.Wait()
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
	var wg sync.WaitGroup
	// Create a channel that is added to the list of owned rumorMongergings
	g.setRumorMongeringChannel(peer, make(chan utils.StatusPacket, 2048))

	wg.Add(1)
	go func() {
		fmt.Printf("%d: Initiating rumor mongering connection \n", time.Now().Second())
		g.sendToPeer(gossipPacket, peer)
		wg.Done()
	}()
Loop:
	for {
		// Make a new channel
		timeout := make(chan bool)
		go utils.TimeoutCounter(timeout, utils.GetRumorMongeringTimeout())

		select {
		case <-timeout:
			fmt.Printf("%d: TIMEOUT\n", time.Now().Second())
			break Loop
		case msg := <-g.getRumorMongeringChannel(peer):
			// Send out additional messages that were requested
			if new, newMsg := g.AdditionalMessages(msg); new {
				fmt.Printf("%d: Sending one of the unknown messages: %s %d\n", time.Now().Second(), newMsg.Origin, newMsg.ID)
				g.sendToPeer(utils.GossipPacket{Rumor: &newMsg}, peer)
			} else if g.HasLessMessagesThan(msg) {
				// If no new messages requested, check whether peer has unknown messages and request them
				wg.Add(1)
				go func() {
					fmt.Printf("%d: Sending acknowledgement to %s to get unknown message\n", time.Now().Second(), peer)
					g.sendAcknowledgement(peer)
					wg.Done()
				}()
			} else {
				// The two peers are in sync
				fmt.Printf("IN SYNC WITH %s\n", peer)
				break Loop
			}
		}
	}
	// Drop this rumorMongering
	g.setRumorMongeringChannel(peer, nil)
	// Don't interrupt any ongoing process
	wg.Wait()
	return utils.FlipCoin()
}

func (g *Gossiper) generateStatusPacket() utils.StatusPacket {
	packet := utils.StatusPacket{}
	g.wantedMessagesLock.RLock()
	for k, v := range g.WantedMessages {
		peer := utils.PeerStatus{Identifier: k, NextID: v}
		//fmt.Printf("Want: %s %d\n", k, v)
		packet.Want = append(packet.Want, peer)
	}
	g.wantedMessagesLock.RUnlock()
	return packet
}

func (g *Gossiper) sendAcknowledgement(adr string) {
	statusPacket := g.generateStatusPacket()
	gossipPacket := utils.GossipPacket{Status: &statusPacket}
	g.sendToPeer(gossipPacket, adr)
}

func (g *Gossiper) sendToPeer(gossipPacket utils.GossipPacket, targetIpPort string) {
	if gossipPacket.Rumor != nil {
		fmt.Printf("MONGERING with %s\n", targetIpPort)
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
				// fmt.Printf("%d: Broadcasting\n", time.Now().Second())
				g.sendToPeer(packet, p)
				wg.Done()
			}(p)
		}
	}
	wg.Wait()
}

// Checks whether this gossiper could get additional messages from the peer that sent the status packet
func (g *Gossiper) HasLessMessagesThan(status utils.StatusPacket) bool {
	for i := range status.Want {
		id := status.Want[i].Identifier
		// Check if Origin and IDs all known
		if g.getWantedMessages(id) == 0 || g.getWantedMessages(id) < status.Want[i].NextID {
			return true
		}
	}
	return false
}

// Checks whether this gossiper has additional messages for the peer that sent the status packet
func (g *Gossiper) AdditionalMessages(status utils.StatusPacket) (bool, utils.RumorMessage) {
	msgAv := false
	var msg utils.RumorMessage

	// Check whether all identifiers exist
	g.wantedMessagesLock.RLock()
	for id, _ := range g.WantedMessages {
		hasId := false
		for i := range status.Want {
			if status.Want[i].Identifier == id {
				// Found the id in database. Not new for other peer.
				hasId = true
				break
			}
		}
		if !hasId {
			// Add the id to the missingIds
			msgAv = true
			msg = g.ReceivedMessages[id].GetById(1)
			break
		}
	}
	g.wantedMessagesLock.RUnlock()
	if msgAv {
		return msgAv, msg
	}

	g.receivedMessagesLock.RLock()
	// Add missing single messages
	for _, ps := range status.Want {
		// Check for new messages for identifier
		if g.getWantedMessages(ps.Identifier) > ps.NextID {
			msg = g.ReceivedMessages[ps.Identifier].GetById(int(ps.NextID))
			msgAv = true
		}
	}
	g.receivedMessagesLock.RUnlock()
	return msgAv, msg
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
	g.setWantedMessages(msg.Origin, msg.ID+1)
	g.appendReceivedMessages(msg.Origin, msg)
}

func (g *Gossiper) updateNextHop(origin, sender string) {
	g.setNextHop(origin, sender)
	fmt.Printf("DSDV %s %s\n", origin, sender)
}

func (g *Gossiper) newRumorMongeringMessage(msg string) {
	var wg sync.WaitGroup
	rumorMessage := utils.RumorMessage{Origin: g.name, ID: g.idCounter, Text: msg}
	g.idCounter = g.idCounter + 1
	g.addToKnownMessages(rumorMessage)
	wg.Add(1)
	go func() {
		g.startRumorMongering(rumorMessage)
		wg.Done()
	}()
	wg.Wait()
}

func (g *Gossiper) newPrivateMessage(msg utils.Message) {
	var wg sync.WaitGroup
	privateMessage := utils.PrivateMessage{Origin: g.name, ID: 0, Text: msg.Text, Destination: msg.Destination, HopLimit: utils.GetHopLimitConstant()}
	gossipMessage := utils.GossipPacket{Private: &privateMessage}
	wg.Add(1)
	go func() {
		fmt.Printf("SENDING PRIVATE MESSAGE %s TO %s\n", msg.Text, msg.Destination)
		g.sendToPeer(gossipMessage, g.getNextHop(msg.Destination))
		wg.Done()
	}()
	wg.Wait()
}
