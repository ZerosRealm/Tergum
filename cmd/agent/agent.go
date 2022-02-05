package main

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/davecgh/go-spew/spew"
	"zerosrealm.xyz/tergum/internal/agent/config"
	"zerosrealm.xyz/tergum/internal/log"
	"zerosrealm.xyz/tergum/internal/restic"
	"zerosrealm.xyz/tergum/internal/types"
)

var conf config.Config
var resticExe *restic.Restic

var jobs = make(map[string]*restic.Job)
var logger *log.Logger

func updateHandler(ctx context.Context) {
	type jobProgress struct {
		Msg json.RawMessage `json:"msg"`
	}
	for {
		select {
		case update := <-resticExe.Updates:
			msg, err := json.Marshal(jobProgress{Msg: update.Msg})
			if err != nil {
				logger.WithFields("function", "updateHandler").Trace("update dump:", spew.Sdump(msg))
				logger.WithFields("function", "updateHandler").Error("marshalling update error:", err)
				panic(err)
			}

			req, err := http.NewRequest("POST", conf.Server+"/api/job/"+update.ID+"/progress", bytes.NewReader(msg))
			if err != nil {
				logger.WithFields("function", "updateHandler").Error("error:", err)
				continue
			}

			req.Header.Add("authorization", fmt.Sprintf("PSK %s", conf.PSK))

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				logger.WithFields("function", "updateHandler").Error("marshalling update error:", err)
				continue
			}
			defer resp.Body.Close()

			if resp.StatusCode != 200 {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					logger.WithFields("function", "updateHandler").Error("non-200 status:", resp.Status, "body read error:", err)
					continue
				}
				logger.WithFields("function", "updateHandler").Error("non-200 status:", resp.Status, "body:", string(body))
			}
		case job := <-resticExe.Jobs:
			if _, ok := jobs[job.ID]; ok {
				return
			}
			jobs[job.ID] = job
		default:
			select {
			case <-ctx.Done():
				return
			default:
			}
		}
	}
}

func handleConnection(c net.Conn) {
	logger.WithFields("function", "handleConnection").Debug("Serving", c.RemoteAddr().String())
	defer c.Close()

	var packet types.JobPacket
	dec := gob.NewDecoder(c)
	err := dec.Decode(&packet)

	if err != nil {
		logger.WithFields("function", "handleConnection").Error("reading conn error:", err)
		return
	}

	if packet.ID == "" {
		logger.WithFields("function", "handleConnection").Error("invalid job, ID is empty")
		return
	}

	if packet.Agent.PSK != conf.PSK {
		logger.WithFields("function", "handleConnection", "job", packet.ID).Warn("job PSK does not match")
		return
	}

	switch packet.Type {
	case "backup":
		var job types.BackupJob
		dec := gob.NewDecoder(bytes.NewReader(packet.Data.([]byte)))
		err := dec.Decode(&job)

		if err != nil {
			logger.WithFields("function", "handleConnection", "job", packet.ID).Trace("job data:", spew.Sdump(packet.Data))
			logger.WithFields("function", "handleConnection", "job", packet.ID).Error("decoding job error:", err)
			return
		}

		logger.WithFields("function", "handleConnection", "job", packet.ID).Info("Starting job")
		out, err := resticExe.Backup(packet.Repo.Repo, job.Backup.Source, packet.Repo.Password, job.Backup.Exclude, packet.ID, packet.Repo.Settings...)
		if err != nil {
			logger.WithFields("function", "handleConnection", "job", packet.ID).Error("restic backup error:", err, "output:", string(out))
			return
		}

		logger.WithFields("function", "handleConnection", "job", packet.ID).Trace("output:", string(out))
	case "restore":
		var job types.RestoreJob
		dec := gob.NewDecoder(bytes.NewReader(packet.Data.([]byte)))
		err := dec.Decode(&job)

		if err != nil {
			logger.WithFields("function", "handleConnection", "job", packet.ID).Trace("job data:", spew.Sdump(packet.Data))
			logger.WithFields("function", "handleConnection", "job", packet.ID).Error("decoding job error:", err)
			return
		}

		logger.WithFields("function", "handleConnection", "job", packet.ID).Info("Starting job")
		out, err := resticExe.Restore(packet.Repo.Repo, packet.Repo.Password, job.Snapshot,
			job.Target, job.Include, job.Exclude, packet.Repo.Settings...)
		if err != nil {
			logger.WithFields("function", "handleConnection", "job", packet.ID).Error("restic backup error:", err, "output:", string(out))
			return
		}

		logger.WithFields("function", "handleConnection", "job", packet.ID).Trace("output:", string(out))
		delete(jobs, packet.ID)
	case "stop":
		var job types.StopJob
		dec := gob.NewDecoder(bytes.NewReader(packet.Data.([]byte)))
		err := dec.Decode(&job)

		if err != nil {
			logger.WithFields("function", "handleConnection", "job", packet.ID).Trace("job data:", spew.Sdump(packet.Data))
			logger.WithFields("function", "handleConnection", "job", packet.ID).Error("decoding job error:", err)
			return
		}

		logger.WithFields("function", "handleConnection", "job", packet.ID).Info("Stopping job")
		resticJob, ok := jobs[packet.ID]
		if !ok {
			logger.WithFields("function", "handleConnection", "job", packet.ID).Info("Job not found")
			return
		}
		resticJob.Cancel()
		delete(jobs, packet.ID)

	default:
		logger.WithFields("function", "handleConnection", "job", packet.ID).Debug("Unknown job type", packet.Type)
	}
}

func main() {
	conf, err := config.Load()
	if err != nil {
		panic(err)
	}

	log, err := log.New(&conf.Log, nil)
	if err != nil {
		panic(err)
	}
	logger = log

	logger.Info("Starting agent")

	if conf.PSK == "" {
		logger.Fatal("no PSK defined - exiting")
	}

	if conf.Restic == "" {
		logger.Fatal("no path to restic defined - exiting")
	}

	if conf.Server == "" {
		logger.Fatal("no server defined - exiting")
	}

	listenStr := fmt.Sprintf("%s:%d", conf.Listen.IP, conf.Listen.Port)
	logger.Info("Listening on", listenStr)

	l, err := net.Listen("tcp4", listenStr)
	if err != nil {
		logger.Error("tcp4 listening error:", err)
		return
	}
	defer l.Close()
	rand.Seed(time.Now().Unix())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	resticExe = restic.New(ctx, conf.Restic)

	go updateHandler(ctx)

	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			c, err := l.Accept()
			if err != nil {
				logger.Error("listener accept error:", err)
				return
			}
			go handleConnection(c)
		}
	}(ctx)

	// Setting up signal capturing
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	logger.Info("Shutting down")
}
