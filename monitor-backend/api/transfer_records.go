package api

import (
	"ethereum-monitor/database"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const defaultLimit = 20
const maxLimit = 100

// TransferRecords 聚合查询：支持 tx_hash、address、start/end 时间范围、limit
// GET /api/transfer-records?tx_hash=0x... | address=0x... | start=...&end=... | limit=20
func TransferRecords(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		JSONErr(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	q := r.URL.Query()
	limit := parseLimit(q.Get("limit"), defaultLimit, maxLimit)
	repo := database.NewTransferRecordRepository()

	// 1) 按交易哈希查单条
	if txHash := strings.TrimSpace(q.Get("tx_hash")); txHash != "" {
		record, err := repo.GetByTxHash(txHash)
		if err != nil {
			JSONErr(w, http.StatusNotFound, "not found")
			return
		}
		JSON(w, http.StatusOK, record)
		return
	}

	// 2) 按地址查流水
	if address := strings.TrimSpace(q.Get("address")); address != "" {
		list, err := repo.GetByAddress(address, limit)
		if err != nil {
			JSONErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		JSON(w, http.StatusOK, list)
		return
	}

	// 3) 按时间范围查
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
		// 应用 limit
		if len(list) > limit {
			list = list[:limit]
		}
		JSON(w, http.StatusOK, list)
		return
	}

	// 4) 默认：最近 N 条
	list, err := repo.GetRecent(limit)
	if err != nil {
		JSONErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	JSON(w, http.StatusOK, list)
}

func parseLimit(s string, defaultVal, maxVal int) int {
	if s == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(s)
	if err != nil || n <= 0 {
		return defaultVal
	}
	if n > maxVal {
		return maxVal
	}
	return n
}
