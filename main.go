package main

import (
	"github.com/IPampurin/DelayedNotifier/pkg/db"
	"github.com/IPampurin/DelayedNotifier/pkg/server"
)

func main() {

	var err error

	err = db.InitDB()
	if err != nil {

	}

	err = server.Run()
	if err != nil {

	}
}
