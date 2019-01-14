package gossiper

import (
	"crypto"
	"encoding/hex"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/carbeer/Peerster/utils"
)

// Index a file privately: Do not publish it to the blockchain but broadcast FileExchangeRequests instead
func (g *Gossiper) PrivateFileIndexing(msg utils.Message) {
	if msg.Replications == -1 {
		msg.Replications = 0
	}
	fmt.Printf("REQUESTING PRIVATE INDEXING filename %s with %d Replications\n", msg.FileName, msg.Replications)
	var hashFunc = crypto.SHA256.New()
	file, e := os.Open(filepath.Join(".", utils.SHARED_FOLDER, msg.FileName))
	utils.HandleError(e)
	fileInfo, e := file.Stat()
	utils.HandleError(e)

	fileSize := fileInfo.Size()
	noChunks := utils.CeilIntDiv(int(fileSize), utils.CHUNK_SIZE)

	// First replication is without encryption --> original file which will be stored locally
	r := make([]utils.Replica, msg.Replications+1)

	// Create replications
	for j := 0; j <= msg.Replications; j++ {
		_, e := file.Seek(0, 0)
		utils.HandleError(e)
		var chunkHashed []byte
		key := utils.GenerateRandomByteArr(utils.AES_KEY_LENGTH)

		for i := 0; i < noChunks; i++ {
			hashFunc.Reset()
			var chunk []byte
			if j == 0 {
				chunk = utils.GetNextDataChunk(file, utils.CHUNK_SIZE)
			} else {
				chunk = utils.AESEncrypt(utils.GetNextDataChunk(file, utils.CHUNK_SIZE), key)
			}
			hashFunc.Write(chunk)
			temp := hashFunc.Sum(nil)
			fmt.Printf("Stored chunk no %d with hash %s\n", noChunks, hex.EncodeToString(temp))
			g.addStoredChunk(hex.EncodeToString(temp), chunk)
			chunkHashed = append(chunkHashed, temp...)
		}

		if len(chunkHashed) > utils.CHUNK_SIZE {
			fmt.Printf("CAN'T INDEX FILE %s: METAFILE LARGER THAN ALLOWED\n", msg.FileName)
			return
		}

		hashFunc.Reset()
		hashFunc.Write(chunkHashed)
		metaHash := hex.EncodeToString(hashFunc.Sum(nil))
		g.addStoredChunk(metaHash, chunkHashed)

		r[j] = utils.Replica{EncryptionKey: key, Metafilehash: metaHash}
		fmt.Printf("Stored replication no %d with metafilehash %s\n", j, metaHash)
	}
	file.Close()

	f := utils.File{Name: msg.FileName, MetafileHash: utils.ByteMetaHash(r[0].Metafilehash), Size: fileSize}
	p := utils.PrivateFile{File: f, Replications: r[1:]}
	// Store seperately from other files and trigger exchange requests
	g.addPrivateFile(p)
	g.newFileExchangeRequest(r[1:])
}

// Trigger broadcast of a FileExchangeRequest for each replica
func (g *Gossiper) newFileExchangeRequest(reps []utils.Replica) {
	for _, i := range reps {
		g.setFileExchangeChannel(i.Metafilehash, make(chan utils.FileExchangeRequest, utils.MSG_BUFFER))
		req := utils.FileExchangeRequest{Origin: g.Name, Status: "OFFER", HopLimit: utils.HOPLIMIT_CONSTANT, MetaFileHash: i.Metafilehash}
		gp := utils.GossipPacket{FileExchangeRequest: &req}
		g.broadcastMessage(gp, "")
		log.Printf("Sent new file exchange request for the replication with MetaFileHash %s\n", i.Metafilehash)
	}
}

// Handles incoming FileExchangeRequest messages
func (g *Gossiper) fileExchangeRequestHandler(msg utils.FileExchangeRequest, sender string) {
	wg := sync.WaitGroup{}
	g.updateNextHop(msg.Origin, 0, sender)

	fmt.Printf("Received fileExchangeRequest: %+v\n", msg)
	if msg.Origin == g.Name {
		return
	}
	switch msg.Status {
	case "OFFER":
		fmt.Println("OFFER")
		r := g.scanOpenFileExchanges(msg.Origin)
		if r.Metafilehash != "" {
			fmt.Println("Interested, create channel")
			if e := g.AssignReplica(r.Metafilehash, msg.Origin, msg.MetaFileHash); e == nil {
				resp := utils.FileExchangeRequest{Origin: g.Name, Destination: msg.Origin, Status: "ACCEPT", MetaFileHash: msg.MetaFileHash, ExchangeMetaFileHash: r.Metafilehash, HopLimit: utils.HOPLIMIT_CONSTANT}
				gp := utils.GossipPacket{FileExchangeRequest: &resp}
				wg.Add(1)
				go func(gp utils.GossipPacket, msg utils.FileExchangeRequest) {
					g.sendToPeer(gp, g.getNextHop(msg.Origin).Address)
					g.holdingFileExchangeRequest(msg, r)
					wg.Done()
				}(gp, msg)
				wg.Wait()
				return
			}
			fmt.Println("Couldn't assign replica")
		} else {
			fmt.Println("Not interested")
		}
		break
	case "ACCEPT":
		if msg.Destination == g.Name {
			fmt.Println("ACCEPT")
			// Check if peer already stores a replication of that private file
			if g.checkReplicationAtPeer(msg.MetaFileHash, msg.Origin) {
				return
			}
			// Assign Replica. Discard message if the Replica is already assigned to another file.
			if e := g.AssignReplica(msg.MetaFileHash, msg.Origin, msg.ExchangeMetaFileHash); e != nil {
				return
			}
			resp := utils.FileExchangeRequest{Origin: g.Name, Destination: msg.Origin, Status: "FIX", MetaFileHash: msg.MetaFileHash, ExchangeMetaFileHash: msg.ExchangeMetaFileHash, HopLimit: utils.HOPLIMIT_CONSTANT}
			gp := utils.GossipPacket{FileExchangeRequest: &resp}
			go g.sendToPeer(gp, g.getNextHop(resp.Destination).Address)
			g.establishFileExchange(msg, true)
			return
		}
		break
	case "FIX":
		if msg.Destination == g.Name {
			g.sendToFileExchangeChannel(msg.MetaFileHash, msg)
			return
		}
	}
	// Not interested - Forward
	msg.HopLimit--
	if msg.HopLimit <= 0 {
		return
	}
	gp := utils.GossipPacket{FileExchangeRequest: &msg}
	g.broadcastMessage(gp, sender)
}

