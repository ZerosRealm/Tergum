package server

import (
	"context"
	"encoding/gob"
	"fmt"
	"log"
	"net"

	"zerosrealm.xyz/tergum/internal/types"
)

var jobQueue = make(chan types.Job, 100)

func enqueue(job types.Job) bool {
	select {
	case jobQueue <- job:
		return true
	default:
		return false
	}
}

func queueHandler(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Println("queueHandler canceled.")
			return

		case job := <-jobQueue:
			if ctx.Err() != nil {
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
