package main

import (
	"log"
	"os"

	"bot/internal/clients/tgclient"
	"bot/internal/config"
	"bot/internal/consumer/eventconsumer"
	"bot/internal/listener"
	"bot/pkg/sqlite/audiostorage"
	"bot/pkg/sqlite/userstorage"
)

const (
	batchSize = 100
)

func main() {
	log.Printf("service started")

	cfg, err := config.Init()
	if err != nil {
		log.Fatal("can't init config: ", err)
	}

	userStorage, err := userstorage.New(cfg.StoragePath)
	if err != nil {
		log.Fatal("can't connect to user storage")
	}

	audioStorage, err := audiostorage.New(cfg.StoragePath, userStorage)
	if err != nil {
		log.Fatal("can't connect to audio storage: ", err)
	}

	if err := audioStorage.Init(); err != nil {
		log.Fatal("can't init audio storage: ", err)
	}

	if err := userStorage.Init(); err != nil {
		log.Fatal("can't init user storage: ", err)
	}

	tgclient := tgclient.New(cfg.Hostname, os.Getenv("TOKEN"))

	if err := tgclient.SetCommandsList(); err != nil {
		log.Fatal("can't set commands list", err)
	}

	listener := listener.New(tgclient, audioStorage, userStorage)

	consumer := eventconsumer.New(*listener, batchSize)

	if err := consumer.Start(); err != nil {
		log.Fatal("service stopped")
	}
}
