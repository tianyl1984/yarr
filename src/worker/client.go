package worker

import (
	"net"
	"net/http"
	"net/url"
	"os"
	"time"
)

type Client struct {
	httpClient      *http.Client
	proxyHttpClient *http.Client
	userAgent       string
}

func (c *Client) get(url string, useProxy bool) (*http.Response, error) {
	return c.getConditional(url, "", "", useProxy)
}

func (c *Client) getConditional(url, lastModified, etag string, useProxy bool) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", c.userAgent)
	if lastModified != "" {
		req.Header.Set("If-Modified-Since", lastModified)
	}
	if etag != "" {
		req.Header.Set("If-None-Match", etag)
	}
	if useProxy {
		return c.proxyHttpClient.Do(req)
	}
	return c.httpClient.Do(req)
}

var client *Client

func init() {
	httpClient := newClient(nil)
	proxyHttpClient := httpClient
	proxy_url := os.Getenv("YARR_PROXY")
	if proxy_url != "" {
		proxyURL, _ := url.Parse(proxy_url)
		proxyHttpClient = newClient(http.ProxyURL(proxyURL))
	}
	client = &Client{
		httpClient:      httpClient,
		proxyHttpClient: proxyHttpClient,
		userAgent:       "Yarr/1.0",
	}
}

func newClient(proxy func(*http.Request) (*url.URL, error)) *http.Client {
	transport := &http.Transport{
		Proxy: proxy,
		DialContext: (&net.Dialer{
			Timeout: 10 * time.Second,
		}).DialContext,
		DisableKeepAlives:   true,
		TLSHandshakeTimeout: time.Second * 10,
	}
	httpClient := &http.Client{
		Timeout:   time.Second * 30,
		Transport: transport,
	}
	return httpClient
}
