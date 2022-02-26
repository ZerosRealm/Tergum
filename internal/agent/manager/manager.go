package manager

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/egonelbre/antifreeze"
	"zerosrealm.xyz/tergum/internal/agent/config"
	"zerosrealm.xyz/tergum/internal/log"
	"zerosrealm.xyz/tergum/internal/restic"
)

func init() {
	antifreeze.SetFrozenLimit(1 * time.Minute)
}

type jobError struct {
	JobID string `json:"job"`
	Error error  `json:"error"`
	Msg   []byte `json:"msg"`
}

type Manager struct {
	ctx    context.Context
	log    *log.Logger
	restic *restic.Restic

	conf *config.Config

	jobMutex  sync.RWMutex
	jobs      map[string]*restic.Job
	jobErrors chan jobError
}

func New(ctx context.Context, conf *config.Config, resticExe *restic.Restic) (*Manager, error) {
	log, err := log.New(&conf.Log, "component", "manager")
	if err != nil {
		return nil, err
	}

	return &Manager{
		ctx:    ctx,
		log:    log,
		restic: resticExe,

		conf: conf,

		jobMutex:  sync.RWMutex{},
		jobs:      make(map[string]*restic.Job, 100),
		jobErrors: make(chan jobError),
	}, nil
}

type jobProgress struct {
	Msg json.RawMessage `json:"msg"`
}

type jobUpdate struct {
	Msg   string `json:"msg"`
	Error string `json:"error"`
}

func (man *Manager) UpdateHandler() {
	man.log.WithFields("function", "UpdateHandler").Debug("Starting")
	for {
		select {
		case <-man.ctx.Done():
			man.log.WithFields("function", "UpdateHandler").Debug("Context done, stopping")
			return
		case update := <-man.restic.Updates:
			msg, err := json.Marshal(jobProgress{Msg: update.Msg})
			if err != nil {
				man.log.WithFields("function", "UpdateHandler", "job", update.ID).Debug("update dump:", spew.Sdump(update))
				man.log.WithFields("function", "UpdateHandler", "job", update.ID).Error("marshalling update error:", err)
				continue
			}

			req, err := http.NewRequest("POST", man.conf.Server+"/api/job/"+update.ID+"/progress", bytes.NewReader(msg))
			if err != nil {
				man.log.WithFields("function", "UpdateHandler", "job", update.ID).Error("error:", err)
				continue
			}

			req.Header.Add("authorization", fmt.Sprintf("PSK %s", man.conf.PSK))

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				man.log.WithFields("function", "UpdateHandler", "job", update.ID).Error("marshalling update error:", err)
				continue
			}
			defer resp.Body.Close()

			if resp.StatusCode > 299 {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					man.log.WithFields("function", "UpdateHandler", "job", update.ID).Error("non-2XX status:", resp.Status, "body read error:", err)
					continue
				}
				man.log.WithFields("function", "UpdateHandler", "job", update.ID).Error("non-2XX status:", resp.Status, "body:", string(body))
			}

			man.log.WithFields("function", "UpdateHandler", "job", update.ID).Debug("Successfully sent update.")
		case job := <-man.restic.Jobs:
			man.jobMutex.Lock()

			if _, ok := man.jobs[job.ID]; ok {
				man.jobMutex.Unlock()
				continue
			}
			man.jobs[job.ID] = job
			man.jobMutex.Unlock()
		case jobErr := <-man.jobErrors:
			man.jobMutex.Lock()

			man.log.WithFields("function", "UpdateHandler", "job", jobErr.JobID).Error("job error:", jobErr.Error)

			msg, err := json.Marshal(jobUpdate{Msg: string(jobErr.Msg), Error: jobErr.Error.Error()})
			if err != nil {
				man.log.WithFields("function", "UpdateHandler").Debug("jobError dump:", spew.Sdump(jobErr))
				man.log.WithFields("function", "UpdateHandler").Error("marshalling jobError error:", err)
				man.jobMutex.Unlock()
				continue
			}

			req, err := http.NewRequest("POST", man.conf.Server+"/api/job/"+jobErr.JobID+"/error", bytes.NewReader(msg))
			if err != nil {
				man.log.WithFields("function", "UpdateHandler").Error("error:", err)
				man.jobMutex.Unlock()
				continue
			}

			req.Header.Add("authorization", fmt.Sprintf("PSK %s", man.conf.PSK))

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				man.log.WithFields("function", "UpdateHandler").Error("marshalling update error:", err)
				man.jobMutex.Unlock()
				continue
			}
			defer resp.Body.Close()

			if resp.StatusCode != 200 {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					man.log.WithFields("function", "UpdateHandler").Error("non-200 status:", resp.Status, "body read error:", err)
					man.jobMutex.Unlock()
					continue
				}
				man.log.WithFields("function", "UpdateHandler").Error("non-200 status:", resp.Status, "body:", string(body))
			}
		default:
			select {
			case <-man.ctx.Done():
				man.log.WithFields("function", "UpdateHandler").Debug("Context done, stopping")
				return
			default:
			}
		}
	}
}

func (man *Manager) Cancel() {
	man.jobMutex.Lock()
	defer man.jobMutex.Unlock()

	for _, job := range man.jobs {
		job.Cancel()
		delete(man.jobs, job.ID)
	}

	defer man.log.Close()
}
