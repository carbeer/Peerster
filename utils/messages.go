package utils

import (
	"fmt"
)

type SimpleMessage struct {
	OriginalName  string
	RelayPeerAddr string
	Contents      string
}

type RumorMessage struct {
	Origin string
	ID     uint32
	Text   string
}

type PrivateMessage struct {
	Origin      string
	ID          uint32
	Text        string
	Destination string
	HopLimit    uint32
}

type Message struct {
	Text        string
	Destination string
}

type PeerStatus struct {
	// Equals the Origin in RumorMessage
	Identifier string
	NextID     uint32
}

type StatusPacket struct {
	Want []PeerStatus
}

func (sp *StatusPacket) ToString() string {
	s := ""
	for status := range sp.Want {
		s = fmt.Sprintf("%s peer %s nextID %v", s, sp.Want[status].Identifier, sp.Want[status].NextID)
	}
	return s
}

type GossipPacket struct {
	Simple  *SimpleMessage
	Rumor   *RumorMessage
	Status  *StatusPacket
	Private *PrivateMessage
}

// Make RumorMessages sortable according to ID
type RumorMessages []RumorMessage

func (rm RumorMessages) Len() int {
	return len(rm)
}

func (rm RumorMessages) GetById(id int) RumorMessage {
	return rm[id-1]
}

func (rm RumorMessages) Less(i, j int) bool {
	return rm[i].ID < rm[j].ID
}

func (rm RumorMessages) Swap(i, j int) {
	rm[i], rm[j] = rm[j], rm[i]
}

func (rm RumorMessage) GetIdentifier() string {
	return fmt.Sprintf("%v%v", rm.Origin, rm.ID)
}

func (rm1 RumorMessage) CompareRumorMessage(rm2 RumorMessage) bool {
	if rm1.Origin != rm2.Origin || rm1.ID != rm2.ID || rm1.Text != rm2.Text {
		return false
	}
	return true
}
