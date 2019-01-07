package gossiper

import (
	"crypto"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/carbeer/Peerster/utils"
)

func (g *Gossiper) PrivateFileIndexing(msg utils.Message) {
	fmt.Printf("REQUESTING PRIVATE INDEXING filename %s with %d Replications\n", msg.FileName, msg.Replications)
	var hashFunc = crypto.SHA256.New()
	file, e := os.Open(filepath.Join(".", utils.SHARED_FOLDER, msg.FileName))
	utils.HandleError(e)
	fileInfo, e := file.Stat()
	utils.HandleError(e)

	fileSize := fileInfo.Size()
	noChunks := utils.CeilIntDiv(int(fileSize), utils.CHUNK_SIZE)

	// First replication is without encryption --> original file
	r := make([]utils.Replica, msg.Replications+1)

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
				chunk = AESEncrypt(utils.GetNextDataChunk(file, utils.CHUNK_SIZE), key)
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
	g.addPrivateFile(p)
	g.newFileExchangeRequest(r[1:])
}

func (g *Gossiper) newFileExchangeRequest(reps []utils.Replica) {
	for _, i := range reps {
		g.setFileExchangeChannel(i.Metafilehash, make(chan utils.FileExchangeRequest, utils.MSG_BUFFER))
		req := utils.FileExchangeRequest{Origin: g.Name, Status: "OFFER", HopLimit: utils.HOPLIMIT_CONSTANT, MetaFileHash: i.Metafilehash}
		gp := utils.GossipPacket{FileExchangeRequest: &req}
		g.broadcastMessage(gp, "")
		log.Printf("Sent new file exchange request for the replication with MetaFileHash %s\n", i.Metafilehash)
	}
}

func (g *Gossiper) fileExchangeRequestHandler(msg utils.FileExchangeRequest, sender string) {
	g.updateNextHop(msg.Origin, 0, sender)

	fmt.Printf("Received fileExchangeRequest: %+v\n", msg)
	switch msg.Status {
	case "OFFER":
		if msg.Origin == g.Name {
			return
		}
		fmt.Println("OFFER")
		r := g.scanOpenFileExchanges()
		// Not interested - Forward
		if r.Metafilehash == "" {
			fmt.Println("Not interested")
			msg.HopLimit--
			if msg.HopLimit <= 0 {
				return
			}
			gp := utils.GossipPacket{FileExchangeRequest: &msg}
			g.broadcastMessage(gp, sender)
			// Interested - Create channel
		} else {
			fmt.Println("Interested, create channel")
			if e := g.assignReplica(r.Metafilehash, msg.Origin, msg.MetaFileHash); e != nil {
				fmt.Println("Couldn't assign replica")
				return
			}
			resp := utils.FileExchangeRequest{Origin: g.Name, Destination: msg.Origin, Status: "ACCEPT", MetaFileHash: msg.MetaFileHash, ExchangeMetaFileHash: r.Metafilehash, HopLimit: utils.HOPLIMIT_CONSTANT}
			gp := utils.GossipPacket{FileExchangeRequest: &resp}
			go g.sendToPeer(gp, g.getNextHop(msg.Origin).Address)
			fmt.Println("Now starting holdingFileExchange")
			g.holdingFileExchangeRequest(msg, r)
		}
		break
	case "ACCEPT":
		fmt.Println("ACCEPT")
		if e := g.assignReplica(msg.MetaFileHash, msg.Origin, msg.ExchangeMetaFileHash); e != nil {
			fmt.Println("Couldn't assign replica")
			return
		}
		fmt.Printf("Rifht agter assgn for %s: %+v\n", msg.MetaFileHash, g.GetReplica(msg.MetaFileHash))
		resp := utils.FileExchangeRequest{Origin: g.Name, Destination: msg.Origin, Status: "FIX", MetaFileHash: msg.MetaFileHash, ExchangeMetaFileHash: msg.ExchangeMetaFileHash, HopLimit: utils.HOPLIMIT_CONSTANT}
		gp := utils.GossipPacket{FileExchangeRequest: &resp}
		go g.sendToPeer(gp, g.getNextHop(resp.Destination).Address)
		g.establishFileExchange(msg, true)
		break
	case "FIX":
		fmt.Printf("FIX: %+v\n", msg)
		g.sendToFileExchangeChannel(msg.MetaFileHash, msg)
		break
	}
}

// Blocks NodeID until timeout is reached
func (g *Gossiper) holdingFileExchangeRequest(msg utils.FileExchangeRequest, r utils.Replica) {
	msg.ExchangeMetaFileHash = r.Metafilehash
	fmt.Printf("Assigning with the flw msg: %+v\n", msg)
	g.setFileExchangeChannel(msg.MetaFileHash, make(chan utils.FileExchangeRequest, utils.MSG_BUFFER))
	fmt.Println("Waiting for response on channel", msg.MetaFileHash)
	timeout := time.After(utils.FILE_EXCHANGE_TIMEOUT)
	select {
	case <-timeout:
		fmt.Println("Timeout, resetting nodeID")
		g.freeReplica(r.Metafilehash)
		g.deleteFileExchangeChannel(msg.MetaFileHash)
		break
	case resp := <-g.getFileExchangeChannel(msg.MetaFileHash):
		fmt.Printf("Received response: %+v\n", resp)
		if resp.Status == "FIX" {
			log.Println("No timeout, received fix response.")
			g.establishFileExchange(resp, false)
		}
	}
}

