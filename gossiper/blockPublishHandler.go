package gossiper

import (
	"fmt"

	"github.com/carbeer/Peerster/utils"
)

func (g *Gossiper) blockPublishHandler(msg utils.BlockPublish, sender string) {
	g.chainLock.Lock()
	defer g.chainLock.Unlock()

	if g.blockHistory[msg.Block.PrevHash].Counter == 0 && msg.Block.PrevHash != [32]byte{0} {
		if !g.receivedBlock && sender != g.name {
			fmt.Printf("This is the first block we received but we don't have its parent\n")
		} else {
			fmt.Printf("Storing the block %s for later usage as parent %s is unknown\n", utils.FixedStringHash(msg.Block.Hash()), utils.FixedStringHash(msg.Block.PrevHash))
			g.detachedBlocks[msg.Block.PrevHash] = msg.Block
			return
		}
	}
	if !utils.ValidateBlockHash(msg.Block) {
		fmt.Printf("Dropping block %s as its hash is corrupted: %+v\n", utils.FixedStringHash(msg.Block.Hash()), msg.Block)
		return
	}
	if !g.ValidateHistory(msg.Block, true) {
		fmt.Printf("Dropping block %s as its history is corrupted: %+v\n", utils.FixedStringHash(msg.Block.Hash()), msg.Block)
		return
	}
	g.updateBlockChain(msg.Block, true)
	msg.HopLimit--
	if msg.HopLimit > 0 {
		g.broadcastMessage(utils.GossipPacket{BlockPublish: &msg}, sender)
	}
	if !g.receivedBlock && sender != g.name {
		g.receivedBlock = true
	}
}
