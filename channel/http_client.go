package channel

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/net/proxy"
)

// NewHTTPClient returns an HTTP client, optionally via HTTP(S)/SOCKS5 proxy.
func NewHTTPClient(proxyAddr string, timeout time.Duration) (*http.Client, error) {
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	if proxyAddr == "" {
		return &http.Client{Timeout: timeout}, nil
	}
	proxyURL, err := url.Parse(proxyAddr)
	if err != nil {
		return nil, fmt.Errorf("invalid proxy url: %w", err)
	}
	if proxyURL.Scheme == "" || proxyURL.Host == "" {
		return nil, fmt.Errorf("invalid proxy url: must be like http://user:pass@host:port or socks5://host:port")
	}
	switch proxyURL.Scheme {
	case "http", "https":
		return &http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				Proxy: http.ProxyURL(proxyURL),
			},
		}, nil
	case "socks5", "socks5h":
		var auth *proxy.Auth
		if proxyURL.User != nil {
			password, _ := proxyURL.User.Password()
			auth = &proxy.Auth{
				User:     proxyURL.User.Username(),
				Password: password,
			}
		}
		dialer, err := proxy.SOCKS5("tcp", proxyURL.Host, auth, proxy.Direct)
		if err != nil {
			return nil, err
		}
		if contextDialer, ok := dialer.(proxy.ContextDialer); ok {
			return &http.Client{
				Timeout: timeout,
				Transport: &http.Transport{
					DialContext: contextDialer.DialContext,
				},
			}, nil
		}
		return &http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				Dial: dialer.Dial,
			},
		}, nil
	default:
		return nil, fmt.Errorf("unsupported proxy scheme: %s", proxyURL.Scheme)
	}
}
