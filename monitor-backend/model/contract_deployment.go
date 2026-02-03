package model

import "time"

// ContractDeployment 合约部署记录
type ContractDeployment struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	ContractAddress string    `gorm:"type:varchar(42);uniqueIndex;not null" json:"contract_address"`
	DeployerAddress string    `gorm:"type:varchar(42);index" json:"deployer_address"`
	TxHash          string    `gorm:"type:varchar(66);uniqueIndex;not null" json:"tx_hash"`
	BlockNumber     uint64    `gorm:"index" json:"block_number"`
	Timestamp       time.Time `gorm:"index" json:"timestamp"`

	// 合约信息
	IsToken      bool   `gorm:"default:false" json:"is_token"`
	IsVerified   bool   `gorm:"default:false" json:"is_verified"`
	ContractType string `gorm:"type:varchar(50)" json:"contract_type"` // "ERC20", "ERC721", "Other"

	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// TableName 指定表名
func (ContractDeployment) TableName() string {
	return "contract_deployments"
}
