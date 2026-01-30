package database

import (
	"ethereum-monitor/model"
	"fmt"
)

// ExampleUsage 数据库使用示例
func ExampleUsage() {
	repo := NewMevBuilderRepository()

	// 1. 查询所有 MEV Builders
	fmt.Println("=== 查询所有 MEV Builders ===")
	builders, err := repo.GetAll()
	if err != nil {
		fmt.Printf("查询失败: %v\n", err)
		return
	}
	fmt.Printf("共找到 %d 个 MEV Builders\n\n", len(builders))

	// 显示前 5 个
	for i, builder := range builders {
		if i >= 5 {
			fmt.Printf("... 还有 %d 个\n\n", len(builders)-5)
			break
		}
		fmt.Printf("%d. %s\n", i+1, builder.Name)
		fmt.Printf("   地址: %s\n", builder.Address)
		if builder.Ens != "" {
			fmt.Printf("   ENS: %s\n", builder.Ens)
		}
		fmt.Println()
	}

	// 2. 根据地址查询
	fmt.Println("=== 根据地址查询 ===")
	testAddress := "0xdafea492d9c6733ae3d56b7ed1adb60692c98bc5"
	builder, err := repo.GetByAddress(testAddress)
	if err != nil {
		fmt.Printf("未找到地址: %s\n", testAddress)
	} else {
		fmt.Printf("找到: %s (%s)\n", builder.Name, builder.Address)
	}
	fmt.Println()

	// 3. 添加新的 MEV Builder（示例）
	fmt.Println("=== 添加新 Builder 示例 ===")
	newBuilder := &model.MevBuilder{
		Name:    "Example Builder",
		Address: "0x1234567890123456789012345678901234567890",
		Ens:     "example.eth",
	}
	fmt.Printf("可以使用 repo.create(newBuilder) 添加新数据\n")
	fmt.Printf("示例: %s (%s)\n", newBuilder.Name, newBuilder.Address)
}
