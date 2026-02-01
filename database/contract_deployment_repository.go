package database

import (
	"ethereum-monitor/model"
	"time"
)

// ContractDeploymentRepository 合约部署数据访问层
type ContractDeploymentRepository struct{}

// NewContractDeploymentRepository 创建 Repository
func NewContractDeploymentRepository() *ContractDeploymentRepository {
	return &ContractDeploymentRepository{}
}

// Create 创建合约部署记录
func (r *ContractDeploymentRepository) Create(deployment *model.ContractDeployment) error {
	return DB.Create(deployment).Error
}

// GetByAddress 根据合约地址查询
func (r *ContractDeploymentRepository) GetByAddress(address string) (*model.ContractDeployment, error) {
	var deployment model.ContractDeployment
	err := DB.Where("contract_address = ?", address).First(&deployment).Error
	return &deployment, err
}

// GetByTxHash 根据交易哈希查询
func (r *ContractDeploymentRepository) GetByTxHash(txHash string) (*model.ContractDeployment, error) {
	var deployment model.ContractDeployment
	err := DB.Where("tx_hash = ?", txHash).First(&deployment).Error
	return &deployment, err
}

// GetRecentDeployments 获取最近的部署记录
func (r *ContractDeploymentRepository) GetRecentDeployments(limit int) ([]model.ContractDeployment, error) {
	var deployments []model.ContractDeployment
	err := DB.Where("is_token = ?", true).
		Order("block_number DESC").
		Limit(limit).
		Find(&deployments).Error
	return deployments, err
}

// GetDeploymentsByTimeRange 根据时间范围查询
func (r *ContractDeploymentRepository) GetDeploymentsByTimeRange(start, end time.Time) ([]model.ContractDeployment, error) {
	var deployments []model.ContractDeployment
	err := DB.Where("timestamp BETWEEN ? AND ?", start, end).
		Order("timestamp DESC").
		Find(&deployments).Error
	return deployments, err
}

// CountDailyDeployments 统计每日部署数量
func (r *ContractDeploymentRepository) CountDailyDeployments(date time.Time) (int64, error) {
	var count int64
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	err := DB.Model(&model.ContractDeployment{}).
		Where("timestamp BETWEEN ? AND ? AND is_token = ?", startOfDay, endOfDay, true).
		Count(&count).Error
	return count, err
}
