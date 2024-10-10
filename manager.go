package main

import (
	"os"
	"time"
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

	retentionDuration := time.Duration(app.config.App.TrackerRetentionDays) * 24 * time.Hour
	cutoffTime := time.Now().Add(-retentionDuration)

	app.db.DeleteTrackersOlderThan(cutoffTime.Format("2006-01-02 15:04:05"))
	return nil
}
