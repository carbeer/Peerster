package utils

import (
	"fmt"
	"strconv"
)

type SimpleMessage struct {
	OriginalName  string
	RelayPeerAddr string
	Contents      string
}

type RumorMessage struct {
	Origin string
	ID     string
	Text   string
}

type Message struct {
	Text string
}

type PeerStatus struct {
	Identifier string
	NextID     uint32
}

type StatusPacket struct {
	Want []PeerStatus
}

func (sp *StatusPacket) ToString() string {
	s := ""
	for status := range sp.Want {
		s = s + fmt.Sprintf("(%s: %v) ", sp.Want[status].Identifier, sp.Want[status].NextID)
	}
	return s
}

type GossipPacket struct {
	Simple *SimpleMessage
	Rumor  *RumorMessage
	Status *StatusPacket
}

// Make RumorMessages sortable according to ID
type RumorMessages []RumorMessage

func (rm RumorMessages) Len() int {
	return len(rm)
}

func (rm RumorMessages) Less(i, j int) bool {
	a, _ := strconv.Atoi(rm[i].ID)
	b, _ := strconv.Atoi(rm[j].ID)
	return a < b
}

func (rm RumorMessages) Swap(i, j int) {
	rm[i], rm[j] = rm[j], rm[i]
}

func (rm RumorMessage) GetIdentifier() string {
	return fmt.Sprintf("%v%v", rm.Origin, rm.ID)
}
