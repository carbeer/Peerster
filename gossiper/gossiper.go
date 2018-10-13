package gossiper

import (
	"fmt"
	"log"
	"net"
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
}

func NewGossiper(gossipIp, name string, gossipPort, clientPort int, peers []string) *Gossiper {
	log.Println("Creating a new gossiper")
	udpAddr, _ := net.ResolveUDPAddr("udp4", fmt.Sprintf("%s:%d", gossipIp, gossipPort))
	udpConn, _ := net.ListenUDP("udp4", udpAddr)
	clientAddr, _ := net.ResolveUDPAddr("udp4", fmt.Sprintf("%s:%d", gossipIp, clientPort))
	clientConn, _ := net.ListenUDP("udp4", clientAddr)
	return &Gossiper{
		address:    *udpAddr,
		udpConn:    *udpConn,
		clientConn: *clientConn,
		name:       name,
		peers:      peers,
	}
}

func (g *Gossiper) ListenClientMessages(quit chan bool) {
	log.Println("Entering ListenClientMessages")
	fmt.Println("address of g", g.address.String())
	fmt.Println("peers has value", g.peers)
	// Listen infinitely long
	for {
		buffer := make([]byte, 4096)
		n, clientAddr, e := g.clientConn.ReadFromUDP(buffer)
		if e != nil {
			log.Fatal(e)
		}
		log.Println(clientAddr)
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
	log.Println("Entering ListenPeerMessages")
	for {
		buffer := make([]byte, 4096)
		log.Println("Listening for peer messages")
		n, peerAddr, e := g.udpConn.ReadFromUDP(buffer)
		if e != nil {
			log.Fatal(e)
		}
		log.Println(peerAddr)
		msg := &utils.SimpleMessage{}
		// TODO: Larger Messages!
		log.Println("Decoding message from peer")
		e = protobuf.Decode(buffer[:n], msg)
		if e != nil {
			log.Fatal(e)
		}
		fmt.Printf("SIMPLE MESSAGE origin %s from %s contents %s\n", msg.OriginalName, msg.RelayPeerAddr, msg.Contents)
		fmt.Printf("PEERS %v\n", fmt.Sprint(strings.Join(g.peers, ",")))
		go g.peerMessageHandler(*msg)
	}
	quit <- true
}

func (g *Gossiper) clientMessageHandler(msg string) {
	log.Println("Entering clientMessageHandler")
	peerMessage := utils.SimpleMessage{OriginalName: g.name, RelayPeerAddr: g.address.String(), Contents: msg}
	packetBytes, e := protobuf.Encode(&peerMessage)
	if e != nil {
		log.Fatal(e)
	}
	for _, p := range g.peers {
		g.sendToPeer(packetBytes, p)
	}
}

func (g *Gossiper) addPeerToListIfApplicable(address string) {
	for i := range g.peers {
		if g.peers[i] == address {
			return
		}
	}
	g.peers = append(g.peers, address)
}

func (g *Gossiper) peerMessageHandler(msg utils.SimpleMessage) {
	log.Println("Origin peer")
	originPeer := msg.RelayPeerAddr
	log.Println(originPeer)

	g.addPeerToListIfApplicable(originPeer)
	msg.RelayPeerAddr = g.address.String()
	packetBytes, e := protobuf.Encode(&msg)
	if e != nil {
		log.Fatal(e)
	}
	// Broadcast to everyone except originPeer
	for _, p := range g.peers {
		if p != originPeer {
			g.sendToPeer(packetBytes, p)
		}
	}
}

func (g *Gossiper) sendToPeer(byteStream []byte, targetIpPort string) {
	fmt.Printf("Origin peer %s is sending to %s\n", g.address.String(), targetIpPort)
	peerConn, e := net.Dial("udp4", targetIpPort)
	if e != nil {
		log.Fatal(e)
	}

	_, e = peerConn.Write(byteStream)
	if e != nil {
		log.Fatal(e)
	}
}

func main() {

}
