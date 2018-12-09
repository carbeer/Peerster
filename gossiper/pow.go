package gossiper

import (
	"fmt"
	"time"

	"github.com/carbeer/Peerster/utils"
)

func (g *Gossiper) MineBlocks() {
	fmt.Println("Starting to mine")
	var block utils.Block
	var lastBlock utils.Block
	var sleepingTime time.Duration
	firstMinedBlock := true

Loop:
	for {

		for len(g.getPendingTransactions()) == 0 {
			fmt.Println("Pause mining. Waiting for new transactions")
			// Block until new transactions to be mined are available
			<-g.miner
			fmt.Println("Resume mining.")
		}

		startTime := time.Now()
		lastBlock = g.getLastBlock().Block

		if firstMinedBlock {
			block = utils.Block{Transactions: g.getPendingTransactions(), PrevHash: [32]byte{0}, Nonce: utils.GetRandomNonce()}
		} else {
			block = utils.Block{Transactions: g.getPendingTransactions(), PrevHash: lastBlock.Hash(), Nonce: utils.GetRandomNonce()}
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
	fmt.Printf("Broadcasting %s now\n", utils.FixedStringHash(msg.Block.Hash()))
	g.broadcastMessage(utils.GossipPacket{BlockPublish: &msg}, "")
}

func (g *Gossiper) updateBlockChain(key utils.Block, locked bool) {
	fmt.Printf("The block %s to be appended contains the following transactions: %+v\n", utils.FixedStringHash(key.Hash()), key.Transactions)

	if !locked {
		g.chainLock.Lock()
		defer g.chainLock.Unlock()
	}
	prevBlock := g.blockHistory[key.PrevHash]

	if g.blockHistory[key.Hash()].Counter != 0 {
		fmt.Printf("Not a new block: %s\n", utils.FixedStringHash(key.Hash()))
		return
	}
	g.blockHistory[key.Hash()] = utils.BlockWrapper{Block: key, Counter: prevBlock.Counter + 1}

	if key.PrevHash == g.lastBlock.Block.Hash() || key.PrevHash == [32]byte{0} && g.lastBlock.Counter == 0 {
		fmt.Println("This is the longest chain")
		g.removePendingTransactions(key, true)
	} else if prevBlock.Counter+1 <= g.lastBlock.Counter {
		g.checkNewFork(key, true)
		return
	} else {
		g.resolveFork(key, true)

	}

	g.lastBlock = utils.BlockWrapper{Block: key, Counter: prevBlock.Counter + 1}
	// notify miner
	g.miner <- true
	g.printChain()

	if g.detachedBlocks[key.Hash()].PrevHash == key.Hash() {
		fmt.Printf("Found detached block that can be used now.")
		detBlock := g.detachedBlocks[key.Hash()]
		delete(g.detachedBlocks, key.Hash())
		g.blockPublishHandler(utils.BlockPublish{Block: detBlock, HopLimit: 0}, "")
	}
}

func (g *Gossiper) checkNewFork(msg utils.Block, locked bool) {
	parent := g.blockHistory[msg.PrevHash]
	currBlock := g.lastBlock

	for currBlock.Counter > parent.Counter {
		currBlock = g.blockHistory[currBlock.Block.PrevHash]
	}

	if currBlock.Block.Hash() == parent.Block.Hash() {
		fmt.Printf("FORK-SHORTER %s\n", utils.FixedStringHash(currBlock.Hash()))
	}
	fmt.Printf("Highest known count: %d This block has a counter of %d\n", g.lastBlock.Counter, parent.Counter+1)
}

func (g *Gossiper) resolveFork(msg utils.Block, locked bool) {
	if !locked {
		g.chainLock.Lock()
		defer g.chainLock.Unlock()
	}
	forkBlock := g.blockHistory[msg.PrevHash].Block // has to be equal to the highest counter of the prviously longest chain
	chainBlock := g.lastBlock.Block
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
			g.pendingTransactions = append(g.pendingTransactions, v)
		}
		// Remove all transactions from pending that are part of the new chain
		g.removePendingTransactions(msg, true)

		if chainBlock.PrevHash == forkBlock.PrevHash {
			break
		}
		chainBlock = g.blockHistory[chainBlock.PrevHash].Block
		forkBlock = g.blockHistory[forkBlock.PrevHash].Block
	}
	fmt.Printf("FORK-LONGER rewind %d blocks\n", count+1)
}

func (g *Gossiper) removePendingTransactions(key utils.Block, locked bool) {
	if !locked {
		g.chainLock.Lock()
		defer g.chainLock.Unlock()
	}
	newPending := []utils.TxPublish{}
	fmt.Printf("Block contains transactions %+v\n", key.Transactions)
	fmt.Printf("Transactions still pending: %+v\n", g.pendingTransactions)
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
	fmt.Printf("New pending transactions: %+v\n", g.pendingTransactions)
}

func (g *Gossiper) ValidateHistory(block utils.Block, locked bool) bool {
	if !locked {
		g.chainLock.Lock()
		defer g.chainLock.Unlock()
	}

	// Check whether the transactions within the block are unique within the chain
	for _, v := range block.Transactions {
		currentBlock := g.blockHistory[block.PrevHash].Block
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
			currentBlock = g.blockHistory[currentBlock.PrevHash].Block
		}
	}
	return true
}

func (g *Gossiper) printChain() {
	currBlock := g.lastBlock.Block
	fmt.Print("CHAIN")
	for {
		fmt.Printf(" %s", utils.FixedStringHash(currBlock.Hash()))
		fmt.Printf(":%s", utils.FixedStringHash(currBlock.PrevHash))
		for ix, trans := range currBlock.Transactions {
			if ix == 0 {
				fmt.Printf(":%s", trans.File.Name)
			} else {
				fmt.Printf(",%s", trans.File.Name)
			}
		}
		if currBlock.PrevHash == [32]byte{0} {
			break
		}
		currBlock = g.blockHistory[currBlock.PrevHash].Block
	}
	fmt.Print("\n")
}
