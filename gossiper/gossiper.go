package gossiper

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"sort"
	"strconv"
	"strings"

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
	// List of queued messages
	messageQueue map[string]utils.RumorMessages
	// Keeps track of wanted messages
	wantedMessages map[string]string
	simple         bool
	idCounter      int
}

func NewGossiper(gossipIp, name string, gossipPort, clientPort int, peers []string, simple bool) *Gossiper {
	udpAddr, _ := net.ResolveUDPAddr("udp4", fmt.Sprintf("%s:%d", gossipIp, gossipPort))
	udpConn, _ := net.ListenUDP("udp4", udpAddr)
	clientAddr, _ := net.ResolveUDPAddr("udp4", fmt.Sprintf("%s:%d", gossipIp, clientPort))
	clientConn, _ := net.ListenUDP("udp4", clientAddr)
	messageQueue := make(map[string]utils.RumorMessages)
	wantedMessages := make(map[string]string)
	idCounter := int(0)
	return &Gossiper{
		address:        *udpAddr,
		udpConn:        *udpConn,
		clientConn:     *clientConn,
		name:           name,
		peers:          peers,
		messageQueue:   messageQueue,
		wantedMessages: wantedMessages,
		simple:         simple,
		idCounter:      idCounter,
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
		fmt.Printf("PEERS %v\n", g.peers)

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
	fmt.Println("Adding to list of known peers", adr)
	g.peers = append(g.peers, adr)
}

func (g *Gossiper) clientMessageHandler(msg string) {
	var gossipPacket utils.GossipPacket
	if g.simple {
		simpleMessage := utils.SimpleMessage{OriginalName: g.name, RelayPeerAddr: g.address.String(), Contents: msg}
		gossipPacket = utils.GossipPacket{Simple: &simpleMessage}
	} else {
		rumorMessage := utils.RumorMessage{Origin: g.name, ID: string(g.idCounter), Text: msg}
		g.idCounter = g.idCounter + 1
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
		g.statusProcessing(*msg.Status, sender)
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
	fmt.Printf("RUMOR MESSAGE origin %s with ID %s contents %s\n", msg.Origin, msg.ID, msg.Text)
	fmt.Printf("PEERS %v\n", fmt.Sprint(strings.Join(g.peers, ",")))
	origin := msg.Origin
	// Check whether the message is new
	if g.wantedMessages[origin] <= msg.ID {
		g.newRumor(msg)
	}
	g.sendAcknowledgement(sender)
}

func (g *Gossiper) statusProcessing(msg utils.StatusPacket, sender string) {
	fmt.Printf("STATUS MESSAGE want %s\n", msg.ToString())
	fmt.Printf("PEERS %v\n", fmt.Sprint(strings.Join(g.peers, ",")))
}

func (g *Gossiper) generateStatusPacket() utils.StatusPacket {
	packet := utils.StatusPacket{}
	for k, v := range g.wantedMessages {
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

func (g *Gossiper) newRumor(msg utils.RumorMessage) {
	origin := msg.Origin
	// msg is the next wanted message
	if g.wantedMessages[origin] == msg.ID {
		id, e := strconv.Atoi(msg.ID)
		if e != nil {
			log.Fatal(e)
		}
		var ix int = 1
		// More messages of this origin in queue
		if len(g.messageQueue[origin]) > 0 {
			sort.Sort(g.messageQueue[msg.Origin])

			log.Println("MQ ID", g.messageQueue[origin][ix-1].ID)
			log.Println("MSG ID", id+ix)
			// Find highest number ix of consecutive messages to update the next wanted message
			for g.messageQueue[origin][ix-1].ID == string(id+ix) {
				ix = ix + 1
			}
			// Cut off the updated values
			g.messageQueue[origin] = g.messageQueue[origin][ix-1:]
		}
		// Update the wanted message
		g.wantedMessages[origin] = string(id + ix)
	} else {
		// msg has to be queued because it's larger then the next wanted message
		g.messageQueue[origin] = append(g.messageQueue[origin], msg)
	}
	gossipPacket := utils.GossipPacket{Rumor: &msg}
	g.sendToPeer(gossipPacket, g.peers[rand.Intn(len(g.peers))])
}

func (g *Gossiper) sendToPeer(gossipPacket utils.GossipPacket, targetIpPort string) {
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
