package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
)

// 模拟配置中的 URL
const HoneypotAPIURL = "https://api.honeypot.is/v2/IsHoneypot"

func main() {
	// 加载 .env
	_ = godotenv.Load("../.env")

	// 设置代理（如果环境变量有）
	setupProxy()

	fmt.Println("🔍 开始诊断 Honeypot.is API...")

	// 1. 测试已知存在的 Token (USDT on Mainnet)
	usdtAddress := "0xdAC17F958D2ee523a2206206994597C13D831ec7"
	fmt.Printf("\n[1/2] 测试已知 Token (USDT: %s)...\n", usdtAddress)
	testHoneypot(usdtAddress)

	// 2. 测试可能不存在的 Token (刚才报错的那个)
	// 如果你不知道具体是哪个，可以填一个随机的新生成地址，或者之前的报错地址
	// 这里使用用户提供的地址
	unknownAddress := "0xaa9652166c2b51eb19d80f72564ff0448e31702b"
	fmt.Printf("\n[2/2] 测试目标 Token (%s)...\n", unknownAddress)
	testHoneypot(unknownAddress)
}

func testHoneypot(address string) {
	url := fmt.Sprintf("%s?address=%s", HoneypotAPIURL, address)
	fmt.Printf("请求 URL: %s\n", url)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		fmt.Printf("❌ 请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	fmt.Printf("状态码: %d %s\n", resp.StatusCode, resp.Status)
	if len(body) > 0 {
		fmt.Printf("响应体: %s\n", string(body))

		// 尝试解析
		var result map[string]interface{}
		if err := json.Unmarshal(body, &result); err == nil {
			if isHoneypot, ok := result["isHoneypot"]; ok {
				fmt.Printf("✅ 解析成功: isHoneypot=%v\n", isHoneypot)
			}
		}
	}

	analyzeResult(resp.StatusCode, len(body))
}

func analyzeResult(code int, bodyLen int) {
	if code == 200 {
		fmt.Println("结论: API 工作正常，且找到了 Token 信息。")
	} else if code == 404 {
		// 关键点：区分路径错误还是资源未找到
		// Honeypot.is 其实如果 address 参数不对或者没找到，行为是什么？
		// 让我们看输出。
		fmt.Println("结论: 返回 404 Not Found。")
		fmt.Println("      -> 如果 USDT 测试也是 404，说明 API URL 彻底失效或被封。")
		fmt.Println("      -> 如果 USDT 正常，而这个 Token 404，说明仅仅是该 Token 未被收录。")
	} else {
		fmt.Printf("结论: 未知状态 %d\n", code)
	}
}

func setupProxy() {
	proxy := os.Getenv("HTTP_PROXY")
	if proxy == "" {
		// 尝试硬编码一个常用的本地代理，方便调试
		// proxy = "http://127.0.0.1:7890"
	}
	if proxy != "" {
		os.Setenv("HTTP_PROXY", proxy)
		os.Setenv("HTTPS_PROXY", proxy)
		fmt.Printf("已设置代理: %s\n", proxy)
	}
}
