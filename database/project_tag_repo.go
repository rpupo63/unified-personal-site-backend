package database

import (
	"github.com/google/uuid"
	"github.com/rpupo63/unified-personal-site-backend/models"
	"gorm.io/gorm"
)

type ProjectTagRepo struct {
	db *gorm.DB
}

func NewProjectTagRepo(db *gorm.DB) *ProjectTagRepo {
	return &ProjectTagRepo{db}
}

// GetDB returns the underlying database connection for debugging purposes
func (r *ProjectTagRepo) GetDB() *gorm.DB {
	return r.db
}

// FindAll returns all project tags from the database
func (r *ProjectTagRepo) FindAll() ([]*models.ProjectTag, error) {
	var projectTags []*models.ProjectTag
	err := r.db.Find(&projectTags).Error
	return projectTags, err
}

// Add inserts a new project tag into the database
func (r *ProjectTagRepo) Add(projectTag *models.ProjectTag) error {
	return r.db.Create(projectTag).Error
}

// Delete removes a project tag from the database by id
func (r *ProjectTagRepo) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.ProjectTag{}, id).Error
}
