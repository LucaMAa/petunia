package repository

import (
	"errors"
	"petunia/internal/config"
	"petunia/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UploadedFileRepository interface {
	Create(f *model.UploadedFile) error
	FindByID(id uuid.UUID) (*model.UploadedFile, error)
	FindByOwner(ownerID uuid.UUID, ownerType model.FileOwnerType, category *model.FileCategory) ([]model.UploadedFile, error)
	Save(f *model.UploadedFile) error
	Delete(id uuid.UUID) error
}

type uploadedFileRepository struct{ db *gorm.DB }

func NewUploadedFileRepository() UploadedFileRepository {
	return &uploadedFileRepository{db: config.DB}
}

func (r *uploadedFileRepository) Create(f *model.UploadedFile) error {
	return r.db.Create(f).Error
}

func (r *uploadedFileRepository) FindByID(id uuid.UUID) (*model.UploadedFile, error) {
	var f model.UploadedFile
	err := r.db.First(&f, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &f, err
}

func (r *uploadedFileRepository) FindByOwner(
	ownerID uuid.UUID,
	ownerType model.FileOwnerType,
	category *model.FileCategory,
) ([]model.UploadedFile, error) {
	q := r.db.Where("owner_id = ? AND owner_type = ? AND deleted_at IS NULL", ownerID, ownerType)
	if category != nil {
		q = q.Where("category = ?", *category)
	}
	var files []model.UploadedFile
	err := q.Order("created_at DESC").Find(&files).Error
	return files, err
}

func (r *uploadedFileRepository) Save(f *model.UploadedFile) error {
	return r.db.Save(f).Error
}

func (r *uploadedFileRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&model.UploadedFile{}, "id = ?", id).Error
}
