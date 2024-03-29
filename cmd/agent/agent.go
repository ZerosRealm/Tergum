package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"time"

	"github.com/davecgh/go-spew/spew"
	"zerosrealm.xyz/tergum/internal/agent/config"
	"zerosrealm.xyz/tergum/internal/restic"
	"zerosrealm.xyz/tergum/internal/types"
)

var conf config.Config
var resticExe *restic.Restic

var updates = make(chan []byte, 100)

func updateHandler() {
	for {
		select {
		case msg := <-updates:
			log.Println("received update!")

			req, err := http.NewRequest("POST", conf.Server+"update", bytes.NewReader(msg))
			if err != nil {
				log.Println(err)
				continue
			}

			req.Header.Add("authorization", fmt.Sprintf("PSK %s", conf.PSK))

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				log.Println(err)
				continue
			}
			log.Println("Sent update, status:", resp.StatusCode)
		default:
		}
	}
}

func handleConnection(c net.Conn) {
	log.Printf("serving %s\n", c.RemoteAddr().String())
	defer c.Close()

	var data types.JobPacket
	dec := gob.NewDecoder(c)
	err := dec.Decode(&data)

	if err != nil {
		log.Println("error:", err)
		return
	}

	if data.ID == "" {
		log.Println("invalid job")
		return
	}

	if data.Agent.PSK != conf.PSK {
		log.Println("job PSK does not match")
		return
	}

	switch data.Type {
	case "backup":
		var job types.BackupJob
		dec := gob.NewDecoder(bytes.NewReader(data.Job))
		err := dec.Decode(&job)

		if err != nil {
			spew.Dump(data.Job)
			panic(err)
		}

		log.Println("running job", data.ID)
		out, err := resticExe.Backup(data.Repo.Repo, job.Backup.Source, data.Repo.Password, job.Backup.Exclude, data.ID, updates, data.Repo.Settings...)
		if err != nil {
			log.Println("job error:", err, "out:", string(out))
			return
		}

		log.Println("job output:", string(out))
	case "restore":
		var job types.RestoreJob
		dec := gob.NewDecoder(bytes.NewReader(data.Job))
		err := dec.Decode(&job)

		if err != nil {
			spew.Dump(data.Job)
			panic(err)
		}

		log.Println("running job", data.ID)
		out, err := resticExe.Restore(data.Repo.Repo, data.Repo.Password, job.Snapshot,
			job.Target, job.Include, job.Exclude, data.Repo.Settings...)
		if err != nil {
			log.Println("job error:", err, "out:", string(out))
			return
		}

		log.Println("job output:", string(out))

	default:
		log.Println("job", data.ID, "has unknown job type")
	}
}

func main() {
	log.Println("starting agent")
	conf = config.Load()

	if conf.PSK == "" {
		log.Fatal("no PSK defined - exiting")
	}

	if conf.Restic == "" {
		log.Fatal("no path to restic defined - exiting")
	}

	if conf.Server == "" {
		log.Fatal("no server defined - exiting")
	}

	l, err := net.Listen("tcp4", fmt.Sprintf("%s:%d", conf.Listen.IP, conf.Listen.Port))
	if err != nil {
		fmt.Println(err)
		return
	}
	defer l.Close()
	rand.Seed(time.Now().Unix())

	resticExe = restic.New(conf.Restic)

	go updateHandler()

	for {
		c, err := l.Accept()
		if err != nil {
			fmt.Println(err)
			return
		}
		go handleConnection(c)
	}
}
