package api

import (
	"net/http"
	"strings"
)

// Route 注册三个聚合查询接口，并包装 CORS
func Route(mux *http.ServeMux) {
	mux.HandleFunc("/api/transfer-records", CORS(TransferRecords))
	mux.HandleFunc("/api/notifications", CORS(Notifications))
	mux.HandleFunc("/api/tokens", CORS(Tokens))
}

// CORS 包装 handler，允许 GET 跨域（可选）
func CORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next(w, r)
	}
}

// StripPrefix 与路由配合：若用 /api 前缀挂载，需去掉前缀再匹配
func Handler(prefix string) http.Handler {
	mux := http.NewServeMux()
	Route(mux)
	if prefix == "" {
		return mux
	}
	return http.StripPrefix(strings.TrimSuffix(prefix, "/"), mux)
}
