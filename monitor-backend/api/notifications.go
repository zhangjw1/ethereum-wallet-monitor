package api

import (
	"ethereum-monitor/database"
	"net/http"
	"strings"
	"time"
)

// Notifications 聚合查询：支持 tx_hash、type、start/end、limit、stats
// GET /api/notifications?tx_hash=0x... | type=ETH_TRANSFER | start=...&end=... | stats=1 | limit=20
func Notifications(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		JSONErr(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	q := r.URL.Query()
	limit := parseLimit(q.Get("limit"), defaultLimit, maxLimit)
	repo := database.NewWechatAlterRepository()

	// 1) 统计
	if q.Get("stats") == "1" || strings.ToLower(q.Get("stats")) == "true" {
		stats, err := repo.GetStatistics()
		if err != nil {
			JSONErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		JSON(w, http.StatusOK, stats)
		return
	}

	// 2) 按交易哈希查单条
	if txHash := strings.TrimSpace(q.Get("tx_hash")); txHash != "" {
		alter, err := repo.GetByTxHash(txHash)
		if err != nil {
			JSONErr(w, http.StatusNotFound, "not found")
			return
		}
		JSON(w, http.StatusOK, alter)
		return
	}

	// 3) 按类型查
	if notifType := strings.TrimSpace(q.Get("type")); notifType != "" {
		list, err := repo.GetByType(notifType, limit)
		if err != nil {
			JSONErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		JSON(w, http.StatusOK, list)
		return
	}

	// 4) 按时间范围查
	if startStr, endStr := q.Get("start"), q.Get("end"); startStr != "" && endStr != "" {
		start, err1 := time.Parse(time.RFC3339, startStr)
		end, err2 := time.Parse(time.RFC3339, endStr)
		if err1 != nil || err2 != nil {
			JSONErr(w, http.StatusBadRequest, "invalid start/end, use RFC3339")
			return
		}
		list, err := repo.GetByDateRange(start, end)
		if err != nil {
			JSONErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		if len(list) > limit {
			list = list[:limit]
		}
		JSON(w, http.StatusOK, list)
		return
	}

	// 5) 默认：最近 N 条
	list, err := repo.GetRecent(limit)
	if err != nil {
		JSONErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	JSON(w, http.StatusOK, list)
}
