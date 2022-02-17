package main

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/davecgh/go-spew/spew"
	"zerosrealm.xyz/tergum/internal/agent/config"
	"zerosrealm.xyz/tergum/internal/entities"
	"zerosrealm.xyz/tergum/internal/log"
	"zerosrealm.xyz/tergum/internal/restic"
)

var conf *config.Config
var resticExe *restic.Restic

var jobs = make(map[string]*restic.Job)
var logger *log.Logger

type jobError struct {
	JobID string `json:"job"`
	Error error  `json:"error"`
	Msg   []byte `json:"msg"`
}

var jobErrors = make(chan jobError, 100)

func updateHandler(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case update := <-resticExe.Updates:
			type jobProgress struct {
				Msg json.RawMessage `json:"msg"`
			}
			msg, err := json.Marshal(jobProgress{Msg: update.Msg})
			if err != nil {
				logger.WithFields("function", "updateHandler").Debug("update dump:", spew.Sdump(update))
				logger.WithFields("function", "updateHandler").Error("marshalling update error:", err)
				continue
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
		case jobErr := <-jobErrors:
			type jobUpdate struct {
				Msg   string `json:"msg"`
				Error string `json:"error"`
			}
			logger.WithFields("function", "updateHandler", "job", jobErr.JobID).Error("job error:", jobErr.Error)

			msg, err := json.Marshal(jobUpdate{Msg: string(jobErr.Msg), Error: jobErr.Error.Error()})
			if err != nil {
				logger.WithFields("function", "updateHandler").Debug("jobError dump:", spew.Sdump(jobErr))
				logger.WithFields("function", "updateHandler").Error("marshalling jobError error:", err)
				continue
			}

			req, err := http.NewRequest("POST", conf.Server+"/api/job/"+jobErr.JobID+"/error", bytes.NewReader(msg))
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

	var packet entities.JobPacket
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
		logger.WithFields("function", "handleConnection", "job", packet.ID, "source", c.RemoteAddr().String()).Debug("job PSK:", packet.Agent.PSK)
		logger.WithFields("function", "handleConnection", "job", packet.ID, "source", c.RemoteAddr().String()).Warn("PSK mismatch")
		return
	}

	switch packet.Type {
	case "backup":
		var job entities.BackupJob
		dec := gob.NewDecoder(bytes.NewReader(packet.Data))
		err := dec.Decode(&job)
		logger.WithFields("function", "handleConnection", "job", packet.ID).Debug("job data:", spew.Sdump(packet))

		if err != nil {
			logger.WithFields("function", "handleConnection", "job", packet.ID).Error("decoding job error:", err)
			return
		}

		logger.WithFields("function", "handleConnection", "job", packet.ID).Info("Starting job")
		out, err := resticExe.Backup(packet.Repo.Repo, job.Backup.Source, packet.Repo.Password, job.Backup.Exclude, packet.ID, packet.Repo.Settings...)
		if err != nil {
			jobErrors <- jobError{JobID: packet.ID, Error: err, Msg: out}
			logger.WithFields("function", "handleConnection", "job", packet.ID, "output", string(out)).Error("restic backup error:", err)
			return
		}

		logger.WithFields("function", "handleConnection", "job", packet.ID).Debug("output:", string(out))
	case "restore":
		var job entities.RestoreJob
		dec := gob.NewDecoder(bytes.NewReader(packet.Data))
		err := dec.Decode(&job)
		logger.WithFields("function", "handleConnection", "job", packet.ID).Debug("job data:", spew.Sdump(packet))

		if err != nil {
			logger.WithFields("function", "handleConnection", "job", packet.ID).Error("decoding job error:", err)
			return
		}

		logger.WithFields("function", "handleConnection", "job", packet.ID).Info("Starting job")
		out, err := resticExe.Restore(packet.Repo.Repo, packet.Repo.Password, job.Snapshot,
			job.Target, job.Include, job.Exclude, packet.Repo.Settings...)
		if err != nil {
			jobErrors <- jobError{JobID: packet.ID, Error: err, Msg: out}
			logger.WithFields("function", "handleConnection", "job", packet.ID, "output", string(out)).Error("restic backup error:", err)
			return
		}

		logger.WithFields("function", "handleConnection", "job", packet.ID).Debug("output:", string(out))
		delete(jobs, packet.ID)
	case "stop":
		var job entities.StopJob
		dec := gob.NewDecoder(bytes.NewReader(packet.Data))
		err := dec.Decode(&job)
		logger.WithFields("function", "handleConnection", "job", packet.ID).Debug("job data:", spew.Sdump(packet))

		if err != nil {
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

func registerAgent() error {
	if conf.Registration == "" {
		logger.WithFields("function", "registerAgent").Debug("No registration token, skipping.")
		return nil
	}

	logger.WithFields("function", "registerAgent").Debug("Registering agent")

	type registrationData struct {
		Hostname string `json:"hostname"`
		Port     int    `json:"port"`
		Token    string `json:"token"`
	}
	hostname, err := os.Hostname()
	if err != nil {
		return fmt.Errorf("registerAgent(): error getting hostname: %w", err)
	}

	msg, err := json.Marshal(registrationData{
		Hostname: hostname,
		Port:     conf.Listen.Port,
		Token:    conf.Registration,
	})
	if err != nil {
		return fmt.Errorf("registerAgent(): error marshalling registration data: %w", err)
	}

	req, err := http.NewRequest("POST", conf.Server+"/api/register", bytes.NewReader(msg))
	if err != nil {
		return fmt.Errorf("registerAgent(): error creating request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("registerAgent(): error sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			logger.WithFields("function", "registerAgent").Error("non-200 status:", resp.Status, "body read error:", err)
			return nil
		}
		logger.WithFields("function", "registerAgent").Info("non-200 status:", resp.Status, "body:", string(body))
		return fmt.Errorf("registerAgent(): non-200 status: %s", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("registerAgent(): error reading response body: %w", err)
	}

	var agent *entities.Agent
	err = json.Unmarshal(body, &agent)
	if err != nil {
		return fmt.Errorf("registerAgent(): error unmarshalling response body: %w", err)
	}

	conf.PSK = agent.PSK

	logger.WithFields("function", "registerAgent").Debug("Agent registered:", agent)

	return nil
}

func main() {
	config, err := config.Load()
	if err != nil {
		panic(err)
	}
	conf = config

	log, err := log.New(&conf.Log)
	if err != nil {
		logger.Error("error creating logger:", err)
		return
	}
	logger = log

	logger.Info("Starting agent")

	if conf.PSK == "" && conf.Registration == "" {
		logger.Info("no PSK defined - exiting")
		return
	}

	if conf.Restic == "" {
		logger.Info("no path to restic defined - exiting")
		return
	}

	if conf.Server == "" {
		logger.Info("no server defined - exiting")
		return
	}

	err = registerAgent()
	if err != nil {
		logger.Error("error registering agent:", err)
		return
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
