package server

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/nkanaev/yarr/src/content/htmlutil"
	"github.com/nkanaev/yarr/src/content/readability"
	"github.com/nkanaev/yarr/src/content/sanitizer"
	"github.com/nkanaev/yarr/src/content/silo"
	"github.com/nkanaev/yarr/src/server/auth"
	"github.com/nkanaev/yarr/src/server/gzip"
	"github.com/nkanaev/yarr/src/server/opml"
	"github.com/nkanaev/yarr/src/server/router"
	"github.com/nkanaev/yarr/src/storage"
	"github.com/nkanaev/yarr/src/worker"
)

func (s *Server) handler() http.Handler {
	r := router.NewRouter()

	r.Use(gzip.Middleware)

	if s.AuthURL != "" {
		a := &auth.Middleware{
			AuthURL: s.AuthURL,
			Secret:  s.AuthSecret,
			Public:  []string{"/api/auth/callback"},
		}
		r.Use(a.Handler)
	}

	r.For("/api/status", s.handleStatus)
	r.For("/api/folders", s.handleFolderList)
	r.For("/api/folders/:id", s.handleFolder)
	r.For("/api/feeds", s.handleFeedList)
	r.For("/api/feeds/refresh", s.handleFeedRefresh)
	r.For("/api/feeds/states", s.handleFeedStates)
	r.For("/api/feeds/:id/icon", s.handleFeedIcon)
	r.For("/api/feeds/:id", s.handleFeed)
	r.For("/api/items", s.handleItemList)
	r.For("/api/items/:id", s.handleItem)
	r.For("/api/settings", s.handleSettings)
	r.For("/api/opml/compare", s.handleOPMLCompare)
	r.For("/api/page", s.handlePageCrawl)
	r.For("/api/htmlFeed", s.handleHtmlFeed)
	r.For("/api/auth/callback", s.handleAuthCallback)
	r.For("/api/logout", s.handleLogout)

	return r
}

func (s *Server) handleStatus(c *router.Context) {
	c.JSON(http.StatusOK, map[string]interface{}{
		"stats":         s.db.FeedStats(),
		"authenticated": s.AuthURL != "",
	})
}

