package gossiper

import (
	"errors"
	"fmt"
	"log"
	"math/rand"

	"github.com/carbeer/Peerster/utils"
)

func (g *Gossiper) challengeHandler(msg utils.Challenge, sender string) {
	g.updateNextHop(msg.Origin, 0, sender)
	// Only forwarding
	if msg.Destination != g.Name {
		msg.HopLimit -= 1
		if msg.HopLimit <= 0 {
			return
		}
		gp := utils.GossipPacket{Challenge: &msg}
		g.sendToPeer(gp, g.getNextHop(msg.Destination).Address)
		return
	}
	// Challenge response
	if len(msg.Solution) != 0 {
		g.sendToChallangeChannel(msg.MetaFileHash, msg)
	} else {
		// Challenged
		msg.Destination = msg.Origin
		msg.Origin = g.Name
		tmp, e := g.solveChallenge(msg)
		if e != nil {
			utils.HandleError(e)
			return
		}
		msg.Solution = tmp
		gp := utils.GossipPacket{Challenge: &msg}
		g.sendToPeer(gp, g.getNextHop(msg.Destination).Address)
	}
}

func (g *Gossiper) poseChallenge(rep utils.Replica) {
	mfh := rep.Metafilehash
	metaFile := g.getStoredChunk(mfh)
	noChunks := utils.CeilIntDiv(len(metaFile), len(mfh))
	// TODO: Move to crypto rand instead
	ch := utils.GetHashAtIndex(metaFile, rand.Intn(noChunks))
	c := utils.Challenge{Origin: g.Name, Destination: rep.NodeID, MetaFileHash: rep.Metafilehash, ChunkHash: utils.StringHash(ch), Postpend: utils.GenerateRandomByteArr(utils.POSTPEND_LENGTH), HopLimit: utils.HOPLIMIT_CONSTANT}
	gp := utils.GossipPacket{Challenge: &c}
	g.sendToPeer(gp, g.getNextHop(rep.NodeID).Address)
}

func (g *Gossiper) solveChallenge(msg utils.Challenge) ([]byte, error) {
	chunk := g.getStoredChunk(msg.ChunkHash)
	if len(chunk) == 0 {
		return []byte{}, errors.New(fmt.Sprintf("Don't have chunk %s from challenge %+v", msg.ChunkHash, msg))
	}
	hasher := utils.HASH_ALGO.New()
	data := append(chunk, msg.Postpend...)
	hasher.Write(data)
	return hasher.Sum(nil), nil
}

func (g *Gossiper) verifyChallenge(msg utils.Challenge) bool {
	res, e := g.solveChallenge(msg)
	utils.HandleError(e)
	if utils.EqualByteArr(res, msg.Solution) {
		log.Printf("Successfully verified solution\n")
		return true
	}
	log.Printf("Solution doesn't match: %s vs %s\n", utils.StringHash(msg.Solution), utils.StringHash(res))
	return false
}
