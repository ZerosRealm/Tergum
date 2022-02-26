package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/xid"

	agentRequest "zerosrealm.xyz/tergum/internal/agent/api/request"
	"zerosrealm.xyz/tergum/internal/entity"
	"zerosrealm.xyz/tergum/internal/log"
	"zerosrealm.xyz/tergum/internal/server/service"
)

type Manager struct {
	ctx context.Context
	// jobs      []*entity.Job
	jobsMutex *sync.Mutex
	services  *service.Services

	log *log.Logger

	wsWrite       chan []byte
	jobQueue      chan *entity.JobRequest
	wsConnections *map[string]*websocket.Conn
}

func NewManager(ctx context.Context, services *service.Services, logger *log.Logger, wsConns *map[string]*websocket.Conn) *Manager {
	return &Manager{
		ctx: ctx,
		// jobs:      make([]*entity.Job, 0),
		jobsMutex: &sync.Mutex{},
		services:  services,

		log: logger.WithFields("component", "manager"),

		wsWrite:       make(chan []byte, 100),
		jobQueue:      make(chan *entity.JobRequest, 100),
		wsConnections: wsConns,
	}
}

func (man *Manager) Start() {
	go man.wsWriter()
	go man.queueHandler()
}

func (man *Manager) NewJob(jobRequest *entity.JobRequest) (*entity.Job, error) {
	man.jobsMutex.Lock()
	defer man.jobsMutex.Unlock()

	id := xid.New().String()
	jobRequest.ID = id
	man.log.Debug("creating new job", id)

	job := &entity.Job{
		ID:        id,
		Done:      false,
		Aborted:   false,
		Progress:  json.RawMessage([]byte(`{}`)),
		StartTime: time.Now(),
		Request:   jobRequest,
	}

	switch jobRequest.Type {
	case "backup":
		req := jobRequest.Data.(*agentRequest.Backup)

		if req.Backup == nil || req.Backup.Source == "" {
			return nil, fmt.Errorf("manager.newJob: backup packet is invalid")
		}

		req.Job.ID = id

		backup, err := man.services.BackupSvc.Get([]byte(strconv.Itoa(req.Backup.ID)))
		if err != nil {
			return nil, fmt.Errorf("manager.newJob: job %s could not get backup %d error: %w", id, req.Backup.ID, err)
		}

		backup.LastRun = time.Now()
		man.services.BackupSvc.Update(backup)

		jobRequest.Data = req
	case "stop":
		req := jobRequest.Data.(*agentRequest.Stop)
		req.Job.ID = id
		jobRequest.Data = req
	case "restore":
		req := jobRequest.Data.(*agentRequest.Restore)
		req.Job.ID = id
		jobRequest.Data = req
	default:
		return nil, fmt.Errorf("manager.newJob: unknown job type %s", jobRequest.Type)
	}

	job, err := man.services.JobSvc.Create(job)
	if err != nil {
		return nil, fmt.Errorf("manager.newJob: job %s could not be created: %w", id, err)
	}

	ok := man.enqueue(jobRequest)
	if !ok {
		job.Aborted = true

		_, updateErr := man.services.JobSvc.Update(job)
		if updateErr != nil {
			man.log.WithFields("job", job.ID).Error("manager.newJob: failed to enqueue job, and could not update to aborted", updateErr)
		}

		return nil, fmt.Errorf("manager.newJob: job %s could not be enqueued", id)
	}

	return job, nil
}

func (man *Manager) UpdateJobProgress(job *entity.Job, data []byte) {
	man.jobsMutex.Lock()
	defer man.jobsMutex.Unlock()
	job.Progress = json.RawMessage(data)

	var msgType struct {
		MessageType string `json:"message_type"`
	}
	err := json.Unmarshal(data, &msgType)
	if err != nil {
		man.log.WithFields("job", job.ID).Error("updateJobProgress: error unmarshalling data", err)
		return
	}

	switch msgType.MessageType {
	case "summary":
		man.log.WithFields("job", job.ID).Debug("updateJobProgress: job done")
		man.jobDone(job)

	case "error":
		man.log.WithFields("job", job.ID).Warn("updateJobProgress: restic returned error", string(data))
		man.jobAborted(job.ID)
	}
}

func (man *Manager) jobDone(job *entity.Job) error {
	job.Done = true
	job.EndTime = time.Now()

	_, err := man.services.JobSvc.Update(job)
	if err != nil {
		return fmt.Errorf("jobDone: could not update job: %w", err)
	}

	return nil
}

