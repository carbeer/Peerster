package main

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/carbeer/Peerster/gossiper"
	"github.com/carbeer/Peerster/utils"
)

func TestHasLessMessagesThan(t *testing.T) {
	g := gossiper.NewGossiper("127.0.0.3", "C", 5002, 12345, []string{"127.0.0.1:5001"}, false)

	ps := utils.PeerStatus{Identifier: "A", NextID: 2}
	ps2 := utils.PeerStatus{Identifier: "B", NextID: 2}
	sp := utils.StatusPacket{Want: []utils.PeerStatus{ps}}

	if !g.HasLessMessagesThan(sp) {
		t.Errorf("Expected to succeed for HasLessThan")
	}

	sp.Want = append(sp.Want, ps2)
	if !g.HasLessMessagesThan(sp) {
		t.Errorf("Expected to succeed for HasLessThan")
	}
}

func TestAdditionalMessages(t *testing.T) {
	g := gossiper.NewGossiper("127.0.0.1", "D", 5001, 12346, []string{"127.0.0.1:5002"}, false)

	ps := utils.PeerStatus{Identifier: "A", NextID: 1}
	sp := utils.StatusPacket{Want: []utils.PeerStatus{ps}}

	rm := utils.RumorMessage{Origin: "A", ID: 0, Text: "Hi"}
	rm2 := utils.RumorMessage{Origin: "A", ID: 1, Text: "Hi2"}
	rm3 := utils.RumorMessage{Origin: "A", ID: 2, Text: "Hi3"}

	g.ReceivedMessages["A"] = []utils.RumorMessage{rm, rm2, rm3}

	boolVal, tmp := g.AdditionalMessages(sp)
	if !boolVal {
		t.Errorf("Doesn't identify new message")
	}
	log.Println(tmp)
	if !boolVal {
		t.Errorf("Didn't get the desired message")
	}
}

func TestTextEncryption(t *testing.T) {
	g := gossiper.NewGossiper("127.0.0.1", "", 5001, 12346, []string{"127.0.0.1:5002"}, false)
	g2 := gossiper.NewGossiper("127.0.0.1", "", 5002, 12347, []string{"127.0.0.1:5001"}, false)
	s := "Hi"
	ctext := gossiper.RSAEncryptText(g.Name, s)
	if ctext == s {
		t.Errorf(fmt.Sprint("The ciphertext is the same as the original text:", s, ctext))
	}
	s_roundtrip := g.RSADecryptText(ctext)

	if s != s_roundtrip {
		t.Errorf(fmt.Sprint("Original and roundtrip message are not the same:", s, s_roundtrip))
	}

	s_roundtrip = g2.RSADecryptText(ctext)

	if s == s_roundtrip {
		t.Errorf("Another peer is not supposed to decrypt a message for someone else.")
	}
}

func TestReplicationReferencing(t *testing.T) {
	g := gossiper.NewGossiper("127.0.0.1", "", 5001, 12346, []string{"127.0.0.1:5002"}, false)
	m := utils.Message{FileName: "file1.txt", Replications: 5}
	g2 := gossiper.NewGossiper("127.0.0.1", "", 5002, 12347, []string{"127.0.0.1:5001"}, false)
	m2 := utils.Message{FileName: "file2.txt", Replications: 2}
	go g.PrivateFileIndexing(m)
	go g2.PrivateFileIndexing(m2)
	<-time.After(5 * time.Second)

	for key, val := range g.Replications {
		if key != val.Metafilehash {
			t.Errorf(fmt.Sprintf("Got mismatching metafilehashes: %s vs %s", key, val.Metafilehash))
		}
		g.AssignReplica(key, val.NodeID, "yoooo")
	}

	for key, _ := range g.PrivFiles {
		log.Printf("%+v\n", g.PrivFiles[key])
		for i, v := range g.PrivFiles[key].Replications {
			log.Printf("%p vs %p", g.Replications[v.Metafilehash], &g.PrivFiles[key].Replications[i])
			if g.Replications[v.Metafilehash] != &g.PrivFiles[key].Replications[i] {
				t.Errorf(fmt.Sprintf("Got mismatching pointers: %p vs %p", g.Replications[v.Metafilehash], &g.PrivFiles[key].Replications[i]))
			}
		}
	}
}

func TestSignatures(t *testing.T) {
	g := gossiper.NewGossiper("127.0.0.1", "", 5001, 12346, []string{"127.0.0.1:5002"}, false)
	pm := utils.PrivateMessage{Origin: g.Name, ID: 0, EncryptedText: "blablabla", Destination: g.Name, HopLimit: utils.HOPLIMIT_CONSTANT}
	pm.EncryptedText = gossiper.RSAEncryptText(g.Name, pm.EncryptedText)
	pm = g.RSASignPM(pm)
	if !g.RSAVerifyPMSignature(pm) {
		t.Errorf("Signature not valid")
	}
}
