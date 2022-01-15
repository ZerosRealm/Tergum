package server

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/gorilla/websocket"
	"github.com/rs/xid"
	"zerosrealm.xyz/tergum/internal/types"
)

type Manager struct {
	ctx       context.Context
	Jobs      []*types.Job
	jobsMutex *sync.Mutex

	wsWrite  chan []byte
	jobQueue chan types.JobPacket
}

func NewManager(ctx context.Context) *Manager {
	return &Manager{
		ctx:       ctx,
		Jobs:      make([]*types.Job, 0),
		jobsMutex: &sync.Mutex{},

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

	buf := bytes.NewBuffer(nil)
	enc := gob.NewEncoder(buf)
	err := enc.Encode(typePacket)

	if err != nil {
		spew.Dump(typePacket)
		return id, err
	}
	packet.Data = buf.Bytes()

	if packet.Type == "backup" {
		backupPacket := typePacket.(*types.BackupJob)

		// Check if it's a valid backup
		if backupPacket.Backup.Schedule == "" {
			return id, fmt.Errorf("backup data is empty")
		}

		for i, backup := range savedData.Backups {
			if backup.ID == backupPacket.Backup.ID {
				savedData.Backups[i].LastRun = time.Now()
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
		msg := fmt.Sprintf("job %s could not be enqueued\n", id)
		return id, fmt.Errorf(msg)
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
	// TODO: Add proper logging
	if err != nil {
		log.Println("job update:", err)
		return
	}

	switch msgType.MessageType {
	case "summary":
		job.Done = true
		job.EndTime = time.Now()
	case "error":
		// TODO: Add proper logging
		log.Println("job error:", string(data))
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
		return err
	}
	packet.Data = buf.Bytes()

	agentAddr, err := net.ResolveTCPAddr("tcp4", fmt.Sprintf("%s:%d", packet.Agent.IP, packet.Agent.Port))
	if err != nil {
		return err
	}

	conn, err := net.DialTCP("tcp4", nil, agentAddr)
	if err != nil {
		return err
	}
	defer conn.Close()

	enc = gob.NewEncoder(conn)
	enc.Encode(packet)
	spew.Dump(packet)

	log.Println(job.ID, "successfully sent to", packet.Agent.Name)
	return nil
}

func (man *Manager) wsWriter() {
	for {
		select {
		case <-man.ctx.Done():
			log.Println("wsWriter canceled.")
			return
		case msg := <-man.wsWrite:
			if man.ctx.Err() != nil {
				return
			}
			for _, c := range wsConnections {
				err := c.WriteMessage(websocket.TextMessage, msg)
				if err != nil {
					log.Println("wsWriter:", err)
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
