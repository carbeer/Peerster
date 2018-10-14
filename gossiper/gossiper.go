package gossiper

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/carbeer/Peerster/utils"
	"github.com/dedis/protobuf"
)

type Gossiper struct {
	// Address ip:port on which the Gossiper instance runs
	address net.UDPAddr
	// UDP connection for peers
	udpConn net.UDPConn
	// Client connecteion (UI)
	clientConn net.UDPConn
	// Identifier of the Gossiper
	name string
	// Addresses of all peers
	peers []string
	// Sprted list of received messages
	ReceivedMessages map[string]utils.RumorMessages
	// Keeps track of wanted messages
	WantedMessages map[string]string
	simple         bool
	idCounter      int
}

func NewGossiper(gossipIp, name string, gossipPort, clientPort int, peers []string, simple bool) *Gossiper {
	udpAddr, _ := net.ResolveUDPAddr("udp4", fmt.Sprintf("%s:%d", gossipIp, gossipPort))
	udpConn, _ := net.ListenUDP("udp4", udpAddr)
	clientAddr, _ := net.ResolveUDPAddr("udp4", fmt.Sprintf("%s:%d", gossipIp, clientPort))
	clientConn, _ := net.ListenUDP("udp4", clientAddr)
	ReceivedMessages := make(map[string]utils.RumorMessages)
	WantedMessages := make(map[string]string)
	idCounter := int(1)
	return &Gossiper{
		address:          *udpAddr,
		udpConn:          *udpConn,
		clientConn:       *clientConn,
		name:             name,
		peers:            peers,
		ReceivedMessages: ReceivedMessages,
		WantedMessages:   WantedMessages,
		simple:           simple,
		idCounter:        idCounter,
	}
}

func (g *Gossiper) ListenClientMessages(quit chan bool) {
	fmt.Println("address of g", g.address.String())
	fmt.Println("peers has value", g.peers)
	// Listen infinitely long
	for {
		fmt.Println("Listening for client messages")
		buffer := make([]byte, 4096)
		n, _, e := g.clientConn.ReadFromUDP(buffer)
		if e != nil {
			log.Fatal(e)
		}
		msg := &utils.Message{}
		// TODO: Larger Messages!
		e = protobuf.Decode(buffer[:n], msg)
		if e != nil {
			log.Fatal(e)
		}
		fmt.Println("CLIENT MESSAGE", msg.Text)
		fmt.Printf("PEERS %v\n", fmt.Sprint(strings.Join(g.peers, ",")))

		go g.clientMessageHandler(msg.Text)
		// broadcast message to all peers
	}
	quit <- true
}

func (g *Gossiper) ListenPeerMessages(quit chan bool) {
	fmt.Println("address of g", g.address.String())
	fmt.Println("peers has value", g.peers)
	for {
		buffer := make([]byte, 4096)
		n, peerAddr, e := g.udpConn.ReadFromUDP(buffer)
		// fmt.Println("Peeraddr", peerAddr.String())
		// fmt.Println(fmt.Sprintf("%s:%d", peerAddr.IP, peerAddr.Port))
		if e != nil {
			log.Fatal(e)
		}

		msg := &utils.GossipPacket{}
		// TODO: Larger Messages!
		e = protobuf.Decode(buffer[:n], msg)
		if e != nil {
			log.Fatal(e)
		}
		go g.peerMessageHandler(*msg, peerAddr.String())
	}
	quit <- true
}

func (g *Gossiper) addPeerToListIfApplicable(adr string) {
	for i := range g.peers {
		if g.peers[i] == adr {
			return
		}
	}
	g.peers = append(g.peers, adr)
}

func (g *Gossiper) clientMessageHandler(msg string) {
	var gossipPacket utils.GossipPacket
	if g.simple {
		simpleMessage := utils.SimpleMessage{OriginalName: g.name, RelayPeerAddr: g.address.String(), Contents: msg}
		gossipPacket = utils.GossipPacket{Simple: &simpleMessage}
	} else {
		rumorMessage := utils.RumorMessage{Origin: g.name, ID: strconv.Itoa(g.idCounter), Text: msg}
		g.idCounter = g.idCounter + 1
		g.addToKnownMessages(rumorMessage)
		gossipPacket = utils.GossipPacket{Rumor: &rumorMessage}
	}
	for _, p := range g.peers {
		g.sendToPeer(gossipPacket, p)
	}
}

// TODO: Handle two non-nil and three nil cases!
func (g *Gossiper) peerMessageHandler(msg utils.GossipPacket, sender string) {
	g.addPeerToListIfApplicable(sender)
	if msg.Simple != nil {
		g.simpleBroadcast(*msg.Simple)
	} else if msg.Rumor != nil {
		g.rumorMongering(*msg.Rumor, sender)
	} else if msg.Status != nil {
		g.StatusProcessing(*msg.Status, sender)
	}
}

func (g *Gossiper) simpleBroadcast(msg utils.SimpleMessage) {
	var receivedFrom string = msg.RelayPeerAddr
	fmt.Printf("SIMPLE MESSAGE origin %s from %s contents %s\n", msg.OriginalName, msg.RelayPeerAddr, msg.Contents)
	fmt.Printf("PEERS %v\n", fmt.Sprint(strings.Join(g.peers, ",")))

	// No need to broadcast one's own message anymore
	if msg.OriginalName == g.name {
		return
	}
	msg.RelayPeerAddr = g.address.String()
	gossipPacket := utils.GossipPacket{Simple: &msg}
	// Broadcast to everyone except originPeer
	for _, p := range g.peers {
		if p != receivedFrom {
			g.sendToPeer(gossipPacket, p)
		}
	}
}