func (s *Server) handleFolderList(c *router.Context) {
	switch c.Req.Method {
	case "GET":
		list := s.db.ListFolders()
		c.JSON(http.StatusOK, list)
	case "POST":
		var body FolderCreateForm
		if err := json.NewDecoder(c.Req.Body).Decode(&body); err != nil {
			log.Print(err)
			c.Out.WriteHeader(http.StatusBadRequest)
			return
		}
		if len(body.Title) == 0 {
			c.JSON(http.StatusBadRequest, map[string]string{"error": "Folder title missing."})
			return
		}
		folder := s.db.CreateFolder(body.Title)
		c.JSON(http.StatusCreated, folder)
	default:
		c.Out.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleFolder(c *router.Context) {
	id, err := c.VarInt64("id")
	if err != nil {
		c.Out.WriteHeader(http.StatusBadRequest)
		return
	}
	switch c.Req.Method {
	case "PUT":
		var body FolderUpdateForm
		if err := json.NewDecoder(c.Req.Body).Decode(&body); err != nil {
			log.Print(err)
			c.Out.WriteHeader(http.StatusBadRequest)
			return
		}
		if body.Title != nil {
			s.db.RenameFolder(id, *body.Title)
		}
		if body.IsExpanded != nil {
			s.db.ToggleFolderExpanded(id, *body.IsExpanded)
		}
		c.Out.WriteHeader(http.StatusOK)
	case "DELETE":
		s.db.DeleteFolder(id)
		c.Out.WriteHeader(http.StatusNoContent)
	}
}

func (s *Server) handleFeedRefresh(c *router.Context) {
	if c.Req.Method == "POST" {
		s.worker.RefreshFeeds()
		c.Out.WriteHeader(http.StatusOK)
	} else {
		c.Out.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleFeedStates(c *router.Context) {
	c.JSON(http.StatusOK, s.db.ListFeedStates())
}

type feedicon struct {
	ctype string
	bytes []byte
	etag  string
}

func (s *Server) handleFeedIcon(c *router.Context) {
	id, err := c.VarInt64("id")
	if err != nil {
		c.Out.WriteHeader(http.StatusBadRequest)
		return
	}

	cachekey := "icon:" + strconv.FormatInt(id, 10)
	s.cache_mutex.Lock()
	cachedat := s.cache[cachekey]
	s.cache_mutex.Unlock()
	if cachedat == nil {
		feed := s.db.GetFeed(id)
		if feed == nil || feed.Icon == nil {
			c.Out.WriteHeader(http.StatusNotFound)
			return
		}

		hash := md5.New()
		hash.Write(*feed.Icon)

		etag := fmt.Sprintf("%x", hash.Sum(nil))[:16]

		cachedat = feedicon{
			ctype: http.DetectContentType(*feed.Icon),
			bytes: *(*feed).Icon,
			etag:  etag,
		}
		s.cache_mutex.Lock()
		s.cache[cachekey] = cachedat
		s.cache_mutex.Unlock()
	}

	icon := cachedat.(feedicon)

	if c.Req.Header.Get("If-None-Match") == icon.etag {
		c.Out.WriteHeader(http.StatusNotModified)
		return
	}

	c.Out.Header().Set("Content-Type", icon.ctype)
	c.Out.Header().Set("Etag", icon.etag)
	c.Out.Write(icon.bytes)
}

func (s *Server) handleFeedList(c *router.Context) {
	switch c.Req.Method {
	case "GET":
		list := s.db.ListFeeds()
		c.JSON(http.StatusOK, list)
	case "POST":
		var form FeedCreateForm
		if err := json.NewDecoder(c.Req.Body).Decode(&form); err != nil {
			log.Print(err)
			c.Out.WriteHeader(http.StatusBadRequest)
			return
		}

		result, err := worker.DiscoverFeed(form.Url, form.UseProxy, s.db, s.htmlfeed)
		switch {
		case err != nil:
			log.Printf("Faild to discover feed for %s: %s", form.Url, err)
			c.JSON(http.StatusOK, map[string]string{"status": "notfound"})
		case len(result.Sources) > 0:
			c.JSON(http.StatusOK, map[string]interface{}{"status": "multiple", "choice": result.Sources})
		case result.Feed != nil:
			feed := s.db.CreateFeed(
				result.Feed.Title,
				"",
				result.Feed.SiteURL,
				result.FeedLink,
				form.FolderID,
				form.UseProxy,
			)
			items := worker.ConvertItems(result.Feed.Items, *feed)
			if len(items) > 0 {
				s.db.CreateItems(items)
			}
			s.db.SetFeedState(feed.Id, len(items), nil)
			s.worker.FindFeedFavicon(*feed)

			c.JSON(http.StatusOK, map[string]interface{}{
				"status": "success",
				"feed":   feed,
			})
		default:
			c.JSON(http.StatusOK, map[string]string{"status": "notfound"})
		}
	}
}

func (s *Server) handleFeed(c *router.Context) {
	id, err := c.VarInt64("id")
	if err != nil {
		c.Out.WriteHeader(http.StatusBadRequest)
		return
	}
	switch c.Req.Method {
	case "PUT":
		feed := s.db.GetFeed(id)
		if feed == nil {
			c.Out.WriteHeader(http.StatusBadRequest)
			return
		}
		body := make(map[string]interface{})
		if err := json.NewDecoder(c.Req.Body).Decode(&body); err != nil {
			log.Print(err)
			c.Out.WriteHeader(http.StatusBadRequest)
			return
		}
		if title, ok := body["title"]; ok {
			if reflect.TypeOf(title).Kind() == reflect.String {
				s.db.RenameFeed(id, title.(string))
			}
		}
		if f_id, ok := body["folder_id"]; ok {
			if f_id == nil {
				s.db.UpdateFeedFolder(id, nil)
			} else if reflect.TypeOf(f_id).Kind() == reflect.Float64 {
				folderId := int64(f_id.(float64))
				s.db.UpdateFeedFolder(id, &folderId)
			}
		}
		if link, ok := body["feed_link"]; ok {
			if reflect.TypeOf(link).Kind() == reflect.String {
				s.db.UpdateFeedLink(id, link.(string))
			}
		}
		c.Out.WriteHeader(http.StatusOK)
	case "DELETE":
		s.db.DeleteFeed(id)
		c.Out.WriteHeader(http.StatusNoContent)
	default:
		c.Out.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleItem(c *router.Context) {
	id, err := c.VarInt64("id")
	if err != nil {
		c.Out.WriteHeader(http.StatusBadRequest)
		return
	}
	switch c.Req.Method {
	case "GET":
		item := s.db.GetItem(id)
		if item == nil {
			c.Out.WriteHeader(http.StatusBadRequest)
			return
		}

		// runtime fix for relative links
		if !htmlutil.IsAPossibleLink(item.Link) {
			if feed := s.db.GetFeed(item.FeedId); feed != nil {
				item.Link = htmlutil.AbsoluteUrl(item.Link, feed.Link)
			}
		}

		item.Content = sanitizer.Sanitize(item.Link, item.Content)
		for i, link := range item.MediaLinks {
			item.MediaLinks[i].Description = sanitizer.Sanitize(item.Link, link.Description)
		}

		c.JSON(http.StatusOK, item)
	case "PUT":
		var body ItemUpdateForm
		if err := json.NewDecoder(c.Req.Body).Decode(&body); err != nil {
			log.Print(err)
			c.Out.WriteHeader(http.StatusBadRequest)
			return
		}
		if body.Status != nil {
			s.db.UpdateItemStatus(id, *body.Status)
		}
		c.Out.WriteHeader(http.StatusOK)
	default:
		c.Out.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleItemList(c *router.Context) {
	switch c.Req.Method {
	case "GET":
		perPage := 20
		query := c.Req.URL.Query()

		filter := storage.ItemFilter{}
		if folderID, err := c.QueryInt64("folder_id"); err == nil {
			filter.FolderID = &folderID
		}
		if feedID, err := c.QueryInt64("feed_id"); err == nil {
			filter.FeedID = &feedID
		}
		if after, err := c.QueryInt64("after"); err == nil {
			filter.After = &after
		}
		if status := query.Get("status"); len(status) != 0 {
			if statusValue, ok := storage.StatusValues[status]; ok {
				filter.Status = &statusValue
			}
		}
		if search := query.Get("search"); len(search) != 0 {
			filter.Search = &search
		}
		items := s.db.ListItems(filter, perPage+1, true)
		hasMore := false
		if len(items) == perPage+1 {
			hasMore = true
			items = items[:perPage]
		}

		for i, item := range items {
			if item.Title == "" {
				text := htmlutil.ExtractText(item.Content)
				items[i].Title = htmlutil.TruncateText(text, 140)
			}
		}
		c.JSON(http.StatusOK, map[string]interface{}{
			"list":     items,
			"has_more": hasMore,
		})
	case "PUT":
		filter := storage.MarkFilter{}

		if folderID, err := c.QueryInt64("folder_id"); err == nil {
			filter.FolderID = &folderID
		}
		if feedID, err := c.QueryInt64("feed_id"); err == nil {
			filter.FeedID = &feedID
		}
		s.db.MarkItemsRead(filter)
		c.Out.WriteHeader(http.StatusOK)
	default:
		c.Out.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleSettings(c *router.Context) {
	switch c.Req.Method {
	case "GET":
		c.JSON(http.StatusOK, s.db.GetSettings())
	case "PUT":
		settings := make(map[string]interface{})
		if err := json.NewDecoder(c.Req.Body).Decode(&settings); err != nil {
			c.Out.WriteHeader(http.StatusBadRequest)
			return
		}
		if s.db.UpdateSettings(settings) {
			c.Out.WriteHeader(http.StatusOK)
		} else {
			c.Out.WriteHeader(http.StatusBadRequest)
		}
	}
}

// Retain one file of at most 5 MiB; serialize uploads and parsing to bound
// transient memory as well as the retained buffer.
const maxOPMLSize = 5 << 20

func (s *Server) handleOPMLCompare(c *router.Context) {
	if c.Req.Method != "POST" && c.Req.Method != "GET" {
		c.Out.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	s.opmlMutex.Lock()
	defer s.opmlMutex.Unlock()
	content, filename := s.opmlContent, s.opmlFilename
	if c.Req.Method == "POST" {
		c.Req.Body = http.MaxBytesReader(c.Out, c.Req.Body, maxOPMLSize+(64<<10))
		defer c.Req.Body.Close()
		err := c.Req.ParseMultipartForm(maxOPMLSize + (64 << 10))
		if c.Req.MultipartForm != nil {
			defer c.Req.MultipartForm.RemoveAll()
		}
		if err != nil {
			var sizeErr *http.MaxBytesError
			if errors.As(err, &sizeErr) {
				c.JSON(http.StatusRequestEntityTooLarge, map[string]string{"error": "OPML 文件不能超过 5 MiB。"})
			} else {
				c.JSON(http.StatusBadRequest, map[string]string{"error": "请选择有效的 OPML 文件。"})
			}
			return
		}
		file, header, err := c.Req.FormFile("opml")
		if err != nil {
			c.JSON(http.StatusBadRequest, map[string]string{"error": "请选择 OPML 文件。"})
			return
		}
		defer file.Close()
		if header.Size > maxOPMLSize {
			c.JSON(http.StatusRequestEntityTooLarge, map[string]string{"error": "OPML 文件不能超过 5 MiB。"})
			return
		}
		content, err = io.ReadAll(io.LimitReader(file, maxOPMLSize+1))
		if err != nil || len(content) > maxOPMLSize {
			c.JSON(http.StatusBadRequest, map[string]string{"error": "无法读取 OPML 文件。"})
			return
		}
		filename = header.Filename
	} else if len(content) == 0 {
		c.JSON(http.StatusNotFound, map[string]string{"error": "服务端暂无对比文件，请重新选择（服务重启后文件会清空）。"})
		return
	}
	doc, err := opml.Parse(bytes.NewReader(content))
	if err != nil {
		c.JSON(http.StatusBadRequest, map[string]string{"error": "OPML 文件格式无效。"})
		return
	}
	s.opmlContent, s.opmlFilename = content, filename
	result := make([]opml.OpmlCompare, 0)
	feeds := s.db.ListFeeds()
	feedMap := make(map[string]storage.Feed)
	for _, f := range feeds {
		feedMap[f.FeedLink] = f
	}
	for _, f := range doc.AllFeeds() {
		if _, exists := feedMap[f.FeedUrl]; exists {
			result = append(result, opml.OpmlCompare{
				Title:   f.Title,
				FeedUrl: f.FeedUrl,
				SiteUrl: f.SiteUrl,
				Tip:     "存在",
			})
		} else {
			result = append(result, opml.OpmlCompare{
				Title:   f.Title,
				FeedUrl: f.FeedUrl,
				SiteUrl: f.SiteUrl,
				Tip:     "不存在",
			})
		}
	}
	c.JSON(http.StatusOK, map[string]interface{}{"filename": filename, "results": result})
}

func (s *Server) handlePageCrawl(c *router.Context) {
	url := c.Req.URL.Query().Get("url")

	if newUrl := silo.RedirectURL(url); newUrl != "" {
		url = newUrl
	}
	if content := silo.VideoIFrame(url); content != "" {
		c.JSON(http.StatusOK, map[string]string{
			"content": sanitizer.Sanitize(url, content),
		})
		return
	}
	if isInternalFromURL(url) {
		log.Printf("attempt to access internal IP %s from %s", url, c.Req.RemoteAddr)
		return
	}

	// TODO read from request
	body, err := worker.GetBody(url, false)
	if err != nil {
		log.Print(err)
		c.Out.WriteHeader(http.StatusBadRequest)
		return
	}
	content, err := readability.ExtractContent(strings.NewReader(body))
	if err != nil {
		c.JSON(http.StatusOK, map[string]string{
			"content": "error: " + err.Error(),
		})
		return
	}
	content = sanitizer.Sanitize(url, content)
	c.JSON(http.StatusOK, map[string]string{
		"content": content,
	})
}

func (s *Server) handleHtmlFeed(c *router.Context) {
	url := c.Req.URL.Query().Get("url")
	reader, err := s.htmlfeed.TryGetFeeds(url, s.db)
	if err != nil {
		log.Print(err)
		c.Out.WriteHeader(http.StatusBadRequest)
		return
	}
	c.XML(reader)
}

// handleAuthCallback is the redirect target of the cf-worker-auth SSO service.
// It exchanges the one-time token for the user's info and, on success,
// establishes a local session cookie.
func (s *Server) handleAuthCallback(c *router.Context) {
	token := c.Req.URL.Query().Get("token")
	if token == "" {
		c.Out.WriteHeader(http.StatusBadRequest)
		c.Out.Write([]byte("missing token"))
		return
	}

	info, err := auth.FetchUserInfo(s.AuthURL, token)
	if err != nil {
		log.Print("auth callback failed: ", err)
		c.Out.WriteHeader(http.StatusUnauthorized)
		c.Out.Write([]byte("authentication failed"))
		return
	}

	auth.SetSession(c.Out, info.Login, s.AuthSecret)
	c.Redirect("/")
}

func (s *Server) handleLogout(c *router.Context) {
	auth.Logout(c.Out)
	c.Out.WriteHeader(http.StatusNoContent)
}
