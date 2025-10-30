package repositories

import (
	"golangBasicPractice/missionFour/source"
	"gorm.io/gorm"
)

type PostRepository interface {
	Create(post *source.Post) error
	GetByID(id uint) (*source.Post, error)
	GetByUserID(userID uint) ([]source.Post, error)
	Update(post *source.Post) error
	Delete(id uint) error
	GetAll() ([]source.Post, error)
}

type postRepository struct {
	db *gorm.DB
}

func NewPostRepository(db *gorm.DB) PostRepository {
	return &postRepository{db: db}
}

func (r *postRepository) Create(post *source.Post) error {
	return r.db.Create(post).Error
}

func (r *postRepository) GetByID(id uint) (*source.Post, error) {
	var post source.Post
	err := r.db.Preload("User").Where("id = ?", id).First(&post).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &post, nil
}

func (r *postRepository) GetByUserID(userID uint) ([]source.Post, error) {
	var posts []source.Post
	err := r.db.Preload("User").Where("user_id = ?", userID).Find(&posts).Error
	return posts, err
}

func (r *postRepository) Update(post *source.Post) error {
	return r.db.Save(post).Error
}

func (r *postRepository) Delete(id uint) error {
	return r.db.Delete(&source.Post{}, id).Error
}

func (r *postRepository) GetAll() ([]source.Post, error) {
	var posts []source.Post
	err := r.db.Preload("User").Find(&posts).Error
	return posts, err
}
