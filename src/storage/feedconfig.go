package storage

import "database/sql"

type FeedConfig struct {
	Id     int64  `json:"id"`
	Url    string `json:"url"`
	Config string `json:"config"`
}

type Config struct {
	Title string     `json:"title"`
	Link  string     `json:"link"`
	Items ConfigItem `json:"items"`
}

type ConfigItem struct {
	Root        string `json:"root"`
	Title       string `json:"title"`
	Link        string `json:"link"`
	LinkHost    string `json:"linkHost,omitempty"`
	PubDate     string `json:"pubDate"`
	PubDateFmt  string `json:"pubDateFmt"`
	Description string `json:"description,omitempty"`
}

func (s *Storage) GetFeedConfig(url string) (*FeedConfig, error) {
	var f FeedConfig
	err := s.db.QueryRow("select id, url, config from feed_config where url = ? ", url).Scan(
		&f.Id, &f.Url, &f.Config,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &f, nil
}
