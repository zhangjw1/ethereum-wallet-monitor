package database

import (
	"ethereum-monitor/model"
	"strings"
	"time"

	"gorm.io/gorm"
)

type TransferRecordRepository struct {
	db *gorm.DB
}

func NewTransferRecordRepository() *TransferRecordRepository {
	return &TransferRecordRepository{
		db: GetDB(),
	}
}

// Create 创建交易流水记录
func (r *TransferRecordRepository) Create(record *model.TransferRecord) error {
	return r.db.Create(record).Error
}

// GetByTxHash 根据交易哈希查询
func (r *TransferRecordRepository) GetByTxHash(txHash string) (*model.TransferRecord, error) {
	var record model.TransferRecord
	err := r.db.Where("tx_hash = ?", txHash).First(&record).Error
	return &record, err
}

// ExistsByTxHash 检查该交易是否已有流水记录
func (r *TransferRecordRepository) ExistsByTxHash(txHash string) bool {
	var count int64
	r.db.Model(&model.TransferRecord{}).Where("tx_hash = ?", txHash).Count(&count)
	return count > 0
}

// GetRecent 获取最近流水（按时间倒序）
func (r *TransferRecordRepository) GetRecent(limit int) ([]*model.TransferRecord, error) {
	var list []*model.TransferRecord
	err := r.db.Order("created_at DESC").Limit(limit).Find(&list).Error
	return list, err
}

// GetByAddress 按地址查询流水（from 或 to 匹配）
func (r *TransferRecordRepository) GetByAddress(address string, limit int) ([]*model.TransferRecord, error) {
	addr := strings.ToLower(address)
	var list []*model.TransferRecord
	err := r.db.Where("from_address = ? OR to_address = ?", addr, addr).
		Order("created_at DESC").Limit(limit).Find(&list).Error
	return list, err
}

// GetByDateRange 按时间范围查询
func (r *TransferRecordRepository) GetByDateRange(start, end time.Time) ([]*model.TransferRecord, error) {
	var list []*model.TransferRecord
	err := r.db.Where("created_at BETWEEN ? AND ?", start, end).
		Order("created_at DESC").Find(&list).Error
	return list, err
}
