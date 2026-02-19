package app

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"

	"github.com/blackfyre/jacketted-transmission/internal/assets/dtos"
	"github.com/blackfyre/jacketted-transmission/internal/assets/errs"
)

func (app *App) processSource(source dtos.JackettSource) error {
	resp, err := http.Get(source.GetRssUrl())
	if err != nil {
		app.logger.Error("failed to get RSS feed", "error", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		app.logger.Error("unexpected status code", "status", resp.StatusCode)
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		app.logger.Error("failed to read RSS feed", "error", err)
		return err
	}

	var rssFeed dtos.JackettRss
	err = xml.Unmarshal(body, &rssFeed)
	if err != nil {
		app.logger.Error("failed to unmarshal RSS feed", "error", err)
		return err
	}

	for _, item := range rssFeed.Channel.Item {
		err = app.addTorrent(dtos.TorrentFile{
			Guid:       md5Hash(item.Guid),
			URL:        item.Link,
			DownloadTo: source.GetTargetFolder(),
			Ratio:      source.GetRatio(),
			SeedTime:   source.GetSeedMinutes(),
		})
		if err != nil {
			app.logger.Error("failed to add torrent", "error", err)
			return err
		}
	}

	return nil
}

func md5Hash(s string) string {
	h := md5.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

func (app *App) checkSources() error {
	if len(app.config.Jackett.Sources) == 0 {
		return &errs.NoJackettSourceError{Message: "no sources defined"}
	}

	if app.sourceToRun < 0 || len(app.config.Jackett.Sources) <= app.sourceToRun {
		return &errs.JobIndexIsOverTheNumberOfSourcesError{Message: "job index is over the number of sources"}
	}

	for sourceIndex, source := range app.config.Jackett.Sources {
		if sourceIndex == app.sourceToRun {
			err := app.processSource(source)
			if err != nil {
				app.logger.Error("failed to process source", "error", err)
				return &errs.SourceProcessingError{Message: "failed to process source", Trace: []error{err}}
			}
		}
	}

	return nil
}
