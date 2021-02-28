package server

import (
	"context"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"

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

	// Schedules []*schedule
}

var savedData = persistentData{}
var resticExe *restic.Restic

func debug(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "hello\n")
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
} // use default options
func ws(w http.ResponseWriter, req *http.Request) {
	c, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
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
			break
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
			msg, err := newBackup(data)
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
			msg, err := updateBackup(data)
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
		}

		err = c.WriteMessage(mt, resp)
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
}

func loadData() {
	if _, err := os.Stat("data"); os.IsNotExist(err) {
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
		// if savedData.Schedules == nil {
		// 	savedData.Schedules = make([]*schedule, 0)
		// }
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
	// if savedData.Schedules == nil {
	// 	savedData.Schedules = make([]*schedule, 0)
	// }
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

// StartServer to serve HTTP.
func StartServer(conf *config.Config) {
	loadData()
	defer saveData()

	buildSchedules()

	if conf.Restic == "" {
		log.Fatal("no path to restic defined - exiting")
	}

	resticExe = restic.New(conf.Restic)

	ctx, cancel := context.WithCancel(context.Background())
	go queueHandler(ctx)

	fs := http.FileServer(http.Dir("www"))
	http.Handle("/", fs)
	http.HandleFunc("/ws", ws)

	server := &http.Server{Addr: fmt.Sprintf("%s:%d", conf.Listen.IP, conf.Listen.Port)}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	// Setting up signal capturing
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	<-stop
	defer cancel()
	defer stopSchedulers()
	if err := server.Shutdown(ctx); err != nil && err != context.DeadlineExceeded {
		log.Fatal(err)
	}

	log.Println("stopping")
}

func startHTTPServer(wg *sync.WaitGroup, connStr string) *http.Server {
	srv := &http.Server{Addr: connStr}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "hello world\n")
	})

	go func() {
		defer wg.Done() // let main know we are done cleaning up

		// always returns error. ErrServerClosed on graceful close
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			// unexpected error. port in use?
			log.Fatalf("ListenAndServe(): %v", err)
		}
	}()

	// returning reference so caller can call Shutdown()
	return srv
}
