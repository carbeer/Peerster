package gossiper

import (
	"fmt"
	"time"

	"github.com/carbeer/Peerster/utils"
)

func (g *Gossiper) MineBlocks() {
	fmt.Println("Starting to mine")
	var block utils.Block
	var LastBlock utils.Block
	var sleepingTime time.Duration
	firstMinedBlock := true

Loop:
	for {
		// Block until new transactions to be mined are available
		for len(g.getPendingTransactions()) == 0 {
			<-g.miner
		}

		startTime := time.Now()
		LastBlock = g.getLastBlock().Block

		if firstMinedBlock {
			block = utils.Block{Transactions: g.getPendingTransactions(), PrevHash: [32]byte{0}, Nonce: utils.GetRandomNonce()}
		} else {
			block = utils.Block{Transactions: g.getPendingTransactions(), PrevHash: LastBlock.Hash(), Nonce: utils.GetRandomNonce()}
		}

		for !utils.ValidateBlockHash(block) {
			select {
			case <-g.miner:
				continue Loop
			default:
				block.Nonce = utils.NextNonce(block.Nonce)
			}
		}
		fmt.Printf("FOUND-BLOCK %s\n", utils.FixedStringHash(block.Hash()))

		// Sleep twice the mining time
		if firstMinedBlock {
			sleepingTime = 5 * time.Second
		} else {
			sleepingTime = 2 * time.Since(startTime)
		}
		go g.proposeBlock(utils.BlockPublish{Block: block, HopLimit: utils.BLOCK_PUBLISH_HOP_LIMIT}, sleepingTime)
		g.updateBlockChain(block, false)
		firstMinedBlock = false
	}
}

func (g *Gossiper) proposeBlock(msg utils.BlockPublish, sleep time.Duration) {
	fmt.Printf("Sleeping %f seconds before broadcasting new block.\n", sleep.Seconds())
	<-time.After(sleep)
	g.broadcastMessage(utils.GossipPacket{BlockPublish: &msg}, "")
}

func (g *Gossiper) updateBlockChain(key utils.Block, locked bool) {
	if !locked {
		g.chainLock.Lock()
		defer g.chainLock.Unlock()
	}
	prevBlock := g.BlockHistory[key.PrevHash]

	if g.BlockHistory[key.Hash()].Counter != 0 {
		return
	}
	g.BlockHistory[key.Hash()] = utils.BlockWrapper{Block: key, Counter: prevBlock.Counter + 1}

	if key.PrevHash == g.LastBlock.Block.Hash() || key.PrevHash == [32]byte{0} && g.LastBlock.Counter == 0 {
		g.removePendingTransactions(key, true)
	} else if prevBlock.Counter+1 <= g.LastBlock.Counter {
		g.checkNewFork(key, true)
		return
	} else {
		g.resolveFork(key, true)

	}

	g.LastBlock = utils.BlockWrapper{Block: key, Counter: prevBlock.Counter + 1}
	// notify miner
	g.miner <- true
	g.printChain()

	if g.DetachedBlocks[key.Hash()].PrevHash == key.Hash() {
		detBlock := g.DetachedBlocks[key.Hash()]
		delete(g.DetachedBlocks, key.Hash())
		g.blockPublishHandler(utils.BlockPublish{Block: detBlock, HopLimit: 0}, "")
	}
}

func (g *Gossiper) checkNewFork(msg utils.Block, locked bool) {
	parent := g.BlockHistory[msg.PrevHash]
	currBlock := g.LastBlock

	for currBlock.Counter > parent.Counter {
		currBlock = g.BlockHistory[currBlock.Block.PrevHash]
	}

	if currBlock.Block.Hash() == parent.Block.Hash() {
		fmt.Printf("FORK-SHORTER %s\n", utils.FixedStringHash(currBlock.Hash()))
	}
}

func (g *Gossiper) resolveFork(msg utils.Block, locked bool) {
	if !locked {
		g.chainLock.Lock()
		defer g.chainLock.Unlock()
	}
	forkBlock := g.BlockHistory[msg.PrevHash].Block // has to be equal to the highest counter of the prviously longest chain
	chainBlock := g.LastBlock.Block
	count := 0

	for ; ; count++ {
		// Add all transactions that are now in orphaned fork back to the pending pool
	Loop:
		for _, v := range chainBlock.Transactions {
			for _, contained := range chainBlock.Transactions {
				if contained.File.Name == v.File.Name {
					continue Loop
				}
			}
			g.PendingTransactions = append(g.PendingTransactions, v)
		}
		// Remove all transactions from pending that are part of the new chain
		g.removePendingTransactions(msg, true)

		if chainBlock.PrevHash == forkBlock.PrevHash {
			break
		}
		chainBlock = g.BlockHistory[chainBlock.PrevHash].Block
		forkBlock = g.BlockHistory[forkBlock.PrevHash].Block
	}
	fmt.Printf("FORK-LONGER rewind %d blocks\n", count+1)
}

func (g *Gossiper) removePendingTransactions(key utils.Block, locked bool) {
	if !locked {
		g.chainLock.Lock()
		defer g.chainLock.Unlock()
	}
	newPending := []utils.TxPublish{}
Loop:
	for _, pending := range g.PendingTransactions {
		for _, mined := range key.Transactions {
			if pending.File.Name == mined.File.Name {
				continue Loop
			}
		}
		newPending = append(newPending, pending)
	}
	g.PendingTransactions = newPending
}

func (g *Gossiper) ValidateHistory(block utils.Block, locked bool) bool {
	if !locked {
		g.chainLock.Lock()
		defer g.chainLock.Unlock()
	}

	// Check whether the transactions within the block are unique within the chain
	for _, v := range block.Transactions {
		currentBlock := g.BlockHistory[block.PrevHash].Block
		for {
			for _, mined := range currentBlock.Transactions {
				if mined.File.Name == v.File.Name {
					fmt.Printf("File %s is already present in previous block %s. Block %s is therefore corrupted\n", mined.File.Name, utils.FixedStringHash(currentBlock.Hash()), utils.FixedStringHash(block.Hash()))
					return false
				}
			}
			if currentBlock.PrevHash == [32]byte{0} {
				break
			}
			currentBlock = g.BlockHistory[currentBlock.PrevHash].Block
		}
	}
	return true
}

func (g *Gossiper) printChain() {
	currBlock := g.LastBlock.Block
	toPrint := "CHAIN"
	for {
		toPrint += fmt.Sprintf(" %s", utils.FixedStringHash(currBlock.Hash()))
		toPrint += fmt.Sprintf(":%s", utils.FixedStringHash(currBlock.PrevHash))
		for ix, trans := range currBlock.Transactions {
			if ix == 0 {
				toPrint += fmt.Sprintf(":%s", trans.File.Name)
			} else {
				toPrint += fmt.Sprintf(",%s", trans.File.Name)
			}
		}
		if currBlock.PrevHash == [32]byte{0} {
			break
		}
		currBlock = g.BlockHistory[currBlock.PrevHash].Block
	}
	toPrint += "\n"
	fmt.Print(toPrint)
}
