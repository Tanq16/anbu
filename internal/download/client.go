package download

import (
	"fmt"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"
)

type ClientConfig struct {
	ProxyURL    string
	UserAgent   string
	Headers     map[string]string
	Connections int
}

type Client struct {
	http      *http.Client
	userAgent string
	headers   map[string]string
}

func NewClient(cfg ClientConfig) (*Client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	idlePerHost := 10
	if cfg.Connections > idlePerHost {
		idlePerHost = cfg.Connections
	}
	transport := &http.Transport{
		DialContext:           (&net.Dialer{Timeout: 10 * time.Second}).DialContext,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 30 * time.Second,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   idlePerHost,
		IdleConnTimeout:       90 * time.Second,
		DisableCompression:    true,
	}
	if cfg.ProxyURL != "" {
		proxyURL, err := url.Parse(cfg.ProxyURL)
		if err != nil {
			return nil, fmt.Errorf("invalid proxy URL: %w", err)
		}
		transport.Proxy = http.ProxyURL(proxyURL)
	}
	ua := cfg.UserAgent
	if ua == "" {
		ua = "Anbu-CLI"
	}
	return &Client{
		http: &http.Client{
			Transport: transport,
			Jar:       jar,
		},
		userAgent: ua,
		headers:   cfg.Headers,
	}, nil
}

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	req.Header.Set("User-Agent", c.userAgent)
	for k, v := range c.headers {
		req.Header.Set(k, v)
	}
	return c.http.Do(req)
}
