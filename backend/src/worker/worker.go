package worker

import (
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/nkanaev/yarr/src/htmlfeed"
	"github.com/nkanaev/yarr/src/storage"
)

const NUM_WORKERS = 4

type Worker struct {
	db       *storage.Storage
	htmlfeed *htmlfeed.HtmlFeed
	pending  *int32
	reflock  sync.Mutex
	stopper  chan bool
}

func NewWorker(db *storage.Storage) *Worker {
	pending := int32(0)
	return &Worker{db: db, htmlfeed: htmlfeed.NewHtmlFeed(), pending: &pending}
}

func (w *Worker) FindFavicons() {
	go func() {
		for _, feed := range w.db.ListFeedsMissingIcons() {
			w.FindFeedFavicon(feed)
		}
	}()
}

func (w *Worker) FindFeedFavicon(feed storage.Feed) {
	icon, err := findFavicon(feed.Link, feed.FeedLink, feed.UseProxy)
	if err != nil {
		log.Printf("Failed to find favicon for %s (%s): %s", feed.FeedLink, feed.Link, err)
	}
	if icon != nil {
		w.db.UpdateFeedIcon(feed.Id, icon)
	}
}

// StartDailyRefresh schedules a feed refresh every day at the given local hour
// (e.g. hour=4 -> 04:00). Calling it again replaces the previous schedule.
func (w *Worker) StartDailyRefresh(hour int) {
	if w.stopper != nil {
		close(w.stopper)
		w.stopper = nil
	}

	stop := make(chan bool)
	w.stopper = stop

	go func() {
		log.Printf("auto-refresh: scheduled daily at %02d:00", hour)
		for {
			wait := time.Until(nextDailyRun(time.Now(), hour))
			timer := time.NewTimer(wait)
			select {
			case <-timer.C:
				log.Printf("auto-refresh: firing (daily %02d:00)", hour)
				w.RefreshFeeds()
			case <-stop:
				timer.Stop()
				log.Print("auto-refresh: stopping")
				return
			}
		}
	}()
}

// nextDailyRun returns the next occurrence of hour:00 local time after now.
func nextDailyRun(now time.Time, hour int) time.Time {
	next := time.Date(now.Year(), now.Month(), now.Day(), hour, 0, 0, 0, now.Location())
	if !next.After(now) {
		next = next.AddDate(0, 0, 1)
	}
	return next
}

func (w *Worker) RefreshFeeds() {
	w.reflock.Lock()
	defer w.reflock.Unlock()

	if *w.pending > 0 {
		log.Print("Refreshing already in progress")
		return
	}

	feeds := w.db.ListFeeds()
	if len(feeds) == 0 {
		log.Print("Nothing to refresh")
		return
	}

	log.Print("Refreshing feeds")
	atomic.StoreInt32(w.pending, int32(len(feeds)))
	go w.refresher(feeds)
}

func (w *Worker) refresher(feeds []storage.Feed) {
	w.db.ResetFeedErrors()

	srcqueue := make(chan storage.Feed, len(feeds))
	dstqueue := make(chan []storage.Item)

	for range NUM_WORKERS {
		go w.worker(srcqueue, dstqueue)
	}

	for _, feed := range feeds {
		srcqueue <- feed
	}
	for range feeds {
		items := <-dstqueue
		if len(items) > 0 {
			w.db.CreateItems(items)
			w.db.SetFeedSize(items[0].FeedId, len(items))
		}
		atomic.AddInt32(w.pending, -1)
	}
	close(srcqueue)
	close(dstqueue)

	log.Printf("Finished refreshing %d feeds", len(feeds))
}

func (w *Worker) worker(srcqueue <-chan storage.Feed, dstqueue chan<- []storage.Item) {
	for feed := range srcqueue {
		items, err := listItems(feed, w.db, w.htmlfeed)
		if err != nil {
			w.db.SetFeedError(feed.Id, err)
			log.Println("同步失败,feed_id:", feed.Id, ",err:", err)
		} else {
			if len(items) == 0 {
				log.Println("同步成功,但无新数据,feed_id:", feed.Id)
			}
		}
		dstqueue <- items
	}
}
