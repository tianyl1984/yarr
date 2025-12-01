package worker

import (
	"net/http"

	"github.com/nkanaev/yarr/src/httpclient"
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
	httpClient := httpclient.NewClient()
	proxyHttpClient := httpclient.NewProxyClient()
	client = &Client{
		httpClient:      httpClient,
		proxyHttpClient: proxyHttpClient,
		userAgent:       "Yarr/1.0",
	}
}
