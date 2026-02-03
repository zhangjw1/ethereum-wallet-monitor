package model

import (
	"time"
)

// MevBuilder MEV Builder 信息模型
type MevBuilder struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	Name       string    `gorm:"type:varchar(255);not null" json:"name"`               // Builder 名称
	Url        string    `gorm:"type:varchar(500)" json:"url"`                         // Builder URL
	Address    string    `gorm:"type:varchar(42);uniqueIndex;not null" json:"address"` // Builder 地址（唯一）
	BotAddress string    `gorm:"type:varchar(42)" json:"bot_address"`                  // 关联的 Bot 地址
	Ens        string    `gorm:"type:varchar(255)" json:"ens"`                         // ENS 域名
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`                     // 创建时间
	UpdatedAt  time.Time `gorm:"autoUpdateTime" json:"updated_at"`                     // 更新时间
}

// TableName 指定表名
func (MevBuilder) TableName() string {
	return "mev_builders"
}
