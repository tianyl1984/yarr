package storage

import (
	"database/sql"
	"log"
	"time"
)

type Feed struct {
	Id          int64   `json:"id"`
	FolderId    *int64  `json:"folder_id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Link        string  `json:"link"`
	FeedLink    string  `json:"feed_link"`
	Icon        *[]byte `json:"icon,omitempty"`
	HasIcon     bool    `json:"has_icon"`
	UseProxy    bool    `json:"use_proxy"`
}

func (s *Storage) CreateFeed(title, description, link, feedLink string, folderId *int64, useProxy bool) *Feed {
	if title == "" {
		title = feedLink
	}
	result, err := s.db.Exec(`
		insert into feeds (title, description, link, feed_link, folder_id, use_proxy) values (?, ?, ?, ?, ?, ?) 
		on duplicate key update id = LAST_INSERT_ID(id), use_proxy = ?
		`,
		title, description, link, feedLink, folderId, useProxy, useProxy,
	)
	if err != nil {
		log.Print(err)
		return nil
	}
	id, err := result.LastInsertId()
	if err != nil {
		log.Print(err)
		return nil
	}
	return &Feed{
		Id:          id,
		Title:       title,
		Description: description,
		Link:        link,
		FeedLink:    feedLink,
		FolderId:    folderId,
	}
}

func (s *Storage) DeleteFeed(feedId int64) bool {
	result, err := s.db.Exec(`delete from feeds where id = ?`, feedId)
	if err != nil {
		log.Print(err)
		return false
	}
	if _, err := s.db.Exec(`delete from feed_states where feed_id = ?`, feedId); err != nil {
		log.Print(err)
	}
	nrows, err := result.RowsAffected()
	if err != nil {
		if err != sql.ErrNoRows {
			log.Print(err)
		}
		return false
	}
	return nrows == 1
}

func (s *Storage) RenameFeed(feedId int64, newTitle string) bool {
	_, err := s.db.Exec(`update feeds set title = ? where id = ?`, newTitle, feedId)
	return err == nil
}

func (s *Storage) UpdateFeedFolder(feedId int64, newFolderId *int64) bool {
	_, err := s.db.Exec(`update feeds set folder_id = ? where id = ?`, newFolderId, feedId)
	return err == nil
}

func (s *Storage) UpdateFeedLink(feedId int64, newLink string) bool {
	_, err := s.db.Exec(`update feeds set feed_link = ? where id = ?`, newLink, feedId)
	return err == nil
}

func (s *Storage) UpdateFeedIcon(feedId int64, icon *[]byte) bool {
	_, err := s.db.Exec(`update feeds set icon = ? where id = ?`, icon, feedId)
	return err == nil
}

func (s *Storage) ListFeeds() []Feed {
	result := make([]Feed, 0)
	rows, err := s.db.Query(`
		select id, folder_id, title, description, link, feed_link,
		       ifnull(length(icon), 0) > 0 as has_icon, use_proxy
		from feeds
		order by title
	`)
	if err != nil {
		log.Print(err)
		return result
	}
	for rows.Next() {
		var f Feed
		err = rows.Scan(
			&f.Id,
			&f.FolderId,
			&f.Title,
			&f.Description,
			&f.Link,
			&f.FeedLink,
			&f.HasIcon,
			&f.UseProxy,
		)
		if err != nil {
			log.Print(err)
			return result
		}
		result = append(result, f)
	}
	return result
}

func (s *Storage) ListFeedsMissingIcons() []Feed {
	result := make([]Feed, 0)
	rows, err := s.db.Query(`
		select id, folder_id, title, description, link, feed_link, use_proxy
		from feeds
		where icon is null
	`)
	if err != nil {
		log.Print(err)
		return result
	}
	for rows.Next() {
		var f Feed
		err = rows.Scan(
			&f.Id,
			&f.FolderId,
			&f.Title,
			&f.Description,
			&f.Link,
			&f.FeedLink,
			&f.UseProxy,
		)
		if err != nil {
			log.Print(err)
			return result
		}
		result = append(result, f)
	}
	return result
}

func (s *Storage) GetFeed(id int64) *Feed {
	var f Feed
	err := s.db.QueryRow(`
		select
			id, folder_id, title, link, feed_link,
			icon, ifnull(icon, '') != '' as has_icon,
			use_proxy
		from feeds where id = ?
	`, id).Scan(
		&f.Id, &f.FolderId, &f.Title, &f.Link, &f.FeedLink,
		&f.Icon, &f.HasIcon, &f.UseProxy,
	)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Print(err)
		}
		return nil
	}
	return &f
}

// FeedState is the outcome of the last refresh attempt for a feed. There is at
// most one row per feed (unique key on feed_id); LastSuccess / ItemCount keep
// the values from the last *successful* refresh even when the latest attempt
// failed, and Error is nil whenever the latest attempt succeeded.
type FeedState struct {
	FeedId        int64      `json:"feed_id"`
	LastRefreshed *time.Time `json:"last_refreshed"`
	LastSuccess   *time.Time `json:"last_success"`
	ItemCount     *int       `json:"item_count"`
	Error         *string    `json:"error"`
}

// SetFeedState records the outcome of a refresh attempt. On success pass
// itemCount and a nil error; on failure pass the error — last_success and
// item_count are then left at their previous values.
func (s *Storage) SetFeedState(feedId int64, itemCount int, refreshErr error) {
	now := time.Now()

	var lastSuccess *time.Time
	var count *int
	var errMsg *string

	if refreshErr != nil {
		msg := refreshErr.Error()
		errMsg = &msg
	} else {
		lastSuccess = &now
		count = &itemCount
	}

	_, err := s.db.Exec(`
		insert into feed_states (feed_id, last_refreshed, last_success, item_count, error)
		values (?, ?, ?, ?, ?)
		on duplicate key update
			last_refreshed = values(last_refreshed),
			last_success = ifnull(values(last_success), last_success),
			item_count = ifnull(values(item_count), item_count),
			error = values(error)`,
		feedId, now, lastSuccess, count, errMsg,
	)
	if err != nil {
		log.Print(err)
	}
}

func (s *Storage) ListFeedStates() []FeedState {
	result := make([]FeedState, 0)

	rows, err := s.db.Query(`
		select feed_id, last_refreshed, last_success, item_count, error
		from feed_states
	`)
	if err != nil {
		log.Print(err)
		return result
	}
	defer rows.Close()

	for rows.Next() {
		var st FeedState
		if err = rows.Scan(&st.FeedId, &st.LastRefreshed, &st.LastSuccess, &st.ItemCount, &st.Error); err != nil {
			log.Print(err)
			return result
		}
		result = append(result, st)
	}
	return result
}
