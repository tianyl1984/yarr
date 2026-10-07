package opml

type OpmlCompare struct {
	Title   string `json:"title"`
	FeedUrl string `json:"feedUrl"`
	SiteUrl string `json:"siteUrl"`
	Tip     string `json:"tip"`
}

type Folder struct {
	Title   string
	Folders []Folder
	Feeds   []Feed
}

type Feed struct {
	Title   string
	FeedUrl string
	SiteUrl string
}

func (f Folder) AllFeeds() []Feed {
	feeds := make([]Feed, 0)
	feeds = append(feeds, f.Feeds...)
	for _, subfolder := range f.Folders {
		feeds = append(feeds, subfolder.AllFeeds()...)
	}
	return feeds
}
