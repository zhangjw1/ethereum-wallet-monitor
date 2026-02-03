package analyzer

import (
	"encoding/json"
	"ethereum-monitor/config"
	"fmt"
	"io"
	"net/http"
	"time"
)

// HoneypotDetector 蜜罐检测器
type HoneypotDetector struct {
	httpClient *http.Client
	apiKey     string // GoPlus API Key（可选）
}

// HoneypotResult 蜜罐检测结果
type HoneypotResult struct {
	IsHoneypot bool
	Reason     string
	BuyTax     float64
	SellTax    float64
	CanBuy     bool
	CanSell    bool
}

// NewHoneypotDetector 创建蜜罐检测器
func NewHoneypotDetector(apiKey string) *HoneypotDetector {
	return &HoneypotDetector{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		apiKey:     apiKey,
	}
}

// CheckHoneypot 检测代币是否是蜜罐
func (h *HoneypotDetector) CheckHoneypot(tokenAddress string) (*HoneypotResult, error) {
	// 优先使用 GoPlus API（更详细）
	if h.apiKey != "" {
		result, err := h.checkWithGoPlus(tokenAddress)
		if err == nil {
			return result, nil
		}
		// GoPlus 失败，降级到 Honeypot.is
	}

	// 使用 Honeypot.is API
	return h.checkWithHoneypotIs(tokenAddress)
}

// checkWithHoneypotIs 使用 Honeypot.is API
func (h *HoneypotDetector) checkWithHoneypotIs(tokenAddress string) (*HoneypotResult, error) {
	url := fmt.Sprintf("%s?address=%s", config.HoneypotAPIURL, tokenAddress)

	resp, err := h.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to call honeypot API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// 如果返回 404，说明 API 还没有该代币数据（通常是因为太新）
		// 这种情况下，我们不能判定为蜜罐，也不能完全判定安全
		// 暂时返回非蜜罐，但标记原因
		if resp.StatusCode == http.StatusNotFound {
			return &HoneypotResult{
				IsHoneypot: false,
				Reason:     "Honeypot API data not found (too new)",
				CanBuy:     true, // 假设可买
				CanSell:    true, // 假设可卖
				BuyTax:     0,    // 假设0税
				SellTax:    0,
			}, nil
		}
		return nil, fmt.Errorf("honeypot API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var apiResp struct {
		IsHoneypot bool `json:"isHoneypot"`
		Summary    struct {
			Risk string `json:"risk"`
		} `json:"summary"`
		SimulationResult struct {
			BuyTax  float64 `json:"buyTax"`
			SellTax float64 `json:"sellTax"`
		} `json:"simulationResult"`
	}

	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	result := &HoneypotResult{
		IsHoneypot: apiResp.IsHoneypot,
		BuyTax:     apiResp.SimulationResult.BuyTax,
		SellTax:    apiResp.SimulationResult.SellTax,
		CanBuy:     true,
		CanSell:    !apiResp.IsHoneypot,
	}

	if apiResp.IsHoneypot {
		result.Reason = "Detected as honeypot by Honeypot.is"
	}

	return result, nil
}

// checkWithGoPlus 使用 GoPlus Security API
func (h *HoneypotDetector) checkWithGoPlus(tokenAddress string) (*HoneypotResult, error) {
	url := fmt.Sprintf("%s?contract_addresses=%s", config.GoPlusAPIURL, tokenAddress)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	if h.apiKey != "" {
		req.Header.Set("Authorization", h.apiKey)
	}

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call GoPlus API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GoPlus API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var apiResp struct {
		Code   int `json:"code"`
		Result map[string]struct {
			IsHoneypot              string `json:"is_honeypot"`
			BuyTax                  string `json:"buy_tax"`
			SellTax                 string `json:"sell_tax"`
			CannotBuy               string `json:"cannot_buy"`
			CannotSellAll           string `json:"cannot_sell_all"`
			HoneypotWithSameCreator string `json:"honeypot_with_same_creator"`
		} `json:"result"`
	}

	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if apiResp.Code != 1 {
		return nil, fmt.Errorf("GoPlus API returned error code %d", apiResp.Code)
	}

	tokenData, ok := apiResp.Result[tokenAddress]
	if !ok {
		return nil, fmt.Errorf("token not found in response")
	}

	result := &HoneypotResult{
		IsHoneypot: tokenData.IsHoneypot == "1" || tokenData.CannotSellAll == "1",
		CanBuy:     tokenData.CannotBuy != "1",
		CanSell:    tokenData.CannotSellAll != "1",
	}

	// 解析税率
	if buyTax, err := parseFloatString(tokenData.BuyTax); err == nil {
		result.BuyTax = buyTax * 100 // 转换为百分比
	}
	if sellTax, err := parseFloatString(tokenData.SellTax); err == nil {
		result.SellTax = sellTax * 100 // 转换为百分比
	}

	// 构造原因
	if result.IsHoneypot {
		if tokenData.IsHoneypot == "1" {
			result.Reason = "Detected as honeypot by GoPlus"
		} else if tokenData.CannotSellAll == "1" {
			result.Reason = "Cannot sell all tokens"
		}
		if tokenData.HoneypotWithSameCreator == "1" {
			result.Reason += "; Creator has deployed other honeypots"
		}
	}

	return result, nil
}

// parseFloatString 解析浮点数字符串
func parseFloatString(s string) (float64, error) {
	var f float64
	_, err := fmt.Sscanf(s, "%f", &f)
	return f, err
}
