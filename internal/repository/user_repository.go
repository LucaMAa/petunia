package repository

import (
	"errors"
	"petunia/internal/config"
	"petunia/internal/dto"
	"petunia/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepository interface {
	Create(user *model.User) error
	FindByID(id uuid.UUID) (*model.User, error)
	FindByEmail(email string) (*model.User, error)
	FindAll(filters dto.ListUsersParams) ([]model.User, int64, error)
	FindByStatus(status model.UserStatus) ([]model.User, error)
	ExistsByEmail(email string) (bool, error)
	Save(user *model.User) error
	Delete(id uuid.UUID) error
	SearchByName(query string, excludeUserID uuid.UUID) ([]model.User, error)
}
type userRepository struct {
	db *gorm.DB
}

func NewUserRepository() UserRepository {
	return &userRepository{db: config.DB}
}

func (r *userRepository) Create(user *model.User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) FindByID(id uuid.UUID) (*model.User, error) {
	var user model.User
	err := r.db.First(&user, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &user, err
}

func (r *userRepository) FindByEmail(email string) (*model.User, error) {
	var user model.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &user, err
}

func (r *userRepository) FindAll(filters dto.ListUsersParams) ([]model.User, int64, error) {
	var users []model.User
	var total int64

	q := r.db.Model(&model.User{})

	if filters.Search != "" {
		search := filters.Search
		like := "%" + search + "%"

		q = q.Where(
			`(
				to_tsvector('simple',
					coalesce(first_name,'') || ' ' ||
					coalesce(last_name,'') || ' ' ||
					coalesce(email,'')
				) @@ plainto_tsquery('simple', ?)
			) OR (
				first_name ILIKE ? OR last_name ILIKE ? OR email ILIKE ?
			)`,
			search, like, like, like,
		)
	}

	if filters.Enabled != nil && *filters.Enabled {
		q = q.Where("status = ?", model.StatusEnabled)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	limit := filters.Limit
	if limit <= 0 {
		limit = 10
	}

	err := q.
		Order("created_at DESC").
		Limit(limit).
		Offset(filters.Offset).
		Find(&users).Error

	return users, total, err
}

func (r *userRepository) FindByStatus(status model.UserStatus) ([]model.User, error) {
	var users []model.User
	err := r.db.Where("status = ?", status).Order("created_at ASC").Find(&users).Error
	return users, err
}

func (r *userRepository) ExistsByEmail(email string) (bool, error) {
	var count int64
	err := r.db.Model(&model.User{}).Where("email = ?", email).Count(&count).Error
	return count > 0, err
}

func (r *userRepository) Save(user *model.User) error {
	return r.db.Save(user).Error
}

func (r *userRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&model.User{}, "id = ?", id).Error
}

func (r *userRepository) SearchByName(query string, excludeUserID uuid.UUID) ([]model.User, error) {
	var users []model.User
	like := "%" + query + "%"

	err := r.db.
		Where(`
			(
				to_tsvector('simple',
					coalesce(first_name,'') || ' ' || coalesce(last_name,'')
				) @@ plainto_tsquery('simple', ?)
			) OR (
				first_name ILIKE ? OR last_name ILIKE ?
				OR (first_name || ' ' || last_name) ILIKE ?
			)
		`, query, like, like, like).
		Where("id != ? AND status = ?", excludeUserID, model.StatusEnabled).
		Limit(10).
		Find(&users).Error

	return users, err
}
