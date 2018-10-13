package utils

type SimpleMessage struct {
	OriginalName  string
	RelayPeerAddr string
	Contents      string
}

type Message struct {
	Text string
}

type GossipPacket struct {
	Simple *SimpleMessage
}
