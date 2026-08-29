package server

import (
	"log"
	"net"
	"net/http"
	"sync"

	"github.com/nkanaev/yarr/src/htmlfeed"
	"github.com/nkanaev/yarr/src/storage"
	"github.com/nkanaev/yarr/src/worker"
)

// dailyRefreshHour is the local hour at which all feeds are auto-refreshed.
const dailyRefreshHour = 4

type Server struct {
	Addr        string
	db          *storage.Storage
	worker      *worker.Worker
	htmlfeed    *htmlfeed.HtmlFeed
	cache       map[string]interface{}
	cache_mutex *sync.Mutex

	// auth (cf-worker-auth SSO)
	AuthURL    string
	AuthSecret string
}

func NewServer(db *storage.Storage, addr string) *Server {
	return &Server{
		db:          db,
		Addr:        addr,
		worker:      worker.NewWorker(db),
		htmlfeed:    htmlfeed.NewHtmlFeed(),
		cache:       make(map[string]interface{}),
		cache_mutex: &sync.Mutex{},
	}
}

func (h *Server) GetAddr() string {
	return "http://" + h.Addr
}

func (s *Server) Start() {
	s.worker.FindFavicons()
	// s.worker.StartFeedCleaner()
	// Feeds are refreshed once a day at 04:00 local time; no refresh on startup.
	s.worker.StartDailyRefresh(dailyRefreshHour)

	ln, err := net.Listen("tcp", s.Addr)
	if err != nil {
		log.Fatal(err)
	}

	httpserver := &http.Server{Handler: s.handler()}
	err = httpserver.Serve(ln)

	if err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
