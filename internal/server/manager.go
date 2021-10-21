package server

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/gorilla/websocket"
	"github.com/rs/xid"
	"zerosrealm.xyz/tergum/internal/types"
)

type Job struct {
	ID       string
	Packet   types.JobPacket
	Progress json.RawMessage

	StartTime time.Time
	EndTime   time.Time
}

type Manager struct {
	ctx       context.Context
	Jobs      []Job
	jobsMutex *sync.Mutex

	wsWrite  chan []byte
	jobQueue chan types.JobPacket
}

func NewManager(ctx context.Context) *Manager {
	return &Manager{
		ctx:       ctx,
		Jobs:      make([]Job, 0),
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

	job := Job{
		ID:        id,
		Packet:    *packet,
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

func (man *Manager) updateProgress(id string, msg []byte) {
	man.jobsMutex.Lock()
	defer man.jobsMutex.Unlock()

	index := -1
	for i, job := range man.Jobs {
		if job.ID == id {
			index = i
			break
		}
	}

	if index == -1 {
		return
	}

	man.Jobs[index].Progress = msg
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
