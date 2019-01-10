package gossiper

import (
	"crypto"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/carbeer/Peerster/utils"
)

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
		r := g.scanOpenFileExchanges(msg.Origin)
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
			if e := g.AssignReplica(r.Metafilehash, msg.Origin, msg.MetaFileHash); e != nil {
				fmt.Println("Couldn't assign replica")
				return
			}
			resp := utils.FileExchangeRequest{Origin: g.Name, Destination: msg.Origin, Status: "ACCEPT", MetaFileHash: msg.MetaFileHash, ExchangeMetaFileHash: r.Metafilehash, HopLimit: utils.HOPLIMIT_CONSTANT}
			gp := utils.GossipPacket{FileExchangeRequest: &resp}
			fmt.Printf("%s nextHop for %s\n", g.getNextHop(msg.Origin).Address, msg.Origin)
			go g.sendToPeer(gp, g.getNextHop(msg.Origin).Address)
			g.holdingFileExchangeRequest(msg, r)
		}
		break
	case "ACCEPT":
		fmt.Println("ACCEPT")
		if g.checkReplicationAtPeer(msg.MetaFileHash, msg.Origin) {
			return
		}
		if e := g.AssignReplica(msg.MetaFileHash, msg.Origin, msg.ExchangeMetaFileHash); e != nil {
			return
		}
		resp := utils.FileExchangeRequest{Origin: g.Name, Destination: msg.Origin, Status: "FIX", MetaFileHash: msg.MetaFileHash, ExchangeMetaFileHash: msg.ExchangeMetaFileHash, HopLimit: utils.HOPLIMIT_CONSTANT}
		gp := utils.GossipPacket{FileExchangeRequest: &resp}
		go g.sendToPeer(gp, g.getNextHop(resp.Destination).Address)
		g.establishFileExchange(msg, true)
		break
	case "FIX":
		g.sendToFileExchangeChannel(msg.MetaFileHash, msg)
		break
	}
}

// Blocks NodeID until timeout is reached
func (g *Gossiper) holdingFileExchangeRequest(msg utils.FileExchangeRequest, r utils.Replica) {
	msg.ExchangeMetaFileHash = r.Metafilehash
	g.setFileExchangeChannel(msg.MetaFileHash, make(chan utils.FileExchangeRequest, utils.MSG_BUFFER))
	timeout := time.After(utils.FILE_EXCHANGE_TIMEOUT)
	select {
	case <-timeout:
		g.freeReplica(r.Metafilehash)
		g.deleteFileExchangeChannel(msg.MetaFileHash)
		break
	case resp := <-g.getFileExchangeChannel(msg.MetaFileHash):
		if resp.Status == "FIX" {
			log.Println("No timeout, received fix response.")
			g.establishFileExchange(resp, false)
		}
	}
}

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
		log.Println("Still in the loop")
		<-time.After(interval)
		g.poseChallenge(ownRep)
		fmt.Printf("Sending out PoR request\n")
		timeout := time.After(utils.FILE_EXCHANGE_TIMEOUT)
		select {
		case <-timeout:
			fmt.Printf("Peer failed to respond\n")
			interval = utils.FILE_EXCHANGE_TIMEOUT
			fails++
			break
		case msg := <-g.getChallengeChannel(ownRep.Metafilehash):
			if g.verifyChallenge(msg) {
				fmt.Printf("Peer solved the challenge.\n")
				interval *= 2
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
