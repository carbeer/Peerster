package gossiper

import (
	"fmt"
	"time"

	"github.com/carbeer/Peerster/utils"
)

func (g *Gossiper) MineBlocks() {
	fmt.Println("Starting to mine")
	lastBlock := g.getLastBlock().Block
	block := utils.Block{Transactions: g.getPendingTransactions(), PrevHash: [32]byte{0}, Nonce: utils.GetRandomNonce()}

	for {
		for len(g.getPendingTransactions()) == 0 {
			fmt.Println("Pause mining. Waiting for new transactions")
			// Block until new transactions to be mined are available
			<-g.miner
		}
		fmt.Println("Resume mining.")
		// Intialize new values

		startTime := time.Now()
		for !utils.ValidateBlockHash(block) {
			select {
			case <-g.miner:
				lastBlock = g.getLastBlock().Block
				fmt.Println("Updating block. Last block hash ", utils.FixedStringHash(lastBlock.Hash()))
				block = utils.Block{Transactions: g.getPendingTransactions(), PrevHash: lastBlock.Hash(), Nonce: utils.GetRandomNonce()}
				break
			default:
				block.Nonce = utils.GetRandomNonce()
			}
		}
		// Sleep twice the mining time
		sleepingTime := 2 * time.Since(startTime)
		fmt.Printf("Found a block. Sleeping %f seconds.\n", sleepingTime.Seconds())
		<-time.After(sleepingTime)
		g.proposeBlock(utils.BlockPublish{Block: block, HopLimit: utils.GetBlockPublishHopLimit()})

		lastBlock = g.getLastBlock().Block
		block = utils.Block{Transactions: g.getPendingTransactions(), PrevHash: lastBlock.Hash(), Nonce: utils.GetRandomNonce()}
	}
}

func (g *Gossiper) proposeBlock(msg utils.BlockPublish) {
	fmt.Printf("FOUND-BLOCK [%s]\n", utils.FixedStringHash(msg.Block.Hash()))
	g.broadcastMessage(utils.GossipPacket{BlockPublish: &msg}, "")
	g.updateBlockChain(msg.Block)
}

func (g *Gossiper) blockPublishHandler(msg utils.BlockPublish, sender string) {
	if g.ValidateBlock(msg.Block) {
		return
	}
	// TODO: Ensure that those transactions do not claim anything that was already claimed (or duplicate transactions)
	g.updateBlockChain(msg.Block)

	msg.HopLimit--
	if msg.HopLimit > 0 {
		g.broadcastMessage(utils.GossipPacket{BlockPublish: &msg}, sender)
	}
}

func (g *Gossiper) updateBlockChain(key utils.Block) {
	g.chainLock.Lock()
	defer g.chainLock.Unlock()
	prevBlock := g.blockHistory[key.PrevHash]

	if prevBlock.Counter == 0 && key.PrevHash != [32]byte{0} {
		fmt.Printf("Don't know the parent block, dropping block %s.\n", utils.FixedStringHash(key.Hash()))
		return
	} else if g.blockHistory[key.Hash()].Counter != 0 {
		fmt.Printf("Not a new block: %s\n", utils.FixedStringHash(key.Hash()))
		return
	}
	g.blockHistory[key.Hash()] = utils.BlockWrapper{Block: key, Counter: prevBlock.Counter + 1}

	if key.PrevHash == g.lastBlock.Block.Hash() || key.PrevHash == [32]byte{0} && g.lastBlock.Counter == 0 {
		fmt.Println("This is the longest chain")
		g.removePendingTransactions(key, true)
	} else if prevBlock.Counter+1 <= g.lastBlock.Counter {
		fmt.Printf("FORK-SHORTER %s\n", utils.FixedStringHash(key.Hash()))
		fmt.Printf("Last block hash: %s but this block is pointing to %s\n", utils.FixedStringHash(g.lastBlock.Block.Hash()), utils.FixedStringHash(key.PrevHash))
		return
	} else {
		rewind := g.resolveFork(key, true)
		fmt.Printf("FORK-LONGER rewind %d blocks\n", rewind)
	}

	g.lastBlock = utils.BlockWrapper{Block: key, Counter: prevBlock.Counter + 1}
	currBlock := g.lastBlock.Block
	// notify miner
	g.miner <- true
	fmt.Printf("CHAIN [%s]", utils.FixedStringHash(currBlock.Hash()))
	for currBlock.PrevHash != [32]byte{0} {
		fmt.Printf(" [%s]", utils.FixedStringHash(currBlock.PrevHash))
		currBlock = g.blockHistory[currBlock.PrevHash].Block
	}
	fmt.Print("\n")
}

func (g *Gossiper) resolveFork(msg utils.Block, lock bool) int {
	if !lock {
		g.chainLock.Lock()
		defer g.chainLock.Unlock()
	}
	forkBlock := g.blockHistory[msg.PrevHash].Block // has to be equal to the highest counter of the prviously longest chain
	chainBlock := g.lastBlock.Block
	count := 0
	initial := len(g.pendingTransactions)
	for ; ; count++ {
		// Add all transactions that are now in orphaned fork back to the pending pool
	Loop:
		for _, v := range chainBlock.Transactions {
			for _, contained := range chainBlock.Transactions {
				if contained.File.Name == v.File.Name {
					continue Loop
				}
			}
			g.pendingTransactions = append(g.pendingTransactions, v)
		}
		// Remove all transactions from pending that are part of the new chain
		g.removePendingTransactions(forkBlock, true)

		if chainBlock.PrevHash == forkBlock.PrevHash {
			break
		}
		chainBlock = g.blockHistory[chainBlock.PrevHash].Block
		forkBlock = g.blockHistory[forkBlock.PrevHash].Block
	}
	fmt.Printf("We have a difference of %d pending transactions after the resolution", initial-len(g.pendingTransactions))
	return count + 1
}

func (g *Gossiper) removePendingTransactions(key utils.Block, locked bool) {
	if !locked {
		g.chainLock.Lock()
		defer g.chainLock.Unlock()
	}
	newPending := []utils.TxPublish{}
Loop:
	for _, pending := range g.pendingTransactions {
		for _, mined := range key.Transactions {
			if pending.File.Name == mined.File.Name {
				fmt.Printf("Transaction %+v is part of the block already\n", pending)
				continue Loop
			}
		}
		fmt.Printf("Transaction %+v is still pending\n", pending)
		newPending = append(newPending, pending)
	}
	g.pendingTransactions = newPending
}

func (g *Gossiper) ValidateBlock(block utils.Block) bool {
	g.chainLock.Lock()
	defer g.chainLock.Unlock()

	if !utils.ValidateBlockHash(block) {
		return false
	}

	for _, v := range block.Transactions {
		// Iterate through all blocks in its chain
		currentBlock := g.blockHistory[block.PrevHash].Block
		for currentBlock.PrevHash != [32]byte{} {
			for _, mined := range currentBlock.Transactions {
				if mined.File.Name == v.File.Name {
					return false
				}
			}
			currentBlock = g.blockHistory[currentBlock.PrevHash].Block
		}
	}
	return true
}
