package database

import (
	"github.com/ProNexus-Startup/ProNexus/backend/errs"
	"gorm.io/gorm"
)

type Database struct {
	blogPostRepo   *BlogPostRepo
	blogTagRepo    *BlogTagRepo
	projectRepo    *ProjectRepo
	projectTagRepo *ProjectTagRepo
}

// New initializes a new Database struct with each repository using a shared GORM database instance
func New(db *gorm.DB) Database {
	return Database{
		blogPostRepo:   NewBlogPostRepo(db),
		blogTagRepo:    NewBlogTagRepo(db),
		projectRepo:    NewProjectRepo(db),
		projectTagRepo: NewProjectTagRepo(db),
	}
}

// Accessor methods for each repository

func (d Database) BlogPostRepo() *BlogPostRepo {
	return d.blogPostRepo
}

func (d Database) BlogTagRepo() *BlogTagRepo {
	return d.blogTagRepo
}

func (d Database) ProjectRepo() *ProjectRepo {
	return d.projectRepo
}

func (d Database) ProjectTagRepo() *ProjectTagRepo {
	return d.projectTagRepo
}

func (d Database) MigrateStep(migrationDir string, steps int) error {
	if migrationDir == "" {
		return errs.BadRequest("migration directory cannot be empty")
	}
	if steps == 0 {
		return errs.BadRequest("steps cannot be zero")
	}
	// Migration logic would go here
	return nil
}
