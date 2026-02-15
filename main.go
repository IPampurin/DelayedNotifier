package main

import (
	"log"

	"github.com/IPampurin/DelayedNotifier/pkg/configuration"
)

func main() {

	var err error

	cfg, err := configuration.ReadConfig()
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}

	/*
		err = db.InitDB()
		if err != nil {

		}

		err = server.Run()
		if err != nil {

		}
	*/
}
