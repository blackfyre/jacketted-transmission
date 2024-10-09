package errs

type NoJackettSourceError struct {
	Message string
}

func (e *NoJackettSourceError) Error() string {
	return e.Message
}

type JobIndexIsOverTheNumberOfSourcesError struct {
	Message string
}

func (e *JobIndexIsOverTheNumberOfSourcesError) Error() string {
	return e.Message
}

type TorrentAlreadyDownloadedError struct {
	Message string
}

func (e *TorrentAlreadyDownloadedError) Error() string {
	return e.Message
}

type SourceProcessingError struct {
	Message string
	Trace   []error
}

func (e *SourceProcessingError) Error() string {
	return e.Message
}
