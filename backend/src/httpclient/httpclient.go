package httpclient

import (
	"net"
	"net/http"
	"net/url"
	"os"
	"time"
)

func NewClient() *http.Client {
	return createClient(nil)
}

func NewProxyClient() *http.Client {
	proxy_url := os.Getenv("YARR_PROXY")
	if proxy_url != "" {
		proxyURL, _ := url.Parse(proxy_url)
		return createClient(http.ProxyURL(proxyURL))
	}
	return createClient(nil)
}

func createClient(proxy func(*http.Request) (*url.URL, error)) *http.Client {
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
