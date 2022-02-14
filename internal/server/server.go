package server

import (
	"context"
	_ "embed"
	"encoding/gob"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"zerosrealm.xyz/tergum/internal/entities"
	"zerosrealm.xyz/tergum/internal/log"
	"zerosrealm.xyz/tergum/internal/restic"
	"zerosrealm.xyz/tergum/internal/server/config"
	manager "zerosrealm.xyz/tergum/internal/server/manager"
	"zerosrealm.xyz/tergum/internal/server/service"
)

type PersistentData struct {
	Mutex sync.Mutex

	// Repos   []*entities.Repo
	// Agents  []*entities.Agent
	// Backups []*entities.Backup

	// RepoIncrement   int
	// AgentIncrement  int
	// BackupIncrement int

	BackupSubscribers map[int][]*entities.Agent

	Jobs map[string][]byte

	// Schedules []*schedule
}

type Server struct {
	ctx       context.Context
	ctxCancel context.CancelFunc
	services  *service.Services

	manager *manager.Manager
	conf    *config.Config
	router  *mux.Router
	log     *log.Logger
}

var savedData = PersistentData{
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
	// if savedData.Agents == nil {
	// 	savedData.Agents = make([]*entities.Agent, 0)
	// }
	// if savedData.Backups == nil {
	// 	savedData.Backups = make([]*entities.Backup, 0)
	// }
	// if savedData.Repos == nil {
	// 	savedData.Repos = make([]*entities.Repo, 0)
	// }
	if savedData.BackupSubscribers == nil {
		savedData.BackupSubscribers = make(map[int][]*entities.Agent)
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

func New(conf *config.Config, services *service.Services) (*Server, error) {
	ctx, cancel := context.WithCancel(context.Background())

	// fields := make(map[string]interface{})
	logger, err := log.New(&conf.Log, nil)
	if err != nil {
		cancel()
		return nil, err
	}
	man := manager.NewManager(ctx, services, logger, &wsConnections)

	srv := &Server{
		ctx:       ctx,
		ctxCancel: cancel,
		services:  services,
		manager:   man,
		conf:      conf,
		router:    mux.NewRouter(),
		log:       logger,
	}
	srv.routes()
	return srv, nil
}

// Start to serve HTTP.
func (srv *Server) Start() {
	loadData()
	defer srv.log.Close()

	if srv.conf.Restic == "" {
		srv.log.Fatal("no path to restic defined - exiting")
	}

	resticExe = restic.New(srv.ctx, srv.conf.Restic)
	go srv.manager.Start()

	srv.manager.BuildSchedules()

	srv.router.Handle("/", http.FileServer(http.Dir("www")))
	srv.router.HandleFunc("/ws", srv.ws)

	listener := &http.Server{
		Handler:      srv,
		Addr:         fmt.Sprintf("%s:%d", srv.conf.Listen.IP, srv.conf.Listen.Port),
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
	}

	srv.log.Info(fmt.Sprintf("Listening on %s:%d", srv.conf.Listen.IP, srv.conf.Listen.Port))

	go func() {
		if err := listener.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			srv.log.Fatal(err)
		}
	}()

	// Setting up signal capturing
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	srv.log.Info("shutting down")

	defer srv.ctxCancel()
	defer manager.StopSchedulers()
	if err := listener.Shutdown(srv.ctx); err != nil && err != context.DeadlineExceeded {
		srv.log.Fatal(err)
	}
}
