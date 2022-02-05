package server

import (
	"context"
	_ "embed"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"zerosrealm.xyz/tergum/internal/log"
	"zerosrealm.xyz/tergum/internal/restic"
	"zerosrealm.xyz/tergum/internal/server/config"
	"zerosrealm.xyz/tergum/internal/server/service"
	"zerosrealm.xyz/tergum/internal/types"
)

type persistentData struct {
	Mutex sync.Mutex

	// Repos   []*types.Repo
	// Agents  []*types.Agent
	// Backups []*types.Backup

	// RepoIncrement   int
	// AgentIncrement  int
	// BackupIncrement int

	BackupSubscribers map[int][]*types.Agent

	Jobs map[string][]byte

	// Schedules []*schedule
}

type Services struct {
	repoSvc   service.RepoService
	agentSvc  service.AgentService
	backupSvc service.BackupService
}

func NewServices(repoSvc *service.RepoService, agentSvc *service.AgentService, backupSvc *service.BackupService) *Services {
	return &Services{
		repoSvc:   *repoSvc,
		agentSvc:  *agentSvc,
		backupSvc: *backupSvc,
	}
}

type Server struct {
	ctx       context.Context
	ctxCancel context.CancelFunc
	services  *Services

	manager *Manager
	conf    *config.Config
	router  *mux.Router
	log     *log.Logger
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
	// if savedData.Agents == nil {
	// 	savedData.Agents = make([]*types.Agent, 0)
	// }
	// if savedData.Backups == nil {
	// 	savedData.Backups = make([]*types.Backup, 0)
	// }
	// if savedData.Repos == nil {
	// 	savedData.Repos = make([]*types.Repo, 0)
	// }
	if savedData.BackupSubscribers == nil {
		savedData.BackupSubscribers = make(map[int][]*types.Agent)
	}
	if savedData.Jobs == nil {
		savedData.Jobs = make(map[string][]byte)
	}
}

func New(conf *config.Config, services *Services) (*Server, error) {
	ctx, cancel := context.WithCancel(context.Background())

	// fields := make(map[string]interface{})
	logger, err := log.New(&conf.Log, nil)
	if err != nil {
		cancel()
		return nil, err
	}
	man := NewManager(ctx, services, logger)

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
	defer srv.log.Close()

	if srv.conf.Restic == "" {
		srv.log.Fatal("no path to restic defined - exiting")
	}

	resticExe = restic.New(srv.ctx, srv.conf.Restic)
	go srv.manager.Start()

	srv.manager.buildSchedules()

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
	defer stopSchedulers()
	if err := listener.Shutdown(srv.ctx); err != nil && err != context.DeadlineExceeded {
		srv.log.Fatal(err)
	}
}
