package repositories

import (
	"golangBasicPractice/missionFour/source"

	"gorm.io/gorm"
)

type CommentRepository interface {
	Create(comment *source.Comment) error
	GetByPostID(postID uint) ([]source.Comment, error)
	GetByID(id uint) (*source.Comment, error)
	Update(comment *source.Comment) error
	Delete(id uint) error
}

type commentRepository struct {
	db *gorm.DB
}

func NewCommentRepository(db *gorm.DB) CommentRepository {
	return &commentRepository{db: db}
}

func (r *commentRepository) Create(comment *source.Comment) error {
	return r.db.Create(comment).Error
}

func (r *commentRepository) GetByPostID(postID uint) ([]source.Comment, error) {
	var comments []source.Comment
	err := r.db.Preload("User").Where("post_id = ?", postID).Find(&comments).Error
	return comments, err
}

func (r *commentRepository) GetByID(id uint) (*source.Comment, error) {
	var comment source.Comment
	err := r.db.Preload("User").Preload("Post").Where("id = ?", id).First(&comment).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &comment, nil
}

func (r *commentRepository) Update(comment *source.Comment) error {
	return r.db.Save(comment).Error
}

func (r *commentRepository) Delete(id uint) error {
	return r.db.Delete(&source.Comment{}, id).Error
}
