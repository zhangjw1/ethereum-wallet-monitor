package model

import "time"

type WechatAlter struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Type         string    `gorm:"type:varchar(50);not null;index" json:"type"`      // 通知类型：USDT_ALERT, MEV_DETECTION, ETH_ALERT
	Direction    string    `gorm:"type:varchar(20)" json:"direction"`                // 转账方向：转入/转出
	FromAddress  string    `gorm:"type:varchar(42);index" json:"from_address"`       // 发送方地址
	ToAddress    string    `gorm:"type:varchar(42);index" json:"to_address"`         // 接收方地址
	Amount       string    `gorm:"type:varchar(100)" json:"amount"`                  // 金额
	Currency     string    `gorm:"type:varchar(20)" json:"currency"`                 // 币种：USDT, ETH
	TxHash       string    `gorm:"type:varchar(66);uniqueIndex" json:"tx_hash"`      // 交易哈希（唯一）
	BlockNum     int       `gorm:"index" json:"block_num"`                           // 区块号
	MevType      string    `gorm:"type:varchar(50)" json:"mev_type"`                 // MEV 类型
	Confidence   float64   `gorm:"type:decimal(5,2)" json:"confidence"`              // 置信度
	Content      string    `gorm:"type:text" json:"content"`                         // 通知内容
	Status       string    `gorm:"type:varchar(20);default:'success'" json:"status"` // 发送状态：success, failed
	ErrorMsg     string    `gorm:"type:text" json:"error_msg"`                       // 错误信息
	PublishType  string    `gorm:"type:varchar(64)" json:"publish_type"`             // 发布类型：pushplus, wechat, serverchan
	PublishToken string    `gorm:"type:varchar(256)" json:"publish_token"`           // 发布 Token
	CreatedAt    time.Time `gorm:"autoCreateTime;index" json:"created_at"`           // 创建时间
	UpdateAt     time.Time `gorm:"autoUpdateTime" json:"update_at"`                  // 更新时间
}

// TableName 指定表名
func (WechatAlter) TableName() string {
	return "wechat_alters"
}
