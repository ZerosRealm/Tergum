package server

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/gorilla/websocket"
	"github.com/rs/xid"

	"zerosrealm.xyz/tergum/internal/entities"
	"zerosrealm.xyz/tergum/internal/log"
	"zerosrealm.xyz/tergum/internal/server/service"
)

type Manager struct {
	ctx context.Context
	// jobs      []*entities.Job
	jobsMutex *sync.Mutex
	services  *service.Services

	log *log.Logger

	wsWrite       chan []byte
	jobQueue      chan entities.JobPacket
	wsConnections *map[string]*websocket.Conn
}

func NewManager(ctx context.Context, services *service.Services, logger *log.Logger, wsConns *map[string]*websocket.Conn) *Manager {
	return &Manager{
		ctx: ctx,
		// jobs:      make([]*entities.Job, 0),
		jobsMutex: &sync.Mutex{},
		services:  services,

		log: logger.WithFields("component", "manager"),

		wsWrite:       make(chan []byte, 100),
		jobQueue:      make(chan entities.JobPacket, 100),
		wsConnections: wsConns,
	}
}

func (man *Manager) Start() {
	go man.wsWriter()
	go man.queueHandler()
}

func (man *Manager) NewJob(packet *entities.JobPacket, typePacket interface{}) (*entities.Job, error) {
	man.jobsMutex.Lock()
	defer man.jobsMutex.Unlock()

	id := xid.New().String()
	packet.ID = id
	man.log.Debug("creating new job", id)

	buf := bytes.NewBuffer(nil)
	enc := gob.NewEncoder(buf)
	err := enc.Encode(typePacket)

	if err != nil {
		spew.Dump(typePacket)
		return nil, fmt.Errorf("manager.newJob: failed to encode job packet: %w", err)
	}
	packet.Data = buf.Bytes()

	if packet.Type == "backup" {
		backupPacket := typePacket.(*entities.BackupJob)

		// Check if it's a valid backup
		if backupPacket.Backup.Schedule == "" {
			return nil, fmt.Errorf("manager.newJob: job %s error: backup data is empty", id)
		}

		backups, err := man.services.BackupSvc.GetAll()
		if err != nil {
			return nil, fmt.Errorf("manager.newJob: job %s error: %w", id, err)
		}

		for _, backup := range backups {
			if backup.ID == backupPacket.Backup.ID {
				backup.LastRun = time.Now()
				man.services.BackupSvc.Update(backup)
				break
			}
		}
	}

	job := &entities.Job{
		ID:      id,
		Done:    false,
		Aborted: false,

		Packet:    packet,
		StartTime: time.Now(),
	}

	job, err = man.services.JobSvc.Create(job)
	if err != nil {
		return nil, fmt.Errorf("manager.newJob: job %s could not get created: %w", id, err)
	}

	ok := man.enqueue(*packet)
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

func (man *Manager) UpdateJobProgress(job *entities.Job, data []byte) {
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
		job.Done = true
		job.EndTime = time.Now()

	case "error":
		man.log.WithFields("job", job.ID).Warn("updateJobProgress: restic returned error", string(data))
		job.Aborted = true
	}
}

func (man *Manager) StopJob(job *entities.Job) error {
	packet := job.Packet
	packet.Type = "stop"

	buf := bytes.NewBuffer(nil)
	enc := gob.NewEncoder(buf)
	err := enc.Encode(entities.StopJob{
		ID: job.ID,
	})

	if err != nil {
		return fmt.Errorf("manager.stopJob: %w", err)
	}
	packet.Data = buf.Bytes()

	agentAddr, err := net.ResolveTCPAddr("tcp4", fmt.Sprintf("%s:%d", packet.Agent.IP, packet.Agent.Port))
	if err != nil {
		return fmt.Errorf("manager.stopJob: %w", err)
	}

	conn, err := net.DialTCP("tcp4", nil, agentAddr)
	if err != nil {
		return fmt.Errorf("manager.stopJob: %w", err)
	}
	defer conn.Close()

	enc = gob.NewEncoder(conn)
	enc.Encode(packet)
	spew.Dump(packet)

	man.log.WithFields("job", job.ID).Debug("successfully sent to", packet.Agent.Name)
	return nil
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
