package gossiper

import (
	"fmt"
	"log"
	"net"

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
		fmt.Printf("PEERS %v", g.peers)

		g.clientMessageHandler(msg.Text)
		// broadcast message to all peers
	}
	quit <- true
}

func (g *Gossiper) ListenPeerMessages(quit chan bool) {
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
		log.Println(fmt.Sprintf("SIMPLE MESSAGE origin %s from %s contents %s", msg.OriginalName, msg.RelayPeerAddr, msg.Contents))
		fmt.Printf("PEERS %v", g.peers)
		// broadcast message to other peers
	}
	quit <- true
}

func (g *Gossiper) clientMessageHandler(msg string) {
	log.Println("Entering clientMessageHandler")
	peerMessage := utils.SimpleMessage{OriginalName: g.name, RelayPeerAddr: g.address.IP.String(), Contents: msg}
	packetBytes, e := protobuf.Encode(&peerMessage)
	if e != nil {
		log.Fatal(e)
	}
	for _, p := range g.peers {
		g.sendToPeer(packetBytes, p)
	}
}

func (g *Gossiper) sendToPeer(byteStream []byte, targetIpPort string) {
	peerConn, e := net.Dial("udp4", targetIpPort)
	if e != nil {
		log.Fatal(e)
	}
	log.Println("Sending message from peer to peer")
	log.Println(targetIpPort)

	_, e = peerConn.Write(byteStream)
	if e != nil {
		log.Fatal(e)
	}
}

func main() {

}
