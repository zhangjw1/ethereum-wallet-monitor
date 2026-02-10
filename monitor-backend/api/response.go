package api

import (
	"encoding/json"
	"net/http"
)

// Response 统一 JSON 响应
type Response struct {
	Code    int         `json:"code"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

func JSON(w http.ResponseWriter, code int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(Response{Code: code, Data: data})
}

func JSONErr(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(Response{Code: code, Message: message})
}

func ParseLimit(q string, defaultLimit, maxLimit int) int {
	// 由各 handler 自己解析 limit，这里仅提供默认值
	return defaultLimit
}
