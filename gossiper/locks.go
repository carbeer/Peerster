package gossiper

import "github.com/carbeer/Peerster/utils"

func (g *Gossiper) getReceivedMessages(key string) utils.RumorMessages {
	g.receivedMessagesLock.RLock()
	val := g.ReceivedMessages[key]
	g.receivedMessagesLock.RUnlock()
	return val
}

func (g *Gossiper) appendReceivedMessages(key string, value utils.RumorMessage) {
	g.receivedMessagesLock.Lock()
	g.ReceivedMessages[key] = append(g.ReceivedMessages[key], value)
	g.receivedMessagesLock.Unlock()
}

func (g *Gossiper) getPrivateMessages(key string) []utils.PrivateMessage {
	g.privateMessagesLock.RLock()
	val := g.PrivateMessages[key]
	g.privateMessagesLock.RUnlock()
	return val
}

func (g *Gossiper) appendPrivateMessages(key string, value utils.PrivateMessage) {
	g.privateMessagesLock.Lock()
	g.PrivateMessages[key] = append(g.PrivateMessages[key], value)
	g.privateMessagesLock.Unlock()
}

func (g *Gossiper) getWantedMessages(key string) uint32 {
	g.wantedMessagesLock.RLock()
	val := g.WantedMessages[key]
	g.wantedMessagesLock.RUnlock()
	return val
}

func (g *Gossiper) setWantedMessages(key string, value uint32) {
	g.wantedMessagesLock.Lock()
	g.WantedMessages[key] = value
	g.wantedMessagesLock.Unlock()
}

func (g *Gossiper) getRumorMongeringChannel(key string) chan utils.StatusPacket {
	g.rumorMongeringChannelLock.RLock()
	val := g.rumorMongeringChannel[key]
	g.rumorMongeringChannelLock.RUnlock()
	return val
}

func (g *Gossiper) setRumorMongeringChannel(key string, value chan utils.StatusPacket) {
	g.rumorMongeringChannelLock.Lock()
	g.rumorMongeringChannel[key] = value
	g.rumorMongeringChannelLock.Unlock()
}

func (g *Gossiper) sendToRumorMongeringChannel(key string, value utils.StatusPacket) {
	g.rumorMongeringChannelLock.Lock()
	g.rumorMongeringChannel[key] <- value
	g.rumorMongeringChannelLock.Unlock()
}

func (g *Gossiper) getNextHop(key string) string {
	g.nextHopLock.RLock()
	val := g.nextHop[key]
	g.nextHopLock.RUnlock()
	return val
}

func (g *Gossiper) setNextHop(key string, value string) {
	g.nextHopLock.Lock()
	g.nextHop[key] = value
	g.nextHopLock.Unlock()
}
