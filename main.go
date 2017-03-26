package main

import (
	"os"
	"os/signal"
	"syscall"

	_ "github.com/mattn/go-sqlite3"
	"github.com/trtstm/budgetr/config"
	"github.com/trtstm/budgetr/db"
	"github.com/trtstm/budgetr/log"
)

func handleInterrupt(quit chan struct{}) {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		quit <- struct{}{}
	}()
}

func main() {
	if conf, err := config.NewConfigFromFile("./config.json"); err != nil {
		log.Fatalf("Failed read configuration file: %v", err)
	} else {
		config.Config = conf
	}

	var logLevel log.Level
	switch config.Config.LogLevel {
	case "debug":
		logLevel = log.DebugLevel
	case "info":
		logLevel = log.InfoLevel
	case "warn":
		logLevel = log.WarnLevel
	case "error":
		logLevel = log.ErrorLevel
	default:
		logLevel = log.InfoLevel
	}
	log.SetLevel(logLevel)

	if err := db.SetupConnection(db.SQLITE, config.Config.Database); err != nil {
		log.Fatalf("Failed to create connection to database: %v", err)
	}

	if err := db.SetupSchema(); err != nil {
		log.Fatalf("Failed to initialize database schema: %v", err)
	}

	quit := make(chan struct{})
	handleInterrupt(quit)

	go startAPI()

	<-quit
	log.Info("Goodbye.")
}