// Blocks NodeID until timeout is reached so that not multiple ACCEPTS are sent out for the same instance
func (g *Gossiper) holdingFileExchangeRequest(msg utils.FileExchangeRequest, r utils.Replica) {
	msg.ExchangeMetaFileHash = r.Metafilehash
	g.setFileExchangeChannel(msg.MetaFileHash, make(chan utils.FileExchangeRequest, utils.MSG_BUFFER))
	timeout := time.After(utils.FILE_EXCHANGE_TIMEOUT * time.Second)
	select {
	case <-timeout:
		fmt.Println("Timeout in holdingFileExchangeRequest")
		g.freeReplica(r.Metafilehash)
		g.deleteFileExchangeChannel(msg.MetaFileHash)
		break
	case resp := <-g.getFileExchangeChannel(msg.MetaFileHash):
		if resp.Status == "FIX" {
			fmt.Println("No timeout, received fix response.")
			g.establishFileExchange(resp, false)
		}
	}
}

// Triggered once a noce received an ACCEPT or FIX status.
// --> Begin to download the peer's file and start sending out and keeping track of challenges.
func (g *Gossiper) establishFileExchange(msg utils.FileExchangeRequest, initiatingNode bool) {
	fmt.Println("Establishing file Exchange", initiatingNode)
	var fails int
	var ownRep utils.Replica
	var otherMFH string
	if initiatingNode {
		ownRep = g.GetReplica(msg.MetaFileHash)
		otherMFH = msg.ExchangeMetaFileHash
	} else {
		ownRep = g.GetReplica(msg.ExchangeMetaFileHash)
		otherMFH = msg.MetaFileHash
	}
	g.deleteFileExchangeChannel(msg.MetaFileHash)
	g.setChallengeChannel(ownRep.Metafilehash, make(chan utils.Challenge, 1024))
	m := utils.Message{FileName: otherMFH, Destination: ownRep.NodeID, Request: otherMFH}
	g.sendDataRequest(m, ownRep.NodeID)
	interval := utils.FILE_EXCHANGE_TIMEOUT

	for {
		<-time.After(time.Duration(interval) * time.Second)
		g.poseChallenge(ownRep)
		fmt.Printf("Sending out PoR request\n")
		timeout := time.After(utils.FILE_EXCHANGE_TIMEOUT * time.Second)
		select {
		case <-timeout:
			fmt.Printf("Peer failed to respond\n")
			interval = utils.FILE_EXCHANGE_TIMEOUT
			fails++
			break
		// Receive challenge response
		case msg := <-g.getChallengeChannel(ownRep.Metafilehash):
			if g.verifyChallenge(msg) {
				fmt.Printf("Peer solved the challenge.\n")
				interval *= (1 + rand.Intn(1))
			} else {
				fmt.Printf("Peer solved the challenge incorrectly.\n")
				interval = utils.FILE_EXCHANGE_TIMEOUT
				fails++
			}
		}
		if fails > utils.CHALLENGE_FAIL_THRESHOLD {
			fmt.Printf("Peer failed too many times. Removing the reference now.\n")
			g.removeFile(otherMFH)
			g.freeReplica(ownRep.Metafilehash)
			return
		}
	}
}

// Removes all chunks of a stored file
func (g *Gossiper) removeFile(mfh string) {
	fmt.Println("Removing the file with metafilehash", mfh)
	metaFile := g.getStoredChunk(mfh)
	noChunks := utils.CeilIntDiv(len(metaFile), len(mfh))
	for i := 0; i < noChunks; i++ {
		hash := utils.GetHashAtIndex(metaFile, i)
		g.removeStoredChunk(utils.StringHash(hash))
	}
	g.removeStoredChunk(mfh)
}
