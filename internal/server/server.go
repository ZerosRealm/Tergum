package server

import (
	"context"
	_ "embed"
	"encoding/gob"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"zerosrealm.xyz/tergum/internal/restic"
	"zerosrealm.xyz/tergum/internal/server/config"
	"zerosrealm.xyz/tergum/internal/types"
)

type persistentData struct {
	Mutex sync.Mutex

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
	router  *mux.Router
}

var savedData = persistentData{
	Mutex: sync.Mutex{},
}
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

func New(conf *config.Config) *Server {
	ctx, cancel := context.WithCancel(context.Background())
	man := NewManager(ctx)
	srv := &Server{
		ctx:       ctx,
		ctxCancel: cancel,
		manager:   man,
		conf:      conf,
		router:    mux.NewRouter(),
	}
	srv.routes()
	return srv
}

// Start to serve HTTP.
func (srv *Server) Start() {
	loadData()
	defer saveData()

	if srv.conf.Restic == "" {
		log.Fatal("no path to restic defined - exiting")
	}

	resticExe = restic.New(srv.ctx, srv.conf.Restic)
	go srv.manager.Start()

	buildSchedules(srv.manager)

	srv.router.Handle("/", http.FileServer(http.Dir("www")))
	srv.router.HandleFunc("/ws", srv.ws)

	listener := &http.Server{
		Handler:      srv,
		Addr:         fmt.Sprintf("%s:%d", srv.conf.Listen.IP, srv.conf.Listen.Port),
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
	}

	go func() {
		if err := listener.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	// Setting up signal capturing
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
	log.Println("shutting down")

	defer srv.ctxCancel()
	defer stopSchedulers()
	if err := listener.Shutdown(srv.ctx); err != nil && err != context.DeadlineExceeded {
		log.Fatal(err)
	}
}
