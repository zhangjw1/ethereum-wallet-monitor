package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"go.uber.org/zap"
)

var (
	// DefaultClient 默认 HTTP 客户端
	DefaultClient *http.Client
	// Logger 日志记录器
	Logger *zap.Logger
)

func init() {
	DefaultClient = &http.Client{
		Timeout: 30 * time.Second,
	}
}

// SetLogger 设置日志记录器
func SetLogger(logger *zap.Logger) {
	Logger = logger
}

// SetGlobalProxy 设置全局代理
func SetGlobalProxy(proxyURL string) error {
	if Logger != nil {
		Logger.Info("正在配置代理...", zap.String("proxy", proxyURL))
	}

	parsedURL, err := url.Parse(proxyURL)
	if err != nil {
		return fmt.Errorf("invalid proxy URL: %w", err)
	}

	http.DefaultTransport = &http.Transport{
		Proxy: http.ProxyURL(parsedURL),
	}

	DefaultClient.Transport = http.DefaultTransport

	if Logger != nil {
		Logger.Info("代理已设置", zap.String("proxy", proxyURL))
	}
	return nil
}

// HTTPRequest 通用 HTTP 请求结构
type HTTPRequest struct {
	Method  string
	URL     string
	Headers map[string]string
	Body    interface{}
	Timeout time.Duration
}

// HTTPResponse 通用 HTTP 响应结构
type HTTPResponse struct {
	StatusCode int
	Body       []byte
	Headers    http.Header
}

// DoRequest 执行通用 HTTP 请求
func DoRequest(req *HTTPRequest) (*HTTPResponse, error) {
	var bodyReader io.Reader

	// 处理请求体
	if req.Body != nil {
		jsonData, err := json.Marshal(req.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewBuffer(jsonData)
	}

	// 创建 HTTP 请求
	httpReq, err := http.NewRequest(req.Method, req.URL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 设置请求头
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	// 如果没有设置 Content-Type 且有 Body，默认设置为 JSON
	if req.Body != nil && httpReq.Header.Get("Content-Type") == "" {
		httpReq.Header.Set("Content-Type", "application/json")
	}

	// 选择客户端
	client := DefaultClient
	if req.Timeout > 0 {
		client = &http.Client{
			Timeout:   req.Timeout,
			Transport: DefaultClient.Transport,
		}
	}

	// 执行请求
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return &HTTPResponse{
		StatusCode: resp.StatusCode,
		Body:       body,
		Headers:    resp.Header,
	}, nil
}

// Get 执行 GET 请求
func Get(url string, headers map[string]string) (*HTTPResponse, error) {
	return DoRequest(&HTTPRequest{
		Method:  http.MethodGet,
		URL:     url,
		Headers: headers,
	})
}

// Post 执行 POST 请求
func Post(url string, body interface{}, headers map[string]string) (*HTTPResponse, error) {
	return DoRequest(&HTTPRequest{
		Method:  http.MethodPost,
		URL:     url,
		Headers: headers,
		Body:    body,
	})
}

// Put 执行 PUT 请求
func Put(url string, body interface{}, headers map[string]string) (*HTTPResponse, error) {
	return DoRequest(&HTTPRequest{
		Method:  http.MethodPut,
		URL:     url,
		Headers: headers,
		Body:    body,
	})
}

// Delete 执行 DELETE 请求
func Delete(url string, headers map[string]string) (*HTTPResponse, error) {
	return DoRequest(&HTTPRequest{
		Method:  http.MethodDelete,
		URL:     url,
		Headers: headers,
	})
}

// GetJSON 执行 GET 请求并解析 JSON 响应
func GetJSON(url string, headers map[string]string, result interface{}) error {
	resp, err := Get(url, headers)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(resp.Body))
	}

	if err := json.Unmarshal(resp.Body, result); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return nil
}

// PostJSON 执行 POST 请求并解析 JSON 响应
func PostJSON(url string, body interface{}, headers map[string]string, result interface{}) error {
	resp, err := Post(url, body, headers)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(resp.Body))
	}

	if result != nil {
		if err := json.Unmarshal(resp.Body, result); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return nil
}
