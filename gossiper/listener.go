package gossiper

import (
	"fmt"
	"strings"
	"sync"

	"github.com/carbeer/Peerster/utils"
	"github.com/dedis/protobuf"
)

func (g *Gossiper) ListenClientMessages() {
	for {
		buffer := make([]byte, 10240)
		n, _, e := g.clientConn.ReadFromUDP(buffer)
		utils.HandleError(e)

		msg := &utils.Message{}
		e = protobuf.Decode(buffer[:n], msg)
		utils.HandleError(e)
		if msg.Text != "" {
			fmt.Println("CLIENT MESSAGE", msg.Text)
			fmt.Printf("PEERS %v\n", fmt.Sprint(strings.Join(g.Peers, ",")))
		}
		go g.ClientMessageHandler(*msg)
	}
}

func (g *Gossiper) ListenPeerMessages() {
	for {
		buffer := make([]byte, 10240)
		n, peerAddr, e := g.udpConn.ReadFromUDP(buffer)
		utils.HandleError(e)
		msg := &utils.GossipPacket{}
		e = protobuf.Decode(buffer[:n], msg)
		utils.HandleError(e)
		go g.peerMessageHandler(*msg, peerAddr.String())
	}
}

func (g *Gossiper) ClientMessageHandler(msg utils.Message) error {
	var err error
	var wg sync.WaitGroup
	var gossipPacket utils.GossipPacket
	if g.Simple {
		simpleMessage := utils.SimpleMessage{OriginalName: g.Name, RelayPeerAddr: g.Address.String(), Contents: msg.Text}
		gossipPacket = utils.GossipPacket{Simple: &simpleMessage}
		for _, p := range g.Peers {
			wg.Add(1)
			go func(p string) {
				g.sendToPeer(gossipPacket, p)
				wg.Done()
			}(p)
		}
	} else {
		if msg.FileName != "" {
			if msg.Request != "" {
				if msg.Encrypted {
					err = g.DownloadPrivateFile(msg)
				} else {
					err = g.DownloadFile(msg)
				}
			} else if msg.Replications != 0 {
				g.PrivateFileIndexing(msg)
			} else if msg.Encrypted {
				g.LoadPrivateFiles(msg.FileName)
			} else {
				g.indexFile(msg)
			}
		} else if msg.Text != "" {
			if msg.Destination == "" {
				g.newRumorMongeringMessage(msg)
			} else {
				if msg.Encrypted {
					g.newEncryptedPrivateMessage(msg)
				} else {
					g.newPrivateMessage(msg)
				}
			}
		} else if msg.Peer != "" {
			g.addPeerToListIfApplicable(msg.Peer)
		} else if len(msg.Keywords) > 0 {
			g.newSearchRequest(msg)
		} else {
			fmt.Printf("\n\nYOUR CLIENT MESSAGE:\n%+v\nWHAT'S THIS SUPPOSED TO BE? NOT PROPAGATING THIS.\n\n\n", msg)
		}
	}
	wg.Wait()
	return err
}

func (g *Gossiper) peerMessageHandler(msg utils.GossipPacket, sender string) {
	g.addPeerToListIfApplicable(sender)
	if msg.Simple != nil {
		g.simpleMessageHandler(*msg.Simple)
	} else if msg.Rumor != nil {
		if msg.Rumor.Text == "" {
			fmt.Printf("Got route rumor message from %s \n", sender)
		} else {
			fmt.Printf("Got rumor message from %s \n", sender)
		}
		g.rumorMessageHandler(*msg.Rumor, sender)
	} else if msg.Status != nil {
		g.statusMessageHandler(*msg.Status, sender)
	} else if msg.Private != nil {
		fmt.Printf("Got private message from %s \n", sender)
		g.privateMessageHandler(*msg.Private)
	} else if msg.DataRequest != nil {
		fmt.Printf("Got data request from %s \n", sender)
		g.dataRequestHandler(*msg.DataRequest, sender)
	} else if msg.DataReply != nil {
		fmt.Printf("Got data reply from %s \n", sender)
		g.dataReplyHandler(*msg.DataReply)
	} else if msg.SearchReply != nil {
		fmt.Printf("Got search reply from %s\n", sender)
		g.searchReplyHandler(*msg.SearchReply)
	} else if msg.SearchRequest != nil {
		fmt.Printf("Got search request from %s \n", sender)
		g.searchRequestHandler(*msg.SearchRequest, false)
	} else if msg.TxPublish != nil {
		fmt.Printf("Got tx publish from %s \n", sender)
		g.txPublishHandler(*msg.TxPublish, sender)
	} else if msg.BlockPublish != nil {
		fmt.Printf("Got block publish from %s \n", sender)
		g.blockPublishHandler(*msg.BlockPublish, sender)
	} else if msg.FileExchangeRequest != nil {
		fmt.Printf("Got file exchange request from %s\n", sender)
		g.fileExchangeRequestHandler(*msg.FileExchangeRequest, sender)
	} else if msg.Challenge != nil {
		g.challengeHandler(*msg.Challenge, sender)
	} else {
		fmt.Printf("\n\nYOUR PEER MESSAGE:\n%+v\nWHAT'S THIS SUPPOSED TO BE? NOT PROPAGATING THIS.\n\n\n", msg)
	}
}
