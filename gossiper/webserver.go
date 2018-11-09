package gossiper

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/carbeer/Peerster/utils"
	"github.com/gorilla/mux"
)

type UserInterface struct {
	gossiper *Gossiper
	uiPort   int
	ip       string
}

func marshalAndWrite(w http.ResponseWriter, msg interface{}) {
	bytes, e := json.Marshal(msg)
	utils.HandleError(e)
	w.Write(bytes)
}

func (g *Gossiper) peerHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		peer := r.FormValue("peer")
		g.addPeerToListIfApplicable(peer)
	case http.MethodGet:
		peers := g.GetAllPeers()
		marshalAndWrite(w, peers)
		break
	default:
		marshalAndWrite(w, http.StatusMethodNotAllowed)
	}
}

func (g *Gossiper) handleMessage(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		message := r.FormValue("message")
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			g.ClientMessageHandler(message)
			wg.Done()
		}()
		wg.Wait()
	case http.MethodGet:
		messages := g.GetAllReceivedMessages()
		marshalAndWrite(w, messages)
		break
	default:
		marshalAndWrite(w, http.StatusMethodNotAllowed)
	}
}

func (g *Gossiper) getId(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		marshalAndWrite(w, g.name)
		return
	}
	marshalAndWrite(w, http.StatusMethodNotAllowed)
}

func (g *Gossiper) BootstrapUI() {
	log.Println("Starting UI on " + fmt.Sprintf("%s:%s", g.Address.IP, "8080"))
	r := mux.NewRouter()
	r.HandleFunc("/message", g.handleMessage).Methods("POST", "GET")
	r.HandleFunc("/node", g.peerHandler).Methods("POST", "GET")
	r.HandleFunc("/id", g.getId).Methods("GET")
	r.Handle("/", http.FileServer(http.Dir("webpage"))).Methods("GET")
	r.PathPrefix("/js/").Handler(http.StripPrefix("/js/", http.FileServer(http.Dir("webpage/js/"))))
	utils.HandleError(http.ListenAndServe(fmt.Sprintf("%s:%s", g.Address.IP, "8080"), r))
}