func (g *Gossiper) establishFileExchange(msg utils.FileExchangeRequest, initiatingNode bool) {
	fmt.Println("Establishing file Exchange", initiatingNode)
	fmt.Printf("With message %+v\n", msg)
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
	fmt.Printf("ownrep at beginning: %+v\n", ownRep)
	g.deleteFileExchangeChannel(msg.MetaFileHash)
	g.setChallengeChannel(ownRep.Metafilehash, make(chan utils.Challenge, 1024))
	m := utils.Message{FileName: otherMFH, Destination: msg.Origin, Request: otherMFH}

	log.Println("beforeRequest", ownRep.NodeID)
	log.Println("beforeRequest2", msg.Origin)

	fmt.Printf("msg in efe: %+v\n", m)
	g.sendDataRequest(m, ownRep.NodeID)

	interval := utils.FILE_EXCHANGE_TIMEOUT

	for {
		<-time.After(interval)
		fmt.Printf("ownrep: %+v\n", ownRep)
		g.poseChallenge(ownRep)
		timeout := time.After(utils.FILE_EXCHANGE_TIMEOUT)
		select {
		case <-timeout:
			log.Println("Didn't get feedback")
			interval = utils.FILE_EXCHANGE_TIMEOUT
			fails++
			break
		case msg := <-g.getChallengeChannel(ownRep.Metafilehash):
			log.Println("Got feedback")
			if g.verifyChallenge(msg) {
				// interval *= 2
			} else {
				interval = utils.FILE_EXCHANGE_TIMEOUT
				fails++
			}
		}
		if fails > utils.CHALLENGE_FAIL_THRESHOLD {
			log.Println("Failed too many times. Removing the reference now.")
			g.removeFile(otherMFH)
			g.freeReplica(ownRep.Metafilehash)
			return
		}
	}
}

func (g *Gossiper) removeFile(mfh string) {
	fmt.Println("Removing the file with metafilehash", mfh)
	metaFile := g.getStoredChunk(mfh)
	noChunks := utils.CeilIntDiv(len(metaFile), len(mfh))
	for i := 0; i < noChunks; i++ {
		hash := utils.GetHashAtIndex(metaFile, i)
		fmt.Println("Removing the file with hash", utils.StringHash(hash))
		g.removeStoredChunk(utils.StringHash(hash))
	}
	g.removeStoredChunk(mfh)
}

func (g *Gossiper) challengeHandler(msg utils.Challenge, sender string) {
	g.updateNextHop(msg.Origin, 0, sender)
	// Only forwarding
	if msg.Destination != g.Name {
		log.Println("Forwarding challenge")
		msg.HopLimit -= 1
		if msg.HopLimit <= 0 {
			return
		}
		gp := utils.GossipPacket{Challenge: &msg}
		g.sendToPeer(gp, msg.Destination)
		return
	}
	// Challenge response
	if len(msg.Solution) != 0 {
		fmt.Println("Received challenge response")
		g.sendToChallangeChannel(msg.MetaFileHash, msg)
	} else {
		fmt.Println("Received challenge for", msg.MetaFileHash)
		// Challenged
		msg.Destination = msg.Origin
		msg.Origin = g.Name
		tmp, e := g.solveChallenge(msg)
		if e != nil {
			utils.HandleError(e)
			return
		}
		msg.Solution = tmp
		log.Println("Solved challenge. Solution:", utils.StringHash(msg.Solution))
		gp := utils.GossipPacket{Challenge: &msg}
		g.sendToPeer(gp, g.getNextHop(msg.Destination).Address)
		log.Println("Sent out a response to a challenge")
	}
}

func (g *Gossiper) poseChallenge(rep utils.Replica) {
	fmt.Printf("Replica: %+v\n", rep)
	mfh := rep.Metafilehash
	metaFile := g.getStoredChunk(mfh)
	noChunks := utils.CeilIntDiv(len(metaFile), len(mfh))
	// TODO: Move to crypto rand instead
	ch := utils.GetHashAtIndex(metaFile, rand.Intn(noChunks))
	c := utils.Challenge{Origin: g.Name, Destination: rep.NodeID, MetaFileHash: rep.Metafilehash, ChunkHash: utils.StringHash(ch), Postpend: utils.GenerateRandomByteArr(utils.POSTPEND_LENGTH), HopLimit: utils.HOPLIMIT_CONSTANT}
	gp := utils.GossipPacket{Challenge: &c}
	log.Println("NodeID", rep.NodeID)
	g.sendToPeer(gp, g.getNextHop(rep.NodeID).Address)
	fmt.Println("Sent out a challenge for ", c.MetaFileHash)
}

func (g *Gossiper) solveChallenge(msg utils.Challenge) ([]byte, error) {
	chunk := g.getStoredChunk(msg.ChunkHash)
	if len(chunk) == 0 {
		return []byte{}, errors.New(fmt.Sprintf("Don't have chunk %s from challenge %+v", msg.ChunkHash, msg))
	}
	fmt.Printf("Have chunk %s from challenge %+v\n", msg.ChunkHash, msg)
	hasher := utils.HASH_ALGO.New()
	data := append(chunk, msg.Postpend...)
	hasher.Write(data)
	return hasher.Sum(nil), nil
}

func (g *Gossiper) verifyChallenge(msg utils.Challenge) bool {
	res, e := g.solveChallenge(msg)
	utils.HandleError(e)
	if utils.EqualByteArr(res, msg.Solution) {
		log.Println("Valid solution")
		return true
	}
	log.Printf("Solution doesn't match: %s vs %s\n", utils.StringHash(msg.Solution), utils.StringHash(res))
	return false
}
