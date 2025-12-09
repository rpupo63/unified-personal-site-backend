package database

import (
	"github.com/rpupo63/unified-personal-site-backend/models"
	"gorm.io/gorm"
)

type BlogTagRepo struct {
	db *gorm.DB
}

func NewBlogTagRepo(db *gorm.DB) *BlogTagRepo {
	return &BlogTagRepo{db}
}

// GetDB returns the underlying database connection for debugging purposes
func (r *BlogTagRepo) GetDB() *gorm.DB {
	return r.db
}

// FindAll returns all blog tags from the database
func (r *BlogTagRepo) FindAll() ([]*models.BlogTag, error) {
	var blogTags []*models.BlogTag
	err := r.db.Find(&blogTags).Error
	return blogTags, err
}

// Add inserts a new blog tag into the database
func (r *BlogTagRepo) Add(blogTag *models.BlogTag) error {
	return r.db.Create(blogTag).Error
}

// Delete removes a blog tag from the database by blog_id and value
func (r *BlogTagRepo) Delete(blogID string, value string) error {
	return r.db.Where("blog_id = ? AND value = ?", blogID, value).Delete(&models.BlogTag{}).Error
}
