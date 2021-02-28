package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"math/rand"
	"net"
	"time"

	"zerosrealm.xyz/tergum/internal/agent/config"
	"zerosrealm.xyz/tergum/internal/restic"
	"zerosrealm.xyz/tergum/internal/types"
)

var conf config.Config
var resticExe *restic.Restic

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

	if data.Job.ID == "" {
		log.Println("invalid job")
		return
	}

	if data.Job.Agent.PSK != conf.PSK {
		log.Println("job PSK does not match")
		return
	}

	// spew.Dump(job)
	log.Println("running job", data.Job.ID)
	out, err := resticExe.Backup(data.Repo.Repo, data.Job.Backup.Source, data.Repo.Password, data.Repo.Settings...)
	if err != nil {
		log.Println("job error:", err, "out:", string(out))
		return
	}

	log.Println("job output:", string(out))
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

	l, err := net.Listen("tcp4", fmt.Sprintf("%s:%d", conf.Listen.IP, conf.Listen.Port))
	if err != nil {
		fmt.Println(err)
		return
	}
	defer l.Close()
	rand.Seed(time.Now().Unix())

	resticExe = restic.New(conf.Restic)

	for {
		c, err := l.Accept()
		if err != nil {
			fmt.Println(err)
			return
		}
		go handleConnection(c)
	}
}
