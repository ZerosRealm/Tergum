package server

import (
	"fmt"

	"github.com/davecgh/go-spew/spew"
	"zerosrealm.xyz/tergum/internal/entity"
)

func (man *Manager) enqueue(job *entity.JobRequest) bool {
	select {
	case man.jobQueue <- job:
		return true
	default:
		return false
	}
}

func (man *Manager) queueHandler() {
	man.log.Debug("queueHandler: starting")
	for {
		select {
		case <-man.ctx.Done():
			man.log.Debug("queueHandler: canceled")
			return

		case job := <-man.jobQueue:
			if man.ctx.Err() != nil {
				return
			}
			man.log.WithFields("job", job.ID).Debug("Sending to", job.Agent.Name, "at", fmt.Sprintf("%s:%d", job.Agent.IP, job.Agent.Port))

			man.log.WithFields("job", job.ID).Debug("Request:", spew.Sdump(job))
			_, err := man.SendRequest(job, job.Agent)
			if err != nil {
				man.log.WithFields("job", job.ID).Error("Sending request returned error:", err)
				return
			}

			// agentAddr, err := net.ResolveTCPAddr("tcp4", fmt.Sprintf("%s:%d", job.Agent.IP, job.Agent.Port))
			// if err != nil {
			// 	man.log.WithFields("job", job.ID).Error("queueHandler: could not resolve agent address:", err)
			// 	man.jobAborted(job.ID)
			// 	continue
			// }

			// conn, err := net.DialTCP("tcp4", nil, agentAddr)
			// if err != nil {
			// 	man.log.WithFields("job", job.ID).Error("queueHandler: could not connect to agent:", err)
			// 	man.jobAborted(job.ID)
			// 	continue
			// }
			// defer conn.Close()

			// enc := gob.NewEncoder(conn)
			// err = enc.Encode(job)
			// if err != nil {
			// 	man.log.WithFields("job", job.ID).Error("queueHandler: could not encode job to connection:", err)
			// 	man.jobAborted(job.ID)
			// 	continue
			// }

			// man.log.WithFields("job", job.ID).Debug("successfully sent to", job.Agent.Name)

		}
	}
}
