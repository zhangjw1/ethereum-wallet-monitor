package api

import (
	"ethereum-monitor/database"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Tokens 聚合查询：支持 address、status、risk_level、max_risk_score、pending_liquidity、date(每日统计)、limit
// GET /api/tokens?address=0x... | status=MONITORING | risk_level=low | max_risk_score=50 | pending_liquidity=1 | date=2025-02-10 | limit=20
func Tokens(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		JSONErr(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	q := r.URL.Query()
	limit := parseLimit(q.Get("limit"), defaultLimit, maxLimit)
	repo := database.NewTokenAnalysisRepository()

	// 1) 按代币地址查单条
	if address := strings.TrimSpace(q.Get("address")); address != "" {
		token, err := repo.GetByAddress(address)
		if err != nil {
			JSONErr(w, http.StatusNotFound, "not found")
			return
		}
		JSON(w, http.StatusOK, token)
		return
	}

	// 2) 指定 date 时返回该日每日统计
	if dateStr := strings.TrimSpace(q.Get("date")); dateStr != "" {
		date, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			JSONErr(w, http.StatusBadRequest, "invalid date, use YYYY-MM-DD")
			return
		}
		stats, err := repo.GetDailyStats(date)
		if err != nil {
			JSONErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		JSON(w, http.StatusOK, stats)
		return
	}

	// 3) 待加池
	if q.Get("pending_liquidity") == "1" || strings.ToLower(q.Get("pending_liquidity")) == "true" {
		list, err := repo.GetPendingLiquidityTokens(limit)
		if err != nil {
			JSONErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		JSON(w, http.StatusOK, list)
		return
	}

	// 4) 低风险（max_risk_score）
	if s := strings.TrimSpace(q.Get("max_risk_score")); s != "" {
		score, err := strconv.ParseFloat(s, 64)
		if err != nil {
			JSONErr(w, http.StatusBadRequest, "invalid max_risk_score")
			return
		}
		list, err := repo.GetLowRiskTokens(score, limit)
		if err != nil {
			JSONErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		JSON(w, http.StatusOK, list)
		return
	}

	// 5) 按状态
	if status := strings.TrimSpace(q.Get("status")); status != "" {
		list, err := repo.GetByStatus(status, limit)
		if err != nil {
			JSONErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		JSON(w, http.StatusOK, list)
		return
	}

	// 6) 按风险等级
	if riskLevel := strings.TrimSpace(q.Get("risk_level")); riskLevel != "" {
		list, err := repo.GetByRiskLevel(riskLevel, limit)
		if err != nil {
			JSONErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		JSON(w, http.StatusOK, list)
		return
	}

	// 7) 默认：最近分析列表
	list, err := repo.GetRecentAnalyses(limit)
	if err != nil {
		JSONErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	JSON(w, http.StatusOK, list)
}