func (g *Gossiper) rumorMongering(msg utils.RumorMessage, sender string) {
	fmt.Printf("RUMOR origin %s from %s ID %s contents %s\n", msg.Origin, sender, msg.ID, msg.Text)
	fmt.Printf("PEERS %v\n", fmt.Sprint(strings.Join(g.peers, ",")))
	origin := msg.Origin

	if origin == g.name {
		return
	}
	// Check whether the message is desired
	if g.WantedMessages[origin] == msg.ID || (g.WantedMessages[origin] == "" && msg.ID == "1") {
		gossipPacket := utils.GossipPacket{Rumor: &msg}
		g.addToKnownMessages(msg)
		// TODO: Make this an independent Thread!
		randPeer := g.spreadRumorToRandomPeer(sender)
		g.sendToPeer(gossipPacket, randPeer)
	}
	g.sendAcknowledgement(sender)
}

func (g *Gossiper) spreadRumorToRandomPeer(origin string) string {
	for {
		newRand := rand.New(rand.NewSource(time.Now().UnixNano())).Intn(len(g.peers))
		if g.peers[newRand] != origin {
			return g.peers[newRand]
		}
	}
}

func (g *Gossiper) addToKnownMessages(msg utils.RumorMessage) {
	id, e := strconv.Atoi(msg.ID)
	if e != nil {
		log.Fatal(e)
	}
	g.WantedMessages[msg.Origin] = strconv.Itoa(id + 1)
	g.ReceivedMessages[msg.Origin] = append(g.ReceivedMessages[msg.Origin], msg)
}

func (g *Gossiper) StatusProcessing(msg utils.StatusPacket, sender string) {
	fmt.Printf("STATUS from %s%s\n", sender, msg.ToString())
	fmt.Printf("PEERS %v\n", fmt.Sprint(strings.Join(g.peers, ",")))
	if g.HasLessMessagesThan(msg) {
		log.Println("Request more messages from", g.name, sender)
		g.sendAcknowledgement(sender)
	} else if msgAv, msgs := g.AdditionalMessages(msg); msgAv {
		log.Println("Sending messages to", g.name, sender)
		sort.Sort(msgs)
		for _, rm := range msgs {
			gossipPacket := utils.GossipPacket{Rumor: &rm}
			g.sendToPeer(gossipPacket, sender)
		}
	} else {
		fmt.Printf("IN SYNC WITH %s\n", sender)
	}
}

func (g *Gossiper) generateStatusPacket() utils.StatusPacket {
	packet := utils.StatusPacket{}
	for k, v := range g.WantedMessages {
		id, e := strconv.Atoi(v)
		if e != nil {
			log.Fatal(e)
		}
		peer := utils.PeerStatus{Identifier: k, NextID: uint32(id)}
		packet.Want = append(packet.Want, peer)
	}
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
	if e != nil {
		log.Fatal(e)
	}
	adr, e := net.ResolveUDPAddr("udp4", targetIpPort)
	if e != nil {
		log.Fatal(e)
	}
	_, e = g.udpConn.WriteToUDP(byteStream, adr)
	if e != nil {
		log.Fatal(e)
	}
}

func (g *Gossiper) HasLessMessagesThan(status utils.StatusPacket) bool {
	for i := range status.Want {
		id := status.Want[i].Identifier
		// Check if Origin and IDs all known
		if g.WantedMessages[id] == "" || g.WantedMessages[id] < strconv.Itoa(int(status.Want[i].NextID)) {
			return true
		}
	}
	return false
}

func (g *Gossiper) AdditionalMessages(status utils.StatusPacket) (bool, utils.RumorMessages) {
	messages := []utils.RumorMessage{}
	missingIds := []string{}
	msgAv := false

	// Check whether all identifiers exist
	for id, _ := range g.WantedMessages {
		hasId := false
		for i := range status.Want {
			log.Println("Checking whether this equals", status.Want[i].Identifier, id)
			if status.Want[i].Identifier == id {
				log.Println("Found the id in database. Not new for other peer.")
				hasId = true
				break
			}
		}
		if !hasId {
			log.Println("Add the id to the missingIds")
			missingIds = append(missingIds, id)
			log.Println("Missing ids now", missingIds)
			msgAv = true
		}
	}
	// Add all messages for missing identifiers
	for _, id := range missingIds {
		log.Println("Adding all messages for missing identifiers")
		log.Println(id)
		for ix, _ := range g.ReceivedMessages[id] {
			log.Println(g.ReceivedMessages[id][ix])
			messages = append(messages, g.ReceivedMessages[id][ix])
			msgAv = true
		}
	}
	log.Println("New messages for peer found:", messages)

	// Add missing single messages
	for i, ps := range status.Want {
		log.Println("Check for new messages for identifier", ps.Identifier)
		j, _ := strconv.Atoi(g.WantedMessages[ps.Identifier])
		log.Println("Our peer has messages until", j-1)
		for j > int(status.Want[i].NextID) {
			messages = append(messages, g.ReceivedMessages[ps.Identifier][int(j-1)])
			msgAv = true
			j--
		}
	}
	return msgAv, messages
}
