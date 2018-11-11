package gossiper

import (
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/carbeer/Peerster/utils"
	"github.com/dedis/protobuf"
)

func (g *Gossiper) ListenClientMessages(quit chan bool) {
	for {
		// TODO: Increase buffer
		buffer := make([]byte, 10240)
		n, _, e := g.ClientConn.ReadFromUDP(buffer)
		utils.HandleError(e)

		msg := &utils.Message{}
		e = protobuf.Decode(buffer[:n], msg)
		utils.HandleError(e)
		if msg.Text != "" {
			fmt.Println("CLIENT MESSAGE", msg.Text)
			fmt.Printf("PEERS %v\n", fmt.Sprint(strings.Join(g.peers, ",")))
		}
		go g.ClientMessageHandler(*msg)
	}
	quit <- true
}

func (g *Gossiper) ListenPeerMessages(quit chan bool) {
	for {
		buffer := make([]byte, 10240)
		n, peerAddr, e := g.udpConn.ReadFromUDP(buffer)
		utils.HandleError(e)

		msg := &utils.GossipPacket{}
		e = protobuf.Decode(buffer[:n], msg)
		utils.HandleError(e)
		go g.peerMessageHandler(*msg, peerAddr.String())
	}
	quit <- true
}

func (g *Gossiper) ClientMessageHandler(msg utils.Message) {
	var wg sync.WaitGroup
	var gossipPacket utils.GossipPacket
	// TODO: Reorder ifs
	if g.simple {
		simpleMessage := utils.SimpleMessage{OriginalName: g.name, RelayPeerAddr: g.Address.String(), Contents: msg.Text}
		gossipPacket = utils.GossipPacket{Simple: &simpleMessage}
		for _, p := range g.peers {
			wg.Add(1)
			go func(p string) {
				// fmt.Printf("%d: Simple Message\n", time.Now().Second())
				g.sendToPeer(gossipPacket, p)
				wg.Done()
			}(p)
		}
	} else {
		if msg.FileName != "" {
			if msg.Request != "" && msg.Destination != "" {
				g.sendDataRequest(msg)
			} else {
				g.indexFile(msg)
			}
		} else if msg.Text != "" {
			if msg.Destination == "" {
				g.newRumorMongeringMessage(msg)
			} else {
				g.newPrivateMessage(msg)
			}
		} else {
			log.Printf("\n\nYOUR CLIENT MESSAGE:\nDestination: %s, FileName: %s, Request: %s, Text: %s\n\nWHAT'S THIS SUPPOSED TO BE?\n\n\n", msg.Destination, msg.FileName, msg.Request, msg.Text)
		}
	}
	wg.Wait()
}

func (g *Gossiper) peerMessageHandler(msg utils.GossipPacket, sender string) {
	g.addPeerToListIfApplicable(sender)
	if msg.Simple != nil {
		g.simpleMessageHandler(*msg.Simple)
	} else if msg.Rumor != nil {
		if msg.Rumor.Text == "" {
			log.Printf("%s: Got route rumor message from %s \n", g.name, sender)
		} else {
			log.Printf("%s: Got rumor message from %s \n", g.name, sender)
		}
		g.rumorMessageHandler(*msg.Rumor, sender)
	} else if msg.Status != nil {
		g.statusMessageHandler(*msg.Status, sender)
	} else if msg.Private != nil {
		log.Printf("%s: Got private message from %s \n", g.name, sender)
		g.privateMessageHandler(*msg.Private)
	} else if msg.DataRequest != nil {
		log.Printf("%s: Got data request from %s \n", g.name, sender)
		g.dataRequestHandler(*msg.DataRequest)
	} else if msg.DataReply != nil {
		log.Printf("%s: Got data reply from %s \n", g.name, sender)
		g.dataReplyHandler(*msg.DataReply)
	} else {
		log.Printf("\n\nWHAT'S THIS PEER MESSAGE SUPPOSED TO BE?.\n\n\n")
	}
}
