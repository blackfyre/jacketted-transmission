package dtos

import "encoding/xml"

type JackettConfig struct {
	Sources []JackettSource `yaml:"sources"`
}

type JackettSource struct {
	RssUrl       string   `yaml:"rss_url" json:"rss_url"`
	Ratio        *float64 `yaml:"ratio,omitempty" json:"ratio,omitempty"`
	TargetFolder *string  `yaml:"target_folder,omitempty" json:"target_folder,omitempty"`
	SeedMinutes  *int     `yaml:"seed_minutes,omitempty" json:"seed_minutes,omitempty"`
}

type JackettRss struct {
	XMLName xml.Name `xml:"rss"`
	Text    string   `xml:",chardata"`
	Version string   `xml:"version,attr"`
	Atom    string   `xml:"atom,attr"`
	Torznab string   `xml:"torznab,attr"`
	Channel struct {
		Text string `xml:",chardata"`
		Link struct {
			Text string `xml:",chardata"`
			Href string `xml:"href,attr"`
			Rel  string `xml:"rel,attr"`
			Type string `xml:"type,attr"`
		} `xml:"link"`
		Title       string `xml:"title"`
		Description string `xml:"description"`
		Language    string `xml:"language"`
		Category    string `xml:"category"`
		Item        []struct {
			Text           string `xml:",chardata"`
			Title          string `xml:"title"`
			Guid           string `xml:"guid"`
			Jackettindexer struct {
				Text string `xml:",chardata"`
				ID   string `xml:"id,attr"`
			} `xml:"jackettindexer"`
			Type        string   `xml:"type"`
			Comments    string   `xml:"comments"`
			PubDate     string   `xml:"pubDate"`
			Size        string   `xml:"size"`
			Description string   `xml:"description"`
			Link        *string  `xml:"link"`
			Category    []string `xml:"category"`
			Enclosure   struct {
				Text   string `xml:",chardata"`
				URL    string `xml:"url,attr"`
				Length string `xml:"length,attr"`
				Type   string `xml:"type,attr"`
			} `xml:"enclosure"`
			Attr []struct {
				Text  string `xml:",chardata"`
				Name  string `xml:"name,attr"`
				Value string `xml:"value,attr"`
			} `xml:"attr"`
		} `xml:"item"`
	} `xml:"channel"`
}

func (s JackettSource) GetRatio() float64 {
	if s.Ratio == nil {
		return 1
	}
	return *s.Ratio
}

func (s JackettSource) GetTargetFolder() string {
	if s.TargetFolder == nil {
		return ""
	}
	return *s.TargetFolder
}

func (s JackettSource) GetRssUrl() string {
	return s.RssUrl
}

func (s JackettSource) GetSeedMinutes() int {
	if s.SeedMinutes == nil {
		return 0
	}
	return *s.SeedMinutes
}
