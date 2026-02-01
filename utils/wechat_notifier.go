package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// WechatNotifier 微信通知器
type WechatNotifier struct {
	webhookURL string // 企业微信机器人 Webhook URL
	client     *http.Client
}

// NewWechatNotifier 创建微信通知器
func NewWechatNotifier(webhookURL string) *WechatNotifier {
	return &WechatNotifier{
		webhookURL: webhookURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// WechatMessage 企业微信消息结构
type WechatMessage struct {
	MsgType  string          `json:"msgtype"`
	Markdown *WechatMarkdown `json:"markdown,omitempty"`
	Text     *WechatText     `json:"text,omitempty"`
}

type WechatMarkdown struct {
	Content string `json:"content"`
}

type WechatText struct {
	Content             string   `json:"content"`
	MentionedList       []string `json:"mentioned_list,omitempty"`
	MentionedMobileList []string `json:"mentioned_mobile_list,omitempty"`
}

// SendMarkdown 发送 Markdown 格式消息
func (w *WechatNotifier) SendMarkdown(content string) error {
	msg := WechatMessage{
		MsgType: "markdown",
		Markdown: &WechatMarkdown{
			Content: content,
		},
	}
	return w.send(msg)
}

// SendText 发送文本消息
func (w *WechatNotifier) SendText(content string, mentionAll bool) error {
	msg := WechatMessage{
		MsgType: "text",
		Text: &WechatText{
			Content: content,
		},
	}

	if mentionAll {
		msg.Text.MentionedList = []string{"@all"}
	}

	return w.send(msg)
}

// SendUSDTAlert 发送 USDT 大额转账告警
func (w *WechatNotifier) SendUSDTAlert(direction, from, to, amount, txHash string, blockNum int) error {
	emoji := "📥"
	if direction == "转出" {
		emoji = "📤"
	}

	content := fmt.Sprintf(`## %s USDT 大额转账告警
> **方向**: <font color="warning">%s</font>
> **金额**: <font color="warning">%s USDT</font>
> **发送方**: %s
> **接收方**: %s
> **区块**: %d
> **交易**: [查看详情](https://etherscan.io/tx/%s)
> **时间**: %s`,
		emoji,
		direction,
		amount,
		from,
		to,
		blockNum,
		txHash,
		time.Now().Format("2006-01-02 15:04:05"))

	return w.SendMarkdown(content)
}

// SendETHAlert 发送 ETH 大额转账告警
func (w *WechatNotifier) SendETHAlert(direction, from, to, amount, txHash string, blockNum uint64, gasPrice string) error {
	emoji := "📥"
	if direction == "转出" {
		emoji = "📤"
	}

	content := fmt.Sprintf(`## %s ETH 大额转账告警
> **方向**: <font color="warning">%s</font>
> **金额**: <font color="warning">%s ETH</font>
> **发送方**: %s
> **接收方**: %s
> **区块**: %d
> **Gas价格**: %s Gwei
> **交易**: [查看详情](https://etherscan.io/tx/%s)
> **时间**: %s`,
		emoji,
		direction,
		amount,
		from,
		to,
		blockNum,
		gasPrice,
		txHash,
		time.Now().Format("2006-01-02 15:04:05"))

	return w.SendMarkdown(content)
}

// SendMEVDetection 发送 MEV 检测通知（仅记录，不告警）
func (w *WechatNotifier) SendMEVDetection(mevType, from, to, amount, txHash string, confidence float64, evidence []string) error {
	evidenceStr := ""
	for i, e := range evidence {
		evidenceStr += fmt.Sprintf("\n> %d. %s", i+1, e)
	}

	content := fmt.Sprintf(`## 🤖 MEV 攻击检测
> **类型**: <font color="info">%s</font>
> **置信度**: %.0f%%
> **发送方**: %s
> **接收方**: %s
> **金额**: %s USDT
> **交易**: [查看详情](https://etherscan.io/tx/%s)
> **证据**: %s
> **时间**: %s`,
		mevType,
		confidence*100,
		from,
		to,
		amount,
		txHash,
		evidenceStr,
		time.Now().Format("2006-01-02 15:04:05"))

	return w.SendMarkdown(content)
}

// send 发送消息到企业微信
func (w *WechatNotifier) send(msg WechatMessage) error {
	if w.webhookURL == "" {
		return fmt.Errorf("webhook URL 未配置")
	}

	jsonData, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("序列化消息失败: %w", err)
	}

	resp, err := w.client.Post(w.webhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("发送失败，状态码: %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("解析响应失败: %w", err)
	}

	if errcode, ok := result["errcode"].(float64); ok && errcode != 0 {
		return fmt.Errorf("企业微信返回错误: %v", result["errmsg"])
	}

	return nil
}

// ServerChanNotifier Server酱通知器
type ServerChanNotifier struct {
	sendKey string
	client  *http.Client
}

// NewServerChanNotifier 创建 Server酱 通知器
func NewServerChanNotifier(sendKey string) *ServerChanNotifier {
	return &ServerChanNotifier{
		sendKey: sendKey,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Send 发送 Server酱 通知
func (s *ServerChanNotifier) Send(title, content string) error {
	if s.sendKey == "" {
		return fmt.Errorf("SendKey 未配置")
	}

	url := fmt.Sprintf("https://sctapi.ftqq.com/%s.send", s.sendKey)

	data := map[string]string{
		"title": title,
		"desp":  content,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("序列化消息失败: %w", err)
	}

	resp, err := s.client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("发送失败，状态码: %d", resp.StatusCode)
	}

	return nil
}

// SendUSDTAlert 发送 USDT 告警
func (s *ServerChanNotifier) SendUSDTAlert(direction, from, to, amount, txHash string, blockNum int) error {
	title := fmt.Sprintf("🚨 USDT 大额%s告警", direction)
	content := fmt.Sprintf(`## 交易详情
- **金额**: %s USDT
- **发送方**: %s
- **接收方**: %s
- **区块**: %d
- **交易哈希**: [%s](https://etherscan.io/tx/%s)
- **时间**: %s`,
		amount,
		from,
		to,
		blockNum,
		txHash,
		txHash,
		time.Now().Format("2006-01-02 15:04:05"))

	return s.Send(title, content)
}

// PushPlusNotifier PushPlus 通知器（推荐，免费200条/天）
type PushPlusNotifier struct {
	token  string
	client *http.Client
}

// NewPushPlusNotifier 创建 PushPlus 通知器
func NewPushPlusNotifier(token string) *PushPlusNotifier {
	return &PushPlusNotifier{
		token: token,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Send 发送 PushPlus 通知
func (p *PushPlusNotifier) Send(title, content string) error {
	if p.token == "" {
		return fmt.Errorf("Token 未配置")
	}

	url := "http://www.pushplus.plus/send"

	data := map[string]string{
		"token":    p.token,
		"title":    title,
		"content":  content,
		"template": "markdown",
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("序列化消息失败: %w", err)
	}

	resp, err := p.client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("发送失败，状态码: %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("解析响应失败: %w", err)
	}

	if code, ok := result["code"].(float64); ok && code != 200 {
		return fmt.Errorf("PushPlus 返回错误: %v", result["msg"])
	}

	return nil
}

// SendUSDTAlert 发送 USDT 告警
func (p *PushPlusNotifier) SendUSDTAlert(direction, from, to, amount, txHash string, blockNum int) error {
	emoji := "📥"
	if direction == "转出" {
		emoji = "📤"
	}

	title := fmt.Sprintf("%s USDT 大额%s告警", emoji, direction)
	content := fmt.Sprintf(`## 交易详情

**金额**: %s USDT  
**发送方**: %s  
**接收方**: %s  
**区块**: %d  
**交易**: [查看详情](https://etherscan.io/tx/%s)  
**时间**: %s`,
		amount,
		from,
		to,
		blockNum,
		txHash,
		time.Now().Format("2006-01-02 15:04:05"))

	return p.Send(title, content)
}

// SendMEVDetection 发送 MEV 检测通知
func (p *PushPlusNotifier) SendMEVDetection(mevType, from, to, amount, txHash string, confidence float64, evidence []string) error {
	evidenceStr := ""
	for i, e := range evidence {
		evidenceStr += fmt.Sprintf("\n%d. %s", i+1, e)
	}

	title := "🤖 MEV 攻击检测"
	content := fmt.Sprintf(`## MEV 检测详情

**类型**: %s  
**置信度**: %.0f%%  
**发送方**: %s  
**接收方**: %s  
**金额**: %s USDT  
**交易**: [查看详情](https://etherscan.io/tx/%s)  
**证据**: %s  
**时间**: %s`,
		mevType,
		confidence*100,
		from,
		to,
		amount,
		txHash,
		evidenceStr,
		time.Now().Format("2006-01-02 15:04:05"))

	return p.Send(title, content)
}

// SendCustomAlert 发送自定义告警
func (p *PushPlusNotifier) SendCustomAlert(title, content string) error {
	return p.Send(title, content)
}
