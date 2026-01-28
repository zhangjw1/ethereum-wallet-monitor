package utils

import (
	"net/http"
	"net/url"
)

func SetGlobalProxy() {
	proxyURL, _ := url.Parse("http://127.0.0.1:7890")
	http.DefaultTransport = &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
	}
}
