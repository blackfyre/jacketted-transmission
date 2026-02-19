package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"runtime/debug"

	"github.com/blackfyre/jacketted-transmission/database"
	appsvc "github.com/blackfyre/jacketted-transmission/internal/app"
	"github.com/blackfyre/jacketted-transmission/internal/assets/dtos"
	"github.com/blackfyre/jacketted-transmission/internal/version"
	"gopkg.in/yaml.v2"
)

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
		return err
	}

	if config.Jackett.Sources == nil {
		return fmt.Errorf("no sources defined")
	}

	iCfg.DB.Dsn = "db.sqlite"
	iCfg.DB.Automigrate = true

	fmt.Printf("version: %s\n", version.Get())

	db, err := database.New(iCfg.DB.Dsn, iCfg.DB.Automigrate)
	if err != nil {
		return err
	}
	defer db.Close()

	sourceToRun := flag.Int("source", 0, "source to run")
	flag.Parse()

	app, err := appsvc.New(config, iCfg, db, logger, *sourceToRun)
	if err != nil {
		return err
	}

	return app.Run()
}
