package main

import (
	"os"
)

func startUpTasks(app *app) error {
	createDirectories(app)
	return nil
}

func createDirectories(app *app) {
	// create a `.tmp` directory if it doesn't exist
	if _, err := os.Stat(".tmp"); os.IsNotExist(err) {
		err = os.Mkdir(".tmp", 0755)
		if err != nil {
			app.logger.Error("failed to create .tmp directory", "error", err)
		}
	}
}

func (app *app) governor() error {

	startUpTasks(app)

	app.RemoveTorrents()
	app.checkSources()
	return nil
}
