package htmlfeed

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/nkanaev/yarr/src/httpclient"
	"github.com/nkanaev/yarr/src/storage"
)

type HtmlFeed struct {
	client      *http.Client
	proxyClient *http.Client
}

type rss struct {
	XMLName xml.Name `xml:"rss"`
	Version string   `xml:"version,attr"`
	Title   string   `xml:"channel>title"`
	Link    string   `xml:"channel>link"`
	Items   []item   `xml:"channel>item"`
}

type item struct {
	GUID        string `xml:"guid"`
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func NewHtmlFeed() *HtmlFeed {
	return &HtmlFeed{
		client:      httpclient.NewClient(),
		proxyClient: httpclient.NewProxyClient(),
	}
}

func (h *HtmlFeed) TryGetFeeds(url string, db *storage.Storage) (io.Reader, error) {
	feedConfig, err := db.GetFeedConfig(url)
	if err != nil {
		return nil, err
	}
	if feedConfig == nil {
		return nil, nil
	}
	html, err := h.getHtml(feedConfig)
	if err != nil {
		return nil, err
	}
	defer html.Close()
	doc, err := goquery.NewDocumentFromReader(html)
	if err != nil {
		return nil, err
	}
	rss, err := parseRss(doc, feedConfig)
	if err != nil {
		return nil, err
	}
	if len(rss.Items) == 0 {
		return nil, errors.New("没有找到任何条目")
	}
	output, err := xml.MarshalIndent(rss, "", "  ")
	if err != nil {
		log.Print(err)
		return nil, err
	}
	return bytes.NewBuffer(output), nil
}

func parseRss(doc *goquery.Document, feedConfig *storage.FeedConfig) (*rss, error) {
	var config storage.Config
	err := json.Unmarshal([]byte(feedConfig.Config), &config)
	if err != nil {
		log.Print("配置错误", err)
		return nil, err
	}
	rss := &rss{
		Version: "2.0",
		Title:   config.Title,
		Link:    config.Link,
	}
	itemConfig := config.Items
	doc.Find(itemConfig.Root).Each(func(i int, s *goquery.Selection) {
		item := &item{}
		title := selectNode(s, itemConfig.Title).Text()
		item.Title = title
		href := selectNode(s, itemConfig.Link).AttrOr("href", "")
		if itemConfig.LinkHost != "" {
			href = itemConfig.LinkHost + href
		}
		item.Link = href
		item.GUID = item.Link
		if itemConfig.Description != "" {
			item.Description = selectNode(s, itemConfig.Description).Text()
		}
		pubDate := strings.TrimSpace(selectNode(s, itemConfig.PubDate).Text())
		pubDate = findTime(pubDate)
		t, err := time.Parse(itemConfig.PubDateFmt, pubDate)
		if err != nil {
			log.Print("错误时间格式:", err)
			item.PubDate = pubDate
		} else {
			item.PubDate = t.Format(time.RFC1123Z)
		}
		if item.Title == "" || item.Link == "" {
			return
		}
		rss.Items = append(rss.Items, *item)
	})
	return rss, nil
}

func findTime(pubDate string) string {
	re := regexp.MustCompile(`\d{4}-\d{2}-\d{2}`)
	result := re.FindString(pubDate)
	if result != "" {
		return result
	}
	return pubDate
}

func selectNode(s *goquery.Selection, selector string) *goquery.Selection {
	if selector == "" {
		return s
	}
	if !strings.Contains(selector, ":") {
		return s.Find(selector)
	}
	parts := strings.Split(selector, ":")
	numStr := parts[1][3 : len(parts[1])-1]
	num, err := strconv.Atoi(numStr)
	if err != nil {
		return s.Find(parts[0])
	}
	return s.Find(parts[0]).Eq(num)
}

func (h *HtmlFeed) getHtml(feedConfig *storage.FeedConfig) (io.ReadCloser, error) {
	if feedConfig.UseBrowserless {
		return h.getHtmlBrowserless(feedConfig.Url)
	}
	client := h.client
	if feedConfig.UseProxy {
		client = h.proxyClient
	}
	res, err := client.Get(feedConfig.Url)
	if err != nil {
		log.Print(err)
		return nil, err
	}
	return res.Body, nil
}

func (h *HtmlFeed) getHtmlBrowserless(url string) (io.ReadCloser, error) {
	browserless := os.Getenv("YARR_BROWSERLESS")
	if browserless == "" {
		return nil, nil
	}
	post_url := browserless + "/content"
	data := map[string]string{
		"url": url,
	}
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}
	res, err := h.client.Post(post_url, "application/json", bytes.NewBuffer(jsonBytes))
	if err != nil {
		log.Print(err)
		return nil, err
	}
	return res.Body, nil
}
