package database

import (
	"github.com/google/uuid"
	"github.com/rpupo63/unified-personal-site-backend/models"
	"gorm.io/gorm"
)

type ProjectRepo struct {
	db *gorm.DB
}

func NewProjectRepo(db *gorm.DB) *ProjectRepo {
	return &ProjectRepo{db}
}

// GetDB returns the underlying database connection for debugging purposes
func (r *ProjectRepo) GetDB() *gorm.DB {
	return r.db
}

// FindAll returns all projects from the database
func (r *ProjectRepo) FindAll() ([]*models.Project, error) {
	var projects []*models.Project
	err := r.db.Preload("Tags").Find(&projects).Error
	return projects, err
}

// FindByID returns a project by its ID
func (r *ProjectRepo) FindByID(id uuid.UUID) (*models.Project, error) {
	var project models.Project
	err := r.db.Preload("Tags").First(&project, id).Error
	if err != nil {
		return nil, err
	}
	return &project, nil
}

// Add inserts a new project into the database
func (r *ProjectRepo) Add(project *models.Project) error {
	return r.db.Create(project).Error
}

// Update updates an existing project in the database
func (r *ProjectRepo) Update(project *models.Project) error {
	return r.db.Save(project).Error
}

// Delete removes a project from the database by id
func (r *ProjectRepo) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Project{}, id).Error
}