func (man *Manager) jobAborted(jobID string) error {
	job, err := man.services.JobSvc.Get([]byte(jobID))
	if err != nil {
		return fmt.Errorf("abortJob: could not get job: %w", err)
	}

	if job == nil {
		man.log.WithFields("job", jobID).Debug("abortJob: no job found with that ID.")
		return nil
	}

	job.Aborted = true

	_, err = man.services.JobSvc.Update(job)
	if err != nil {
		return fmt.Errorf("abortJob: could not update job: %w", err)
	}

	return nil
}

type wsToast struct {
	Type  string `json:"type"`
	Error error  `json:"error"`
	Msg   string `json:"msg"`
}

type errorResponse struct {
	Code    int    `json:"code"`
	Error   string `json:"error"`
	Message string `json:"message"`
}

func (man *Manager) WriteErrorWS(err error, msg string) {
	resp := wsToast{
		Type:  "error",
		Error: err,
		Msg:   msg,
	}

	errorJSON, err := json.Marshal(resp)
	if err != nil {
		man.log.Error("manager.SendError: could not marshal error:", err)
		return
	}

	man.WriteWS([]byte(errorJSON))
}

func (man *Manager) SendRequest(job *entity.JobRequest, agent *entity.Agent) ([]byte, error) {
	msg, err := json.Marshal(job.Data)
	if err != nil {
		return nil, fmt.Errorf("manager.sendRequest: error marshalling agent stop request: %w", err)
	}

	// TODO: Switch over to API request package.
	var endpoint string
	var method string
	switch job.Type {
	case "backup":
		endpoint = "/backup"
		method = "POST"
	case "restore":
		endpoint = "/snapshot/restore"
		method = "POST"
	case "stop":
		endpoint = "/stop"
		method = "POST"
	case "list":
		endpoint = "/snapshot/list"
		method = "POST"
	case "forget":
		endpoint = "/snapshot/forget"
		method = "POST"
	case "deletesnapshot":
		endpoint = "/snapshot"
		method = "DELETE"
	case "getsnapshots":
		endpoint = "/snapshot"
		method = "POST"
	default:
		return nil, fmt.Errorf("manager.sendRequest: unknown job type %s", job.Type)
	}

	// TODO: Change to HTTPS when we have a proper TLS support.
	req, err := http.NewRequest(method, "http://"+path.Join(fmt.Sprintf("%s:%d", agent.IP, agent.Port), "/api/"+endpoint), bytes.NewReader(msg))
	if err != nil {
		return nil, fmt.Errorf("manager.sendRequest: error creating request: %w", err)
	}
	req.Header.Set("X-PSK", agent.PSK)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("manager.sendRequest: error sending request: %w", err)
	}
	defer resp.Body.Close()

	man.log.WithFields("job", job.ID).Debug("successfully sent to", agent.Name)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		err = fmt.Errorf("manager.sendRequest: error reading response body: %w", err)
		man.log.WithFields("job", job.ID).Error(err)
		return nil, err
	}
	man.log.WithFields("job", job.ID).Debug("manager.sendRequest: status:", resp.Status, "body:", string(body))

	// If we got an error back from the agent, return this error.
	if resp.StatusCode > 299 {
		var errResp errorResponse
		err := json.Unmarshal(body, &errResp)
		if err != nil {
			return nil, fmt.Errorf("manager.sendRequest: error unmarshalling error response: %w", err)
		}
		man.log.WithFields("job", job.ID).Warn("non-2XX status:", resp.Status, "error:", errResp.Error)
		return nil, fmt.Errorf(errResp.Error)
	}

	return body, nil
}

func (man *Manager) wsWriter() {
	for {
		select {
		case <-man.ctx.Done():
			man.log.Debug("wsWriter canceled")
			return
		case msg := <-man.wsWrite:
			if man.ctx.Err() != nil {
				man.log.Debug("wsWriter canceled")
				return
			}
			for _, c := range *man.wsConnections {
				err := c.WriteMessage(websocket.TextMessage, msg)
				if err != nil {
					man.log.Error("wsWriter: ", err)
					continue
				}
			}
		default:
			if man.ctx.Err() != nil {
				man.log.Debug("wsWriter canceled")
				return
			}
		}
	}
}

func (man *Manager) WriteWS(data []byte) {
	select {
	case man.wsWrite <- data:
	default:
	}
}
