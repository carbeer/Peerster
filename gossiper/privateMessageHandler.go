package gossiper

import (
	"fmt"
	"sync"

	"github.com/carbeer/Peerster/utils"
)

func (g *Gossiper) privateMessageHandler(msg utils.PrivateMessage) {
	if msg.Destination == g.name {
		fmt.Printf("PRIVATE origin %s hop-limit %d contents %s\n", msg.Origin, msg.HopLimit, msg.Text)
		g.appendPrivateMessages(msg.Origin, msg)
	} else {
		var wg sync.WaitGroup
		msg.HopLimit -= 1
		if msg.HopLimit <= 0 {
			// log.Printf("%s: ATTENTION: Dropping a private message for %s\n", g.name, msg.Destination)
			return
		}
		gossipMessage := utils.GossipPacket{Private: &msg}
		wg.Add(1)
		go func() {
			// fmt.Printf("%d: Send the private message\n", time.Now().Second())
			g.sendToPeer(gossipMessage, g.getNextHop(msg.Destination).Address)
			wg.Done()
		}()
		wg.Wait()
	}
}
