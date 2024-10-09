package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/xml"
	"io"
	"net/http"

	"github.com/blackfyre/jacketted-transmission/assets/dtos"
	"github.com/blackfyre/jacketted-transmission/assets/errs"
)

func (app *app) ProcessSource(source dtos.JackettSource) error {
	resp, err := http.Get(source.RssUrl)
	if err != nil {
		app.logger.Error("failed to get RSS feed", "error", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		app.logger.Error("unexpected status code", "status", resp.StatusCode)
		return err
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

		err = app.AddTorrent(dtos.TorrentFile{
			Guid:       md5Hash(item.Guid),
			URL:        item.Link,
			DownloadTo: *source.TargetFolder,
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

// md5Hash takes a string input and returns its MD5 hash as a hexadecimal string.
// It uses the md5 package to create a new hash, writes the input string as bytes to the hash,
// and then encodes the resulting hash sum to a hexadecimal string.
func md5Hash(s string) string {
	h := md5.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

// checkSources verifies the sources defined in the application's configuration.
// It performs the following checks:
// 1. Ensures that there is at least one source defined in the configuration.
// 2. Ensures that the source index to run is within the bounds of the defined sources.
// 3. Processes the source at the specified index and logs any errors encountered during processing.
//
// Returns an error if any of the checks fail or if processing the source fails.
func (app *app) checkSources() error {

	if len(app.config.Jackett.Sources) == 0 {
		return &errs.NoJackettSourceError{Message: "no sources defined"}
	}

	if len(app.config.Jackett.Sources) < *app.sourceToRun {
		return &errs.JobIndexIsOverTheNumberOfSourcesError{Message: "job index is over the number of sources"}
	}

	for sourceIndex, source := range app.config.Jackett.Sources {
		if sourceIndex == *app.sourceToRun {
			err := app.ProcessSource(source)
			if err != nil {
				app.logger.Error("failed to process source", "error", err)
				return &errs.SourceProcessingError{Message: "failed to process source", Trace: []error{err}}
			}
		}
	}

	return nil
}
