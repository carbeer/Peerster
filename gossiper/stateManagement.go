package gossiper

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
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
	fmt.Printf("Got this: %+v\n", &g)
	return &g
}

func (g *Gossiper) ExportPrivateFiles() string {
	obj, e := json.MarshalIndent(g.PrivFiles, "", "\t")
	utils.HandleError(e)
	fmt.Printf("Exported the following state: %+v\n", string(obj))
	return string(obj)
}

func (g *Gossiper) LoadPrivateFiles(name string) {
	data, e := ioutil.ReadFile(filepath.Join(".", utils.SHARED_FOLDER, name))
	utils.HandleError(e)
	p := make(map[string]utils.PrivateFile)
	e = json.Unmarshal(data, &p)
	utils.HandleError(e)
	for _, el := range p {
		g.addPrivateFile(el)
	}
	fmt.Println("UPLOADED the private file state")
}
