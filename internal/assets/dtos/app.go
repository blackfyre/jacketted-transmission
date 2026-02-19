package dtos

type Config struct {
	Jackett      JackettConfig      `yaml:"jackett"`
	Transmission TransmissionConfig `yaml:"transmission"`
	App          struct {
		LogRetentionDays     int `yaml:"log_retention_days"`
		TrackerRetentionDays int `yaml:"tracker_retention_days"`
	} `yaml:"app"`
}

type InternalConfig struct {
	DB struct {
		Dsn         string
		Automigrate bool
	}
}

type TorrentFile struct {
	Guid       string
	URL        *string
	DownloadTo string
	Ratio      float64
	SeedTime   int
}
