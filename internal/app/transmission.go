package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/blackfyre/jacketted-transmission/internal/assets/dtos"
	"github.com/hekmon/transmissionrpc/v3"
)

func (app *App) createTransmissionClient(host string, port int, user string, password string) error {
	endpoint, err := url.Parse(fmt.Sprintf("http://%s:%s@%s:%d/transmission/rpc", user, password, host, port))
	if err != nil {
		app.logger.Error("failed to parse transmission RPC endpoint", "error", err)
		return err
	}

	transmissionClient, err := transmissionrpc.New(endpoint, nil)
	if err != nil {
		app.logger.Error("failed to create transmission RPC client", "error", err)
		return err
	}

	ok, serverVersion, serverMinimumVersion, err := transmissionClient.RPCVersion(context.Background())
	if err != nil {
		if strings.Contains(err.Error(), "session-get") && strings.Contains(err.Error(), "cannot unmarshal number") {
			app.logger.Warn("failed to parse transmission session values while checking RPC version; continuing without strict version validation", "error", err)
		} else {
			app.logger.Error("failed to get transmission RPC version", "error", err)
			return err
		}
	}

	if err == nil && !ok {
		app.logger.Error("wrong transmission RPC version, please update Transmission", "version", serverVersion, "minimum version", serverMinimumVersion)
		return fmt.Errorf("wrong transmission RPC version, please update Transmission")
	}

	if err == nil {
		app.logger.Info("transmission RPC version info", "supported", transmissionrpc.RPCVersion, "reported version", serverVersion, "minimum version", serverMinimumVersion)
	}
	app.logger.Info("created transmission RPC client", "host", host, "port", port, "user", user)

	app.transmissionClient = transmissionClient
	return nil
}

func (app *App) addTorrent(tf dtos.TorrentFile) error {
	tracker, err := app.db.GetTracker(tf.Guid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			tracker.Guid = ""
		} else {
			app.logger.Error("failed to get tracker", "error", err)
			return err
		}
	}

	if tracker.Guid != "" {
		app.logger.Info("torrent already added", "guid", tf.Guid)
		return nil
	}

	paused := false
	torrent, err := app.transmissionClient.TorrentAdd(context.Background(), transmissionrpc.TorrentAddPayload{
		Filename:    tf.URL,
		Paused:      &paused,
		DownloadDir: &tf.DownloadTo,
	})
	if err != nil {
		app.logger.Error("failed to add torrent to transmission", "error", err)
		return err
	}

	if err = app.db.CreateTracker(tf.Guid); err != nil {
		app.logger.Error("failed to create tracker", "error", err)
		return err
	}

	if err = app.db.UpdateTracker(tf.Guid, "added", tf.Ratio, tf.SeedTime, *torrent.HashString); err != nil {
		app.logger.Error("failed to update tracker", "error", err)
		return err
	}

	app.logger.Info("added torrent", "torrent", tf.Guid)
	return nil
}

func (app *App) removeTorrents() error {
	list, err := app.transmissionClient.TorrentGetAll(context.Background())
	if err != nil {
		app.logger.Error("failed to get torrents from transmission", "error", err)
		return err
	}

	removalList := []int64{}
	for _, torrent := range list {
		tracker, err := app.db.GetTrackerByHash(*torrent.HashString)
		if err != nil {
			app.logger.Error(fmt.Sprintf("failed to get tracker by hash %s", *torrent.HashString), "error", err)
			app.logger.Info("Skipping torrent", "name", *torrent.Name)
			continue
		}

		minSeedDuration := time.Duration(tracker.TransmissionSeedTime) * time.Minute
		if tracker.TransmissionRatio <= *torrent.UploadRatio || *torrent.TimeSeeding > minSeedDuration {
			removalList = append(removalList, *torrent.ID)
		}
	}

	if len(removalList) == 0 {
		app.logger.Info("no torrents to remove")
		return nil
	}

	err = app.transmissionClient.TorrentRemove(context.Background(), transmissionrpc.TorrentRemovePayload{
		IDs:             removalList,
		DeleteLocalData: true,
	})
	if err != nil {
		app.logger.Error("failed to remove torrents", "error", err)
		return err
	}

	app.logger.Info("removed torrents", "number", len(removalList))
	return nil
}
