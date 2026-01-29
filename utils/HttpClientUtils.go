package utils

import (
	"etherum-monitor/logger"
	"go.uber.org/zap"
	"net/http"
	"net/url"
)

func SetGlobalProxy() {
	logger.Info("正在配置代理...")
	proxyURL, _ := url.Parse("http://127.0.0.1:7890")
	http.DefaultTransport = &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
	}
	logger.Info("代理已设置", zap.String("proxy", "http://127.0.0.1:7890"))
}
