package server

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/gorilla/websocket"
	"github.com/rs/xid"

	"zerosrealm.xyz/tergum/internal/log"
	"zerosrealm.xyz/tergum/internal/types"
)

type Manager struct {
	ctx       context.Context
	Jobs      []*types.Job
	jobsMutex *sync.Mutex
	services  *Services

	log *log.Logger

	wsWrite  chan []byte
	jobQueue chan types.JobPacket
}

func NewManager(ctx context.Context, services *Services, logger *log.Logger) *Manager {
	return &Manager{
		ctx:       ctx,
		Jobs:      make([]*types.Job, 0),
		jobsMutex: &sync.Mutex{},
		services:  services,

		log: logger.WithFields("component", "manager"),

		wsWrite:  make(chan []byte, 100),
		jobQueue: make(chan types.JobPacket, 100),
	}
}

func (man *Manager) Start() {
	go man.wsWriter()
	go man.queueHandler()
}

func (man *Manager) NewJob(packet *types.JobPacket, typePacket interface{}) (string, error) {
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
		return id, fmt.Errorf("manager.newJob: failed to encode job packet: %w", err)
	}
	packet.Data = buf.Bytes()

	if packet.Type == "backup" {
		backupPacket := typePacket.(*types.BackupJob)

		// Check if it's a valid backup
		if backupPacket.Backup.Schedule == "" {
			return id, fmt.Errorf("manager.newJob: job %s error: backup data is empty", id)
		}

		backups, err := man.services.backupSvc.GetAll()
		if err != nil {
			return id, fmt.Errorf("manager.newJob: job %s error: %w", id, err)
		}

		for _, backup := range backups {
			if backup.ID == backupPacket.Backup.ID {
				backup.LastRun = time.Now()
				man.services.backupSvc.Update(backup)
				break
			}
		}
	}

	job := &types.Job{
		ID:      id,
		Done:    false,
		Aborted: false,

		Packet:    packet,
		StartTime: time.Now(),
	}
	man.Jobs = append(man.Jobs, job)

	ok := man.enqueue(*packet)
	if !ok {
		return id, fmt.Errorf("manager.newJob: job %s could not be enqueued", id)
	}

	return id, nil
}

func (man *Manager) updateJobProgress(job *types.Job, data []byte) {
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

func (man *Manager) getJob(id string) *types.Job {
	for _, job := range man.Jobs {
		if strings.EqualFold(job.ID, id) {
			return job
		}
	}

	return nil
}

func (man *Manager) stopJob(job *types.Job) error {
	packet := job.Packet
	packet.Type = "stop"

	buf := bytes.NewBuffer(nil)
	enc := gob.NewEncoder(buf)
	err := enc.Encode(types.StopJob{
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
			for _, c := range wsConnections {
				err := c.WriteMessage(websocket.TextMessage, msg)
				if err != nil {
					man.log.Error("wsWriter: ", err)
					continue
				}
			}
		default:
		}
	}
}

func (man *Manager) WriteWS(data []byte) {
	select {
	case man.wsWrite <- data:
	default:
	}
}
