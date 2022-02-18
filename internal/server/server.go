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
	"zerosrealm.xyz/tergum/internal/entity"
	"zerosrealm.xyz/tergum/internal/log"
	"zerosrealm.xyz/tergum/internal/restic"
	"zerosrealm.xyz/tergum/internal/server/config"
	manager "zerosrealm.xyz/tergum/internal/server/manager"
	"zerosrealm.xyz/tergum/internal/server/service"
)

type PersistentData struct {
	Mutex sync.Mutex

	// Repos   []*entity.Repo
	// Agents  []*entity.Agent
	// Backups []*entity.Backup

	// RepoIncrement   int
	// AgentIncrement  int
	// BackupIncrement int

	BackupSubscribers map[int][]*entity.Agent

	Jobs map[string][]byte

	// Schedules []*schedule
}

type Server struct {
	ctx       context.Context
	ctxCancel context.CancelFunc
	services  *service.Services

	restic  *restic.Restic
	manager *manager.Manager
	router  *mux.Router

	conf *config.Config
	log  *log.Logger
}

var savedData = PersistentData{
	Mutex: sync.Mutex{},
}

var wsConnections = make(map[string]*websocket.Conn)

func closeWS(c *websocket.Conn) {
	key := c.RemoteAddr().String()
	_, ok := wsConnections[key]
	if ok {
		delete(wsConnections, key)
	}
	c.Close()
}

func New(conf *config.Config, services *service.Services) (*Server, error) {
	ctx, cancel := context.WithCancel(context.Background())

	// fields := make(map[string]interface{})
	logger, err := log.New(&conf.Log)
	if err != nil {
		cancel()
		return nil, err
	}
	man := manager.NewManager(ctx, services, logger, &wsConnections)

	if conf.Restic == "" {
		defer logger.Close()
		logger.Fatal("no path to restic defined - exiting")
	}

	if _, err := os.Stat(conf.Restic); os.IsNotExist(err) {
		defer logger.Close()
		logger.Fatal("no restic executable found - exiting")
	}

	resticExe := restic.New(ctx, conf.Restic)

	srv := &Server{
		ctx:       ctx,
		ctxCancel: cancel,
		services:  services,
		manager:   man,
		restic:    resticExe,
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
