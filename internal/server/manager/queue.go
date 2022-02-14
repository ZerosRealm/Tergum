package server

import (
	"encoding/gob"
	"fmt"
	"net"

	"zerosrealm.xyz/tergum/internal/entities"
)

func (man *Manager) enqueue(job entities.JobPacket) bool {
	select {
	case man.jobQueue <- job:
		return true
	default:
		return false
	}
}

func (man *Manager) queueHandler() {
	man.log.Debug("Starting queueHandler")
	for {
		select {
		case <-man.ctx.Done():
			man.log.Debug("queueHandler: canceled")
			return

		case job := <-man.jobQueue:
			if man.ctx.Err() != nil {
				return
			}
			man.log.WithFields("job", job.ID).Debug("sending to", job.Agent.Name, "at", fmt.Sprintf("%s:%d", job.Agent.IP, job.Agent.Port))

			agentAddr, err := net.ResolveTCPAddr("tcp4", fmt.Sprintf("%s:%d", job.Agent.IP, job.Agent.Port))
			if err != nil {
				man.log.WithFields("job", job.ID).Error("queueHandler:", err)
				continue
			}

			conn, err := net.DialTCP("tcp4", nil, agentAddr)
			if err != nil {
				man.log.WithFields("job", job.ID).Error("queueHandler:", err)
				continue
			}
			defer conn.Close()

			enc := gob.NewEncoder(conn)
			err = enc.Encode(job)
			if err != nil {
				man.log.WithFields("job", job.ID).Error("queueHandler:", err)
			}

			man.log.WithFields("job", job.ID).Debug("successfully sent to", job.Agent.Name)
		}
	}
}
