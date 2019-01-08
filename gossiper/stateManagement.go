package gossiper

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/carbeer/Peerster/utils"
)

func (g *Gossiper) SaveState() {
	obj, e := json.MarshalIndent(g, "", "\t")
	utils.HandleError(e)
	_ = os.Mkdir(utils.STATE_FOLDER, os.ModePerm)
	cwd, _ := os.Getwd()
	e = ioutil.WriteFile(filepath.Join(cwd, utils.STATE_FOLDER, fmt.Sprint(g.Name, ".json")), obj, 0644)
	utils.HandleError(e)
}

func LoadState(id string) *Gossiper {
	g := Gossiper{}
	cwd, _ := os.Getwd()
	f, e := ioutil.ReadFile(filepath.Join(cwd, utils.STATE_FOLDER, fmt.Sprint(id, ".json")))
	utils.HandleError(e)
	json.Unmarshal(f, &g)
	log.Printf("Got this: %+v\n", &g)
	return &g
}

func (g *Gossiper) ExportPrivateFiles() string {
	obj, e := json.MarshalIndent(g.PrivFiles, "", "\t")
	utils.HandleError(e)
	return string(obj)
}

func (g *Gossiper) LoadPrivateFiles(name string) []utils.PrivateFile {
	data, e := ioutil.ReadFile(filepath.Join(".", utils.SHARED_FOLDER, name))
	utils.HandleError(e)
	p := []utils.PrivateFile{}
	json.Unmarshal(data, &p)
	for _, el := range p {
		g.addPrivateFile(el)
	}
	fmt.Println("UPLOADED the private file state")
	return p
}

func (g *Gossiper) updateNextHop(origin string, id uint32, sender string) {
	if id > g.getNextHop(origin).HighestID {
		g.setNextHop(origin, utils.HopInfo{Address: sender, HighestID: id})
		fmt.Printf("DSDV %s %s\n", origin, sender)
	}
}
