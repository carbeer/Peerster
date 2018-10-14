package main

import (
	"log"
	"testing"

	"github.com/carbeer/Peerster/gossiper"
	"github.com/carbeer/Peerster/utils"
)

func TestHasLessMessagesThan(t *testing.T) {
	// g1 := gossiper.NewGossiper("127.0.0.1", "A", 5000, 12345, []string{"127.0.0.1:5001"}, false)
	// g2 := gossiper.NewGossiper("127.0.0.2", "B", 5001, 12345, []string{"127.0.0.1:5002"}, false)
	g3 := gossiper.NewGossiper("127.0.0.3", "C", 5002, 12345, []string{"127.0.0.1:5001"}, false)

	ps := utils.PeerStatus{Identifier: "A", NextID: 1}
	ps2 := utils.PeerStatus{Identifier: "B", NextID: 1}
	sp := utils.StatusPacket{Want: []utils.PeerStatus{ps}}

	if !g3.HasLessMessagesThan(sp) {
		t.Errorf("Expected to succeed for HasLessThan")
	}

	g3.WantedMessages["A"] = "1"
	if g3.HasLessMessagesThan(sp) {
		t.Errorf("Expected to fail for HasLessThan")
	}

	g3.WantedMessages["A"] = "2"
	if g3.HasLessMessagesThan(sp) {
		t.Errorf("Expected to fail for HasLessThan")
	}

	sp.Want = append(sp.Want, ps2)
	if !g3.HasLessMessagesThan(sp) {
		t.Errorf("Expected to succeed for HasLessThan")
	}
}

func TestAdditionalMessages(t *testing.T) {
	g1 := gossiper.NewGossiper("127.0.0.4", "D", 5001, 12346, []string{"127.0.0.1:5002"}, false)

	ps := utils.PeerStatus{Identifier: "A", NextID: 1}
	// ps2 := utils.PeerStatus{Identifier: "B", NextID: 1}
	sp := utils.StatusPacket{Want: []utils.PeerStatus{ps}}

	rm := utils.RumorMessage{Origin: "A", ID: "0", Text: "Hi"}
	rm2 := utils.RumorMessage{Origin: "A", ID: "1", Text: "Hi2"}
	rm3 := utils.RumorMessage{Origin: "A", ID: "2", Text: "Hi3"}

	g1.ReceivedMessages["A"] = utils.RumorMessages{rm, rm2, rm3}
	g1.WantedMessages["A"] = "3"
	g1.WantedMessages["B"] = "1"

	boolVal, tmp := g1.AdditionalMessages(sp)
	log.Println(boolVal)
	log.Println(tmp)
	if !boolVal {
		t.Errorf("Doesn't identify new message")
	}
	log.Println(tmp)
	if len(tmp) == 0 || (!tmp[0].CompareRumorMessage(rm2) && !tmp[0].CompareRumorMessage(rm3)) {
		t.Errorf("Didn't get the desired message")
	}
}
