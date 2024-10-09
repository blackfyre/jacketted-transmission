package main

import (
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"runtime/debug"
	"sync"

	"github.com/blackfyre/jacketted-transmission/assets/dtos"
	"github.com/blackfyre/jacketted-transmission/database"
	"github.com/blackfyre/jacketted-transmission/version"
	"github.com/hekmon/transmissionrpc/v3"
	"gopkg.in/yaml.v2"
)

type app struct {
	config             dtos.Config
	ic                 dtos.InternalConfig
	db                 *database.DB
	logger             *slog.Logger
	wg                 sync.WaitGroup
	transmissionClient *transmissionrpc.Client
	sourceToRun        *int
}

func main() {

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	err := run(logger)
	if err != nil {
		trace := string(debug.Stack())
		logger.Error(err.Error(), "trace", trace)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	var iCfg dtos.InternalConfig
	var config dtos.Config

	data, err := os.ReadFile("config.yml")
	if err != nil {
		logger.Error("failed to read config file", "error", err)
		logger.Info("creating default config file")

		// copy config.example.yml to config.yml
		data, err = os.ReadFile("config.example.yml")
		if err != nil {
			return err
		}

		err = os.WriteFile("config.yml", data, 0644)
		if err != nil {
			return err
		}

		logger.Info("default config file created, please edit config.yml and run the program again")
		return nil
	}

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	if config.Jackett.Sources == nil {
		log.Fatalf("error: no sources defined")
	}

	iCfg.DB.Dsn = "db.sqlite"
	iCfg.DB.Automigrate = true

	fmt.Printf("version: %s\n", version.Get())

	db, err := database.New(iCfg.DB.Dsn, iCfg.DB.Automigrate)
	if err != nil {
		return err
	}
	defer db.Close()

	app := &app{
		ic:     iCfg,
		config: config,
		db:     db,
		logger: logger,
	}

	app.createTransmissionClient(config.Transmission.Host, config.Transmission.Port, config.Transmission.User, config.Transmission.Password)

	app.sourceToRun = flag.Int("source", 0, "source to run")
	flag.Parse()

	fmt.Printf("sourceToRun: %d\n", *app.sourceToRun)

	return app.governor()
}
