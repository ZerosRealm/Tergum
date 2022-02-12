package main

import (
	"log"

	"zerosrealm.xyz/tergum/internal/server"
	"zerosrealm.xyz/tergum/internal/server/config"
	"zerosrealm.xyz/tergum/internal/server/service"
	"zerosrealm.xyz/tergum/internal/server/service/adapter/agent"
	"zerosrealm.xyz/tergum/internal/server/service/adapter/backup"
	"zerosrealm.xyz/tergum/internal/server/service/adapter/forget"
	"zerosrealm.xyz/tergum/internal/server/service/adapter/repo"
)

func main() {
	conf, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	var repoCache service.RepoCache
	var repoStorage service.RepoStorage
	var agentCache service.AgentCache
	var agentStorage service.AgentStorage
	var backupCache service.BackupCache
	var backupStorage service.BackupStorage
	var forgetCache service.ForgetCache
	var forgetStorage service.ForgetStorage

	switch conf.Database.Driver {
	case "memory":
		repoStorage = repo.NewMemoryStorage()
		agentStorage = agent.NewMemoryStorage()
		backupStorage = backup.NewMemoryStorage()
	case "postgres":
		log.Fatal("postgres storage not implemented")
	case "sqlite":
		repoSQL, err := repo.NewSQLiteStorage(conf.Database.DataSourceName)
		if err != nil {
			log.Fatal(err)
		}
		defer repoSQL.Close()

		agentSQL, err := agent.NewSQLiteStorage(conf.Database.DataSourceName)
		if err != nil {
			log.Fatal(err)
		}
		defer agentSQL.Close()

		backupSQL, err := backup.NewSQLiteStorage(conf.Database.DataSourceName)
		if err != nil {
			log.Fatal(err)
		}
		defer backupSQL.Close()

		forgetSQL, err := forget.NewSQLiteStorage(conf.Database.DataSourceName)
		if err != nil {
			log.Fatal(err)
		}
		defer forgetSQL.Close()

		repoStorage = repoSQL
		agentStorage = agentSQL
		backupStorage = backupSQL
		forgetStorage = forgetSQL
	default:
		log.Fatal("unsupported database driver")
	}

	switch conf.Cache {
	case "redis":
		// TODO: implement redis cache
		log.Fatal("redis cache not implemented")
	case "memory":
		repoCache = repo.NewMemoryCache()
		agentCache = agent.NewMemoryCache()
		backupCache = backup.NewMemoryCache()
		forgetCache = forget.NewMemoryCache()
	default:
		log.Println("continuing without cache")
	}

	repoSvc := service.NewRepoService(&repoCache, &repoStorage)
	agentSvc := service.NewAgentService(&agentCache, &agentStorage)
	backupSvc := service.NewBackupService(&backupCache, &backupStorage)
	forgetSvc := service.NewForgetService(&forgetCache, &forgetStorage)

	services := server.NewServices(repoSvc, agentSvc, backupSvc, forgetSvc)

	log.Println("starting server")
	server, err := server.New(conf, services)
	if err != nil {
		log.Fatal(err)
	}
	server.Start()
}
