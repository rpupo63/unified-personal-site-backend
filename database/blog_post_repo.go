package database

import (
	"github.com/ProNexus-Startup/ProNexus/backend/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BlogPostRepo struct {
	db *gorm.DB
}

func NewBlogPostRepo(db *gorm.DB) *BlogPostRepo {
	return &BlogPostRepo{db}
}

// GetDB returns the underlying database connection for debugging purposes
func (r *BlogPostRepo) GetDB() *gorm.DB {
	return r.db
}

// FindAll returns all blog posts from the database
func (r *BlogPostRepo) FindAll() ([]*models.BlogPost, error) {
	var blogPosts []*models.BlogPost
	err := r.db.Preload("Tags").Find(&blogPosts).Error
	return blogPosts, err
}

// FindByID returns a blog post by its ID
func (r *BlogPostRepo) FindByID(id uuid.UUID) (*models.BlogPost, error) {
	var blogPost models.BlogPost
	err := r.db.Preload("Tags").First(&blogPost, id).Error
	if err != nil {
		return nil, err
	}
	return &blogPost, nil
}

// Add inserts a new blog post into the database
func (r *BlogPostRepo) Add(blogPost *models.BlogPost) error {
	return r.db.Create(blogPost).Error
}

// Update updates an existing blog post in the database
func (r *BlogPostRepo) Update(blogPost *models.BlogPost) error {
	return r.db.Save(blogPost).Error
}

// Delete removes a blog post from the database by id
func (r *BlogPostRepo) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.BlogPost{}, id).Error
}

