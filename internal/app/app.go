package app

import (
	"log/slog"
	"sync"

	"github.com/blackfyre/jacketted-transmission/database"
	"github.com/blackfyre/jacketted-transmission/internal/assets/dtos"
	"github.com/hekmon/transmissionrpc/v3"
)

type App struct {
	config             dtos.Config
	ic                 dtos.InternalConfig
	db                 *database.DB
	logger             *slog.Logger
	wg                 sync.WaitGroup
	transmissionClient *transmissionrpc.Client
	sourceToRun        int
}

func New(config dtos.Config, ic dtos.InternalConfig, db *database.DB, logger *slog.Logger, sourceToRun int) (*App, error) {
	app := &App{
		ic:          ic,
		config:      config,
		db:          db,
		logger:      logger,
		sourceToRun: sourceToRun,
	}

	if err := app.createTransmissionClient(config.Transmission.Host, config.Transmission.Port, config.Transmission.User, config.Transmission.Password); err != nil {
		return nil, err
	}

	return app, nil
}

func (app *App) Run() error {
	return app.governor()
}
