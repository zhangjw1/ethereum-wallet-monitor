package model

import "time"

// TransferRecord 钱包监控交易流水（仅记录会触发通知的转账）
type TransferRecord struct {
	ID uint `gorm:"primaryKey" json:"id"`

	// 监控与交易
	MonitorLabel string `gorm:"type:varchar(100);index" json:"monitor_label"` // 监控地址标签，如 "OKX钱包"
	Direction    string `gorm:"type:varchar(20);not null" json:"direction"`   // 转入 / 转出
	FromAddress  string `gorm:"type:varchar(42);index;not null" json:"from_address"`
	ToAddress    string `gorm:"type:varchar(42);index;not null" json:"to_address"`
	Amount       string `gorm:"type:varchar(100);not null" json:"amount"`
	Currency     string `gorm:"type:varchar(20);not null;index" json:"currency"` // ETH, USDT, USDC 等
	TxHash       string `gorm:"type:varchar(66);uniqueIndex;not null" json:"tx_hash"`
	BlockNumber  int    `gorm:"index;not null" json:"block_number"`

	// 通知状态（与 wechat_alters 对应，便于对账）
	Notified     bool   `gorm:"default:true" json:"notified"`          // 是否已发送通知
	NotifyStatus string `gorm:"type:varchar(20)" json:"notify_status"` // success / failed

	CreatedAt time.Time `gorm:"autoCreateTime;index" json:"created_at"`
}

// TableName 指定表名
func (TransferRecord) TableName() string {
	return "transfer_records"
}
