package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// CovalentClient Covalent API 客户端
type CovalentClient struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// TokenBalance 代币余额信息
type TokenBalance struct {
	ContractAddress      string  `json:"contract_address"`
	ContractName         string  `json:"contract_name"`
	ContractTickerSymbol string  `json:"contract_ticker_symbol"`
	ContractDecimals     int     `json:"contract_decimals"`
	Balance              string  `json:"balance"`
	Quote                float64 `json:"quote"`      // USD 价值
	QuoteRate            float64 `json:"quote_rate"` // 单价
	LogoURL              string  `json:"logo_url"`
	Type                 string  `json:"type"`         // "cryptocurrency" 或 "nft"
	NativeToken          bool    `json:"native_token"` // 是否为原生代币
}

// BalanceResponse Covalent API 余额响应
type BalanceResponse struct {
	Data struct {
		Address       string         `json:"address"`
		UpdatedAt     string         `json:"updated_at"`
		NextUpdateAt  string         `json:"next_update_at"`
		QuoteCurrency string         `json:"quote_currency"`
		ChainID       int            `json:"chain_id"`
		ChainName     string         `json:"chain_name"`
		Items         []TokenBalance `json:"items"`
	} `json:"data"`
	Error        bool   `json:"error"`
	ErrorMessage string `json:"error_message"`
	ErrorCode    int    `json:"error_code"`
}

// NewCovalentClient 创建 Covalent API 客户端
func NewCovalentClient(apiKey string) *CovalentClient {
	return &CovalentClient{
		apiKey:  apiKey,
		baseURL: "https://api.covalenthq.com/v1",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetTokenBalances 获取指定地址在指定链上的所有代币余额
// chainName: 链名称，如 "eth-mainnet", "matic-mainnet", "bsc-mainnet" 等
// address: 钱包地址
func (c *CovalentClient) GetTokenBalances(chainName, address string) (*BalanceResponse, error) {
	// 构建 API URL
	// 格式: /v1/{chainName}/address/{address}/balances_v2/
	url := fmt.Sprintf("%s/%s/address/%s/balances_v2/", c.baseURL, chainName, address)

	// 创建请求
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	// 检查 HTTP 状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API 返回错误状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	// 解析 JSON 响应
	var balanceResp BalanceResponse
	if err := json.Unmarshal(body, &balanceResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w, 响应内容: %s", err, string(body))
	}

	// 检查 API 错误
	if balanceResp.Error {
		return nil, fmt.Errorf("API 返回错误: %s (错误码: %d)", balanceResp.ErrorMessage, balanceResp.ErrorCode)
	}

	return &balanceResp, nil
}

// GetNativeTokenBalance 仅获取原生代币余额（如 ETH, BNB, MATIC 等）
func (c *CovalentClient) GetNativeTokenBalance(chainName, address string) (*TokenBalance, error) {
	balances, err := c.GetTokenBalances(chainName, address)
	if err != nil {
		return nil, err
	}

	// 查找原生代币
	for _, token := range balances.Data.Items {
		if token.NativeToken {
			return &token, nil
		}
	}

	return nil, fmt.Errorf("未找到原生代币余额")
}

// GetMultiChainBalances 获取多条链上的余额
func (c *CovalentClient) GetMultiChainBalances(address string, chains []string) (map[string]*BalanceResponse, error) {
	results := make(map[string]*BalanceResponse)

	for _, chain := range chains {
		balance, err := c.GetTokenBalances(chain, address)
		if err != nil {
			// 记录错误但继续查询其他链
			results[chain] = nil
			continue
		}
		results[chain] = balance
	}

	return results, nil
}
