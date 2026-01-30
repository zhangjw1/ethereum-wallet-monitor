package database

import (
	"ethereum-monitor/model"
	"time"

	"gorm.io/gorm"
)

type WechatAlterRepository struct {
	db *gorm.DB
}

func NewWechatAlterRepository() *WechatAlterRepository {
	return &WechatAlterRepository{
		db: GetDB(),
	}
}

// Create 创建通知记录
func (r *WechatAlterRepository) Create(alter *model.WechatAlter) error {
	return r.db.Create(alter).Error
}

// BatchCreate 批量创建
func (r *WechatAlterRepository) BatchCreate(alters []model.WechatAlter) error {
	return r.db.CreateInBatches(alters, len(alters)).Error
}

// GetByTxHash 根据交易哈希查询
func (r *WechatAlterRepository) GetByTxHash(txHash string) (*model.WechatAlter, error) {
	var alter model.WechatAlter
	err := r.db.Where("tx_hash = ?", txHash).First(&alter).Error
	return &alter, err
}

// ExistsByTxHash 检查交易是否已记录
func (r *WechatAlterRepository) ExistsByTxHash(txHash string) bool {
	var count int64
	r.db.Model(&model.WechatAlter{}).Where("tx_hash = ?", txHash).Count(&count)
	return count > 0
}

// GetRecent 获取最近的通知记录
func (r *WechatAlterRepository) GetRecent(limit int) ([]*model.WechatAlter, error) {
	var alters []*model.WechatAlter
	err := r.db.Order("created_at DESC").Limit(limit).Find(&alters).Error
	return alters, err
}

// GetByType 根据类型查询
func (r *WechatAlterRepository) GetByType(notifType string, limit int) ([]*model.WechatAlter, error) {
	var alters []*model.WechatAlter
	err := r.db.Where("type = ?", notifType).Order("created_at DESC").Limit(limit).Find(&alters).Error
	return alters, err
}

// GetByDateRange 根据日期范围查询
func (r *WechatAlterRepository) GetByDateRange(startTime, endTime time.Time) ([]*model.WechatAlter, error) {
	var alters []*model.WechatAlter
	err := r.db.Where("created_at BETWEEN ? AND ?", startTime, endTime).Order("created_at DESC").Find(&alters).Error
	return alters, err
}

// GetStatistics 获取统计信息
func (r *WechatAlterRepository) GetStatistics() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 总通知数
	var total int64
	r.db.Model(&model.WechatAlter{}).Count(&total)
	stats["total"] = total

	// 成功数
	var success int64
	r.db.Model(&model.WechatAlter{}).Where("status = ?", "success").Count(&success)
	stats["success"] = success

	// 失败数
	var failed int64
	r.db.Model(&model.WechatAlter{}).Where("status = ?", "failed").Count(&failed)
	stats["failed"] = failed

	// 按类型统计
	var typeStats []struct {
		Type  string
		Count int64
	}
	r.db.Model(&model.WechatAlter{}).Select("type, count(*) as count").Group("type").Scan(&typeStats)
	stats["by_type"] = typeStats

	// 今日通知数
	today := time.Now().Truncate(24 * time.Hour)
	var todayCount int64
	r.db.Model(&model.WechatAlter{}).Where("created_at >= ?", today).Count(&todayCount)
	stats["today"] = todayCount

	return stats, nil
}
