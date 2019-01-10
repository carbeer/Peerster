package gossiper

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/carbeer/Peerster/utils"
	"github.com/gorilla/mux"
)

// Helper functions
func (g *Gossiper) unmarshalAndForward(r *http.Request) error {
	var msg utils.Message
	e := json.NewDecoder(r.Body).Decode(&msg)
	utils.HandleError(e)
	return g.ClientMessageHandler(msg)
}

// Handler functions
func (g *Gossiper) handlePeer(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		g.unmarshalAndForward(r)
		utils.MarshalAndWrite(w, http.StatusOK)
		break
	case http.MethodGet:
		peers := g.GetAllPeers()
		utils.MarshalAndWrite(w, peers)
		break
	default:
		utils.MarshalAndWrite(w, http.StatusMethodNotAllowed)
	}
}

func (g *Gossiper) handleMessage(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		g.unmarshalAndForward(r)
		utils.MarshalAndWrite(w, http.StatusOK)
		break
	case http.MethodGet:
		utils.MarshalAndWrite(w, g.getAllRumorMessages())
		break
	default:
		utils.MarshalAndWrite(w, http.StatusMethodNotAllowed)
	}
}

func (g *Gossiper) handlePrivateMessage(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		g.unmarshalAndForward(r)
		utils.MarshalAndWrite(w, http.StatusOK)
		break
	case http.MethodGet:
		keys, _ := r.URL.Query()["peer"]
		utils.MarshalAndWrite(w, g.getAllPrivateMessages(keys[0]))
		break
	default:
		utils.MarshalAndWrite(w, http.StatusMethodNotAllowed)
	}
}

func (g *Gossiper) handleId(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		utils.MarshalAndWrite(w, g.Name)
		return
	}
	utils.MarshalAndWrite(w, http.StatusMethodNotAllowed)
}

func (g *Gossiper) handleFile(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		g.unmarshalAndForward(r)
		utils.MarshalAndWrite(w, http.StatusOK)
		break
	default:
		utils.MarshalAndWrite(w, http.StatusMethodNotAllowed)
	}
}

func (g *Gossiper) handlePrivateFile(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		mapping := g.getAllPrivateFiles()
		utils.MarshalAndWrite(w, mapping)
		break
	case http.MethodPost:
		g.unmarshalAndForward(r)
		utils.MarshalAndWrite(w, http.StatusOK)
		break
	default:
		utils.MarshalAndWrite(w, http.StatusMethodNotAllowed)
	}
}

func (g *Gossiper) handleGetOrigin(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		utils.MarshalAndWrite(w, g.GetAllOrigins())
		break
	default:
		utils.MarshalAndWrite(w, http.StatusMethodNotAllowed)
	}
}

func (g *Gossiper) handleDownload(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		err := g.unmarshalAndForward(r)
		if err == nil {
			utils.MarshalAndWrite(w, http.StatusInternalServerError)
		} else {
			utils.MarshalAndWrite(w, http.StatusOK)
		}
		break
	default:
		utils.MarshalAndWrite(w, http.StatusMethodNotAllowed)
	}
}

func (g *Gossiper) handleSearchRequest(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		g.unmarshalAndForward(r)
		utils.MarshalAndWrite(w, http.StatusOK)
		break
	case http.MethodGet:
		files := g.getAvailableFileResults(nil)
		utils.MarshalAndWrite(w, files)
		break
	default:
		utils.MarshalAndWrite(w, http.StatusMethodNotAllowed)
	}
}

func (g *Gossiper) handleExport(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		g.unmarshalAndForward(r)
		utils.MarshalAndWrite(w, http.StatusOK)
		break
	case http.MethodGet:
		export := g.ExportPrivateFiles()
		w.Write([]byte(export))
		break
	default:
		utils.MarshalAndWrite(w, http.StatusMethodNotAllowed)
	}
}

func serveFavicon(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "webpage/favicon.ico")
}

func (g *Gossiper) BootstrapUI() {
	fmt.Println("Starting UI on " + fmt.Sprintf("%s:%s", g.Address.IP, utils.UI_PORT))
	r := mux.NewRouter()
	r.SkipClean(true)
	r.HandleFunc("/message", g.handleMessage).Methods("POST", "GET")
	r.HandleFunc("/privateMessage", g.handlePrivateMessage).Methods("POST", "GET")
	r.HandleFunc("/node", g.handlePeer).Methods("POST", "GET")
	r.HandleFunc("/id", g.handleId).Methods("GET")
	r.HandleFunc("/origins", g.handleGetOrigin).Methods("GET")
	r.HandleFunc("/file", g.handleFile).Methods("POST")
	r.HandleFunc("/privateFile", g.handlePrivateFile).Methods("GET", "POST")
	r.HandleFunc("/download", g.handleDownload).Methods("POST")
	r.HandleFunc("/searchRequest", g.handleSearchRequest).Methods("GET", "POST")
	r.HandleFunc("/favicon.ico", serveFavicon)
	r.HandleFunc("/export", g.handleExport).Methods("GET", "POST")
	r.Handle("/", http.FileServer(http.Dir("webpage/"))).Methods("GET")
	r.PathPrefix("/js/").Handler(http.StripPrefix("/js/", http.FileServer(http.Dir("webpage/js/"))))
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("webpage/static/"))))

	utils.HandleError(http.ListenAndServe(fmt.Sprintf("%s:%s", utils.CLIENT_IP, utils.UI_PORT), r))
}
