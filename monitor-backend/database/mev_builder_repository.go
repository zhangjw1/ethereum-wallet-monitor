package database

import (
	"ethereum-monitor/model"
	"gorm.io/gorm"
)

type MevBuilderRepository struct {
	db *gorm.DB
}

func NewMevBuilderRepository() *MevBuilderRepository {
	return &MevBuilderRepository{
		db: GetDB(),
	}
}

func (r *MevBuilderRepository) create(builder *model.MevBuilder) error {
	return r.db.Create(builder).Error
}

func (r *MevBuilderRepository) batchCreate(builders []model.MevBuilder) error {
	return r.db.CreateInBatches(builders, len(builders)).Error
}

func (r *MevBuilderRepository) update(builder *model.MevBuilder) error {
	return r.db.Save(builder).Error
}

func (r *MevBuilderRepository) GetById(id uint) (*model.MevBuilder, error) {
	var builder model.MevBuilder
	err := r.db.First(&builder, id).Error
	return &builder, err
}

func (r *MevBuilderRepository) GetByAddress(address string) (*model.MevBuilder, error) {
	var builder model.MevBuilder
	err := r.db.Where("address = ?", address).First(&builder).Error
	return &builder, err
}

func (r *MevBuilderRepository) GetAll() ([]*model.MevBuilder, error) {
	var builders []*model.MevBuilder
	err := r.db.Find(&builders).Error
	return builders, err
}

func (r *MevBuilderRepository) count() (int64, error) {
	var count int64
	err := r.db.Model(&model.MevBuilder{}).Count(&count).Error
	return count, err
}

func (r *MevBuilderRepository) deleteById(id uint) error {
	err := r.db.Delete(&model.MevBuilder{}, id).Error
	return err
}
