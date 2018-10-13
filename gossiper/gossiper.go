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

func (g *Gossiper) ListenClientMessages() {
	// Listen infinitely long
	for {
		buffer := make([]byte, 4096)
		n, clientAddr, e := g.clientConn.ReadFromUDP(buffer)
		if e != nil {
			log.Fatal(e)
		}
		log.Println(clientAddr)
		msg := &utils.SimpleMessage{}
		// TODO: Larger Messages!
		e = protobuf.Decode(buffer[:n], msg)
		if e != nil {
			log.Fatal(e)
		}
		fmt.Println("CLIENT MESSAGE", msg.Text)
		fmt.Printf("PEERS %v", g.peers)

		// broadcast message to all peers
	}
}

func (g *Gossiper) ListenPeerMessages() {
	for {
		buffer := make([]byte, 4096)
		n, peerAddr, e := g.clientConn.ReadFromUDP(buffer)
		if e != nil {
			log.Fatal(e)
		}
		log.Println(peerAddr)
		msg := &utils.SimpleMessage{}
		// TODO: Larger Messages!
		e = protobuf.Decode(buffer[:n], msg)
		if e != nil {
			log.Fatal(e)
		}
		log.Println("SIMPLE MESSAGE origin %v from %v contents %v", msg.OriginalName, msg.RelayPeerAddr, msg.Contents)
		fmt.Printf("PEERS %v", g.peers)
		// broadcast message to all peers
	}
}

func (g *Gossiper) SendMessage(msg string) {
	packetBytes, _ := protobuf.Encode(msg)
	g.udpConn.WriteToUDP(packetBytes, &g.address)
	log.Println("Sent message")
}

func main() {
	// msg := &utils.Message{"hello"}

}
