package utils

import (
	"fmt"
	"log"
)

// ExampleUsage MEV 检测器使用示例
func ExampleUsage() {
	// 创建 MEV 检测器
	detector, err := NewMevDetector("https://eth.llamarpc.com")
	if err != nil {
		log.Fatal("创建检测器失败:", err)
	}
	defer detector.Close()

	// 检测交易（替换为实际的交易哈希）
	txHash := "0xf16a64bd6884a235afc153bd35f093e19243682f8bb804441aaf0736c3194b79"

	result, err := detector.DetectMev(txHash)
	if err != nil {
		log.Fatal("检测失败:", err)
	}

	// 输出结果
	fmt.Println("=== MEV 检测结果 ===")
	fmt.Printf("是否为 MEV 攻击: %v\n", result.IsMev)
	fmt.Printf("攻击类型: %s\n", result.MevType)
	fmt.Printf("置信度: %.2f%%\n", result.Confidence*100)

	if len(result.Evidence) > 0 {
		fmt.Println("\n证据:")
		for i, evidence := range result.Evidence {
			fmt.Printf("  %d. %s\n", i+1, evidence)
		}
	}

	if result.Description != "" {
		fmt.Printf("\n描述: %s\n", result.Description)
	}
}
