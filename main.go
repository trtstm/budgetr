package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	_ "github.com/mattn/go-sqlite3"
	"github.com/trtstm/budgetr/log"
)

// Config contains the config for the application.
type Config struct {
	Hostname string `json:"hostname"`
	Port     uint32 `json:"port"`
	Database string `json:"database"`
	LogLevel string `json:"log_level"`
}

var config = Config{}

func loadConfig() {
	raw, err := ioutil.ReadFile("./config.json")
	if err != nil {
		log.WithFields(log.Fields{"reason": err.Error()}).Fatal("Could not load config.json.")
	} else {
		err = json.Unmarshal(raw, &config)
		if err != nil {
			log.WithFields(log.Fields{"reason": err.Error()}).Fatal("Could not parse config.json.")
		}
	}

	config.Hostname = strings.TrimSpace(config.Hostname)

	if len(config.Hostname) == 0 {
		config.Hostname = "127.0.0.1"
	}

	if config.Port == 0 {
		config.Port = 8080
	}

	if len(config.Database) == 0 {
		config.Database = "file::memory:?mode=memory&cache=shared"
	}

	config.LogLevel = strings.ToLower(config.LogLevel)

	switch config.LogLevel {
	case "debug":
	case "info":
	case "warn":
	case "error":
	default:
		config.LogLevel = "info"
	}

}

var db *sqlx.DB

func createSchema() {
	schema, err := ioutil.ReadFile("./schema.sql")
	if err != nil {
		log.WithFields(log.Fields{"reason": err.Error()}).Fatal("Could not load schema.sql")
	}

	_, err = db.Exec(string(schema))
	if err != nil {
		log.WithFields(log.Fields{"reason": err.Error()}).Fatal("Could not create schema.")
	}

	log.Info("Schema created.")
}

func loadDatabase() {
	var err error
	db, err = sqlx.Open("sqlite3", config.Database)
	if err != nil {
		log.WithFields(log.Fields{"reason": err.Error()}).Fatal("Could not open connection to database.")
	}

	createSchema()
	log.WithFields(log.Fields{"database": config.Database}).Info("Database initialized.")
}

func handleInterrupt(quit chan struct{}) {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		quit <- struct{}{}
	}()
}

func main() {
	loadConfig()
	log.SetLevel(log.ToLogLevel(config.LogLevel))

	loadDatabase()

	e := echo.New()
	e.Use(middleware.CORSWithConfig(middleware.DefaultCORSConfig))

	e.Static("/", "./web/dist")

	api := e.Group("/api")

	api.GET("/expenditures", indexExpenditure)
	api.POST("/expenditures", createExpenditure)

	quit := make(chan struct{})
	handleInterrupt(quit)

	go func() {
		log.Error(e.Start(config.Hostname + ":" + strconv.Itoa(int(config.Port))))
	}()

	<-quit
	log.Info("Goodbye.")
}
