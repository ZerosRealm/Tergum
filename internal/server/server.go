package server

import (
	"context"
	_ "embed"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"

	"github.com/gorilla/websocket"
	"zerosrealm.xyz/tergum/internal/restic"
	"zerosrealm.xyz/tergum/internal/server/config"
	"zerosrealm.xyz/tergum/internal/types"
)

type persistentData struct {
	Repos   []*types.Repo
	Agents  []*types.Agent
	Backups []*types.Backup

	RepoIncrement   int
	AgentIncrement  int
	BackupIncrement int

	BackupSubscribers map[int][]*types.Agent

	Jobs map[string][]byte

	// Schedules []*schedule
}

type Server struct {
	ctx       context.Context
	ctxCancel context.CancelFunc

	manager *Manager
	conf    *config.Config
}

var savedData = persistentData{}
var resticExe *restic.Restic

var wsConnections = make(map[string]*websocket.Conn)

func closeWS(c *websocket.Conn) {
	key := c.RemoteAddr().String()
	_, ok := wsConnections[key]
	if ok {
		delete(wsConnections, key)
	}
	c.Close()
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
} // use default options
func (srv *Server) ws(w http.ResponseWriter, req *http.Request) {
	c, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	wsConnections[c.RemoteAddr().String()] = c
	defer closeWS(c)
	for {
		mt, msg, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}

		var data map[string]interface{}
		err = json.Unmarshal(msg, &data)
		if err != nil {
			log.Println(err)
			break
		}

		var msgType string
		switch v := data["type"].(type) {
		case string:
			msgType = v
		default:
			log.Println("message type data sent was invalid")
		}

		var resp []byte
		switch strings.ToLower(msgType) {
		case "getbackups":
			msg, err := getBackups()
			if err != nil {
				log.Println(err)
			}

			resp = msg
		case "getrepos":
			msg, err := getRepos()
			if err != nil {
				log.Println(err)
			}

			resp = msg
		case "getagents":
			msg, err := getAgents()
			if err != nil {
				log.Println(err)
			}

			resp = msg
		case "newbackup":
			msg, err := srv.newBackup(data)
			if err != nil {
				log.Println(err)
			}

			resp = msg
		case "newrepo":
			msg, err := newRepo(data)
			if err != nil {
				log.Println(err)
			}

			resp = msg
		case "newagent":
			msg, err := newAgent(data)
			if err != nil {
				log.Println(err)
			}

			resp = msg
		case "updatebackup":
			msg, err := srv.updateBackup(data)
			if err != nil {
				log.Println(err)
			}

			resp = msg
		case "updaterepo":
			msg, err := updateRepo(data)
			if err != nil {
				log.Println(err)
			}

			resp = msg
		case "updateagent":
			msg, err := updateAgent(data)
			if err != nil {
				log.Println(err)
			}

			resp = msg
		case "deletebackup":
			msg, err := deleteBackup(data)
			if err != nil {
				log.Println(err)
			}

			resp = msg
		case "deleterepo":
			msg, err := deleteRepo(data)
			if err != nil {
				log.Println(err)
			}

			resp = msg
		case "deleteagent":
			msg, err := deleteAgent(data)
			if err != nil {
				log.Println(err)
			}

			resp = msg
		case "getsubscribers":
			msg, err := getSubscribers()
			if err != nil {
				log.Println(err)
			}

			resp = msg
		case "updatesubscribers":
			msg, err := updateSubscribers(data)
			if err != nil {
				log.Println(err)
			}

			resp = msg
		case "getsnapshots":
			msg, err := getSnapshots(data)
			if err != nil {
				log.Println(err)
			}

			resp = msg
		case "restoresnapshot":
			msg, err := srv.restoreSnapshot(data)
			if err != nil {
				log.Println(err)
			}

			resp = msg
		case "deletesnapshot":
			msg, err := deleteSnapshot(data)
			if err != nil {
				log.Println(err)
			}

			resp = msg
		case "getjobs":
			msg, err := srv.getJobs()
			if err != nil {
				log.Println(err)
			}

			resp = msg
		}

		err = c.WriteMessage(mt, resp)
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
}

func prepareSavedData() {
	if savedData.Agents == nil {
		savedData.Agents = make([]*types.Agent, 0)
	}
	if savedData.Backups == nil {
		savedData.Backups = make([]*types.Backup, 0)
	}
	if savedData.Repos == nil {
		savedData.Repos = make([]*types.Repo, 0)
	}
	if savedData.BackupSubscribers == nil {
		savedData.BackupSubscribers = make(map[int][]*types.Agent)
	}
	if savedData.Jobs == nil {
		savedData.Jobs = make(map[string][]byte)
	}
}

func loadData() {
	defer prepareSavedData()

	if _, err := os.Stat("data"); os.IsNotExist(err) {
		return
	}

	f, err := os.Open("data")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	dec := gob.NewDecoder(f)
	err = dec.Decode(&savedData)
	if err != nil {
		panic(err)
	}
}

func saveData() {
	f, err := os.Create("data")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	enc := gob.NewEncoder(f)
	enc.Encode(savedData)
}

func (srv *Server) update() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("authorization")
		auth := strings.SplitN(authHeader, " ", 2)
		if len(auth) != 2 || strings.ToLower(auth[0]) != "psk" {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		psk := auth[1]
		access := false
		for _, agent := range savedData.Agents {
			if agent.PSK == psk {
				access = true
				break
			}
		}

		if !access {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()

		var data map[string]interface{}
		err = json.Unmarshal(body, &data)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		jobMsg, err := json.Marshal(data["msg"])
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		jobID := data["job"].(string)
		savedData.Jobs[jobID] = jobMsg

		srv.manager.updateProgress(jobID, jobMsg)
		srv.manager.WriteWS(body)

		w.WriteHeader(http.StatusOK)
	}
}

func New(conf *config.Config) *Server {
	ctx, cancel := context.WithCancel(context.Background())
	man := NewManager(ctx)
	srv := &Server{
		ctx:       ctx,
		ctxCancel: cancel,
		manager:   man,
		conf:      conf,
	}
	return srv
}

// Start to serve HTTP.
func (srv *Server) Start() {
	loadData()
	defer saveData()

	if srv.conf.Restic == "" {
		log.Fatal("no path to restic defined - exiting")
	}

	resticExe = restic.New(srv.conf.Restic)
	go srv.manager.Start()

	buildSchedules(srv.manager)

	http.Handle("/", http.FileServer(http.Dir("www")))
	http.HandleFunc("/ws", srv.ws)
	http.HandleFunc("/update", srv.update())

	server := &http.Server{Addr: fmt.Sprintf("%s:%d", srv.conf.Listen.IP, srv.conf.Listen.Port)}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	// Setting up signal capturing
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	<-stop
	defer srv.ctxCancel()
	defer stopSchedulers()
	if err := server.Shutdown(srv.ctx); err != nil && err != context.DeadlineExceeded {
		log.Fatal(err)
	}

	log.Println("stopping")
}
