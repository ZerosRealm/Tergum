package server

import (
	"encoding/gob"
	"fmt"
	"log"
	"net"

	"zerosrealm.xyz/tergum/internal/types"
)

func (man *Manager) enqueue(job types.JobPacket) bool {
	select {
	case man.jobQueue <- job:
		return true
	default:
		return false
	}
}

func (man *Manager) queueHandler() {
	for {
		select {
		case <-man.ctx.Done():
			log.Println("queueHandler canceled.")
			return

		case job := <-man.jobQueue:
			if man.ctx.Err() != nil {
				return
			}
			log.Println("sending job", job.ID, "to", job.Agent.Name)

			agentAddr, err := net.ResolveTCPAddr("tcp4", fmt.Sprintf("%s:%d", job.Agent.IP, job.Agent.Port))
			if err != nil {
				log.Println(err)
				continue
			}

			conn, err := net.DialTCP("tcp4", nil, agentAddr)
			if err != nil {
				log.Println(err)
				continue
			}
			defer conn.Close()

			enc := gob.NewEncoder(conn)
			enc.Encode(job)

			log.Println(job.ID, "successfully sent to", job.Agent.Name)
		}
	}
}
