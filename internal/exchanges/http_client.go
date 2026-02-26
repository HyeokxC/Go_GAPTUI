package exchanges

import (
	"net/http"
	"net/url"
	"time"
)

func BuildHTTPClient(timeoutSecs uint64, proxyURL string) *http.Client {
	client := &http.Client{
		Timeout: time.Duration(timeoutSecs) * time.Second,
	}

	if proxyURL == "" {
		return client
	}

	proxy, err := url.Parse(proxyURL)
	if err != nil {
		return client
	}

	client.Transport = &http.Transport{Proxy: http.ProxyURL(proxy)}
	return client
}

func BuildSimpleHTTPClient(timeoutSecs uint64) *http.Client {
	return &http.Client{
		Timeout: time.Duration(timeoutSecs) * time.Second,
	}
}
