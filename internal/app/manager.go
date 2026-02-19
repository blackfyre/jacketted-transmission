package app

import (
	"os"
	"time"
)

func startupTasks(app *App) error {
	createDirectories(app)
	return nil
}

func createDirectories(app *App) {
	if _, err := os.Stat(".tmp"); os.IsNotExist(err) {
		err = os.Mkdir(".tmp", 0755)
		if err != nil {
			app.logger.Error("failed to create .tmp directory", "error", err)
		}
	}
}

func (app *App) governor() error {
	if err := startupTasks(app); err != nil {
		return err
	}

	if err := app.removeTorrents(); err != nil {
		return err
	}

	if err := app.checkSources(); err != nil {
		return err
	}

	retentionDuration := time.Duration(app.config.App.TrackerRetentionDays) * 24 * time.Hour
	cutoffTime := time.Now().Add(-retentionDuration)

	return app.db.DeleteTrackersOlderThan(cutoffTime.Format("2006-01-02 15:04:05"))
}
