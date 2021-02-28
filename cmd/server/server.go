package main

import (
	"log"

	"zerosrealm.xyz/tergum/internal/server"
	"zerosrealm.xyz/tergum/internal/server/config"
)

func main() {
	log.Println("starting server")
	conf := config.Load()

	server.StartServer(&conf)
}
