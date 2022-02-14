package main

import (
	"log"

	"zerosrealm.xyz/tergum/internal/server"
	"zerosrealm.xyz/tergum/internal/server/config"
	"zerosrealm.xyz/tergum/internal/server/service"
	"zerosrealm.xyz/tergum/internal/server/service/adapter/agent"
	"zerosrealm.xyz/tergum/internal/server/service/adapter/backup"
	"zerosrealm.xyz/tergum/internal/server/service/adapter/backupSubscribers"
	"zerosrealm.xyz/tergum/internal/server/service/adapter/forget"
	"zerosrealm.xyz/tergum/internal/server/service/adapter/job"
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

	var backupSubCache service.BackupSubscriberCache
	var backupSubStorage service.BackupSubscriberStorage

	var forgetCache service.ForgetCache
	var forgetStorage service.ForgetStorage

	var jobCache service.JobCache
	var jobStorage service.JobStorage

	switch conf.Database.Driver {
	case "memory":
		repoStorage = repo.NewMemoryStorage()
		agentStorage = agent.NewMemoryStorage()
		backupStorage = backup.NewMemoryStorage()
		backupSubStorage = backupSubscribers.NewMemoryStorage()
		forgetStorage = forget.NewMemoryStorage()
		jobStorage = job.NewMemoryStorage()
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

		backupSubSQL, err := backupSubscribers.NewSQLiteStorage(conf.Database.DataSourceName)
		if err != nil {
			log.Fatal(err)
		}
		defer backupSubSQL.Close()

		forgetSQL, err := forget.NewSQLiteStorage(conf.Database.DataSourceName)
		if err != nil {
			log.Fatal(err)
		}
		defer forgetSQL.Close()

		jobSQL, err := job.NewSQLiteStorage(conf.Database.DataSourceName)
		if err != nil {
			log.Fatal(err)
		}
		defer jobSQL.Close()

		repoStorage = repoSQL
		agentStorage = agentSQL
		backupStorage = backupSQL
		backupSubStorage = backupSubSQL
		forgetStorage = forgetSQL
		jobStorage = jobSQL
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
	backupSubSvc := service.NewBackupSubscriberService(&backupSubCache, &backupSubStorage)
	forgetSvc := service.NewForgetService(&forgetCache, &forgetStorage)
	jobSvc := service.NewJobService(&jobCache, &jobStorage)

	services := service.NewServices(repoSvc, agentSvc, backupSvc, backupSubSvc, forgetSvc, jobSvc)

	log.Println("starting server")
	server, err := server.New(conf, services)
	if err != nil {
		log.Fatal(err)
	}
	server.Start()
}
