package repositories

import (
	"golangBasicPractice/missionFour/source"

	"gorm.io/gorm"
)

type UserRepository interface {
	Create(user *source.User) error
	GetByID(id uint) (*source.User, error)
	GetByUsername(username string) (*source.User, error)
	GetByEmail(email string) (*source.User, error)
	Update(user *source.User) error
	Delete(id uint) error
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(user *source.User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) GetByID(id uint) (*source.User, error) {
	var user source.User
	err := r.db.Where("id = ?", id).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByUsername(username string) (*source.User, error) {
	var user source.User
	err := r.db.Where("username = ?", username).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByEmail(email string) (*source.User, error) {
	var user source.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Update(user *source.User) error {
	return r.db.Save(user).Error
}

func (r *userRepository) Delete(id uint) error {
	return r.db.Delete(&source.User{}, id).Error
}
