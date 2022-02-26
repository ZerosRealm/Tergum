package agent

import (
	"context"
	_ "embed"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/egonelbre/antifreeze"
	"github.com/gorilla/mux"
	"zerosrealm.xyz/tergum/internal/agent/config"
	"zerosrealm.xyz/tergum/internal/agent/manager"
	"zerosrealm.xyz/tergum/internal/log"
	"zerosrealm.xyz/tergum/internal/restic"
)

type Server struct {
	ctx       context.Context
	ctxCancel context.CancelFunc

	manager *manager.Manager
	restic  *restic.Restic
	router  *mux.Router

	conf *config.Config
	log  *log.Logger
}

func NewServer(conf *config.Config) (*Server, error) {
	ctx, cancel := context.WithCancel(context.Background())

	logger, err := log.New(&conf.Log)
	if err != nil {
		cancel()
		return nil, err
	}

	resticExe := restic.New(ctx, conf.Restic)

	manager, err := manager.New(ctx, conf, resticExe)
	if err != nil {
		cancel()
		return nil, err
	}

	srv := &Server{
		ctx:       ctx,
		ctxCancel: cancel,

		manager: manager,
		restic:  resticExe,
		router:  mux.NewRouter(),

		conf: conf,
		log:  logger,
	}
	srv.routes()

	return srv, nil
}

// Start to serve HTTP.
func (srv *Server) Start() {
	defer srv.log.Close()

	go srv.manager.UpdateHandler()

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
	antifreeze.Exclude()
	<-stop
	srv.log.Info("Shutting down")

	defer srv.ctxCancel()
	defer srv.manager.Cancel()
	if err := listener.Shutdown(srv.ctx); err != nil && err != context.DeadlineExceeded {
		srv.log.Error(err)
	}
}
