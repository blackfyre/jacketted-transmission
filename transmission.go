package main

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/blackfyre/jacketted-transmission/assets/dtos"
	"github.com/hekmon/transmissionrpc/v3"
)

func (app *app) createTransmissionClient(host string, port int, user string, password string) {
	endpoint, err := url.Parse(fmt.Sprintf("http://%s:%s@%s:%d/transmission/rpc", user, password, host, port))
	if err != nil {
		app.logger.Error("failed to parse transmission RPC endpoint", "error", err)
		panic(err)
	}

	transmissionClient, err := transmissionrpc.New(endpoint, nil)
	if err != nil {
		app.logger.Error("failed to create transmission RPC client", "error", err)
		panic(err)

	}

	ok, serverVersion, serverMinimumVersion, err := transmissionClient.RPCVersion(context.TODO())
	if err != nil {
		app.logger.Error("failed to get transmission RPC version", "error", err)
		panic(err)
	}
	if !ok {
		app.logger.Error("Wrong transmission RPC version, please update Transmission", "version", serverVersion, "minimum version", serverMinimumVersion)
		panic("Wrong transmission RPC version, please update Transmission")
	}

	app.logger.Info("transmission RPC version info", "supported", transmissionrpc.RPCVersion, "reported version", serverVersion, "minimum version", serverMinimumVersion)
	app.logger.Info("created transmission RPC client", "host", host, "port", port, "user", user)

	app.transmissionClient = transmissionClient
}

func (app *app) AddTorrent(tf dtos.TorrentFile) error {

	tracker, err := app.db.GetTracker(tf.Guid)

	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			tracker.Guid = ""
		} else {
			fmt.Fprintln(os.Stderr, err)
			return err
		}
	}

	if tracker.Guid != "" {
		app.logger.Info("torrent already added", "guid", tf.Guid)
	} else {

		paused := false

		torrent, err := app.transmissionClient.TorrentAdd(context.TODO(), transmissionrpc.TorrentAddPayload{
			Filename:    tf.URL,
			Paused:      &paused,
			DownloadDir: &tf.DownloadTo,
		})
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return err
		} else {
			err = app.db.CreateTracker(tf.Guid)

			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				return err
			}

			err = app.db.UpdateTracker(tf.Guid, "added", tf.Ratio, tf.SeedTime, *torrent.HashString)

			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				return err
			}
		}

		app.logger.Info("added torrent", "torrent", tf.Guid)
	}
	return nil
}

func (app *app) RemoveTorrents() error {
	list, err := app.transmissionClient.TorrentGetAll(context.TODO())

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

	if len(removalList) > 0 {
		err = app.transmissionClient.TorrentRemove(context.TODO(), transmissionrpc.TorrentRemovePayload{
			IDs:             removalList,
			DeleteLocalData: true,
		})

		if err != nil {
			app.logger.Error("failed to remove torrents", "error", err)
			return err
		}

		app.logger.Info("removed torrents", "number", len(removalList))
	} else {
		app.logger.Info("no torrents to remove")
	}

	return nil
}
