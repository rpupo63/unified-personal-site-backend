package models

import "github.com/google/uuid"

// Project represents a complete project with metadata
type Project struct {
	ID          uuid.UUID    `json:"id" db:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid();not null"`
	Title       string       `json:"title" db:"title" gorm:"type:text;not null;unique"`
	Description string       `json:"description" db:"description" gorm:"type:text;not null"`
	GithubLink  string       `json:"github_link" db:"github_link" gorm:"type:text;not null"`
	DemoLink    string       `json:"demo_link" db:"demo_link" gorm:"type:text;not null"`
	Type        string       `json:"type" db:"type" gorm:"type:text;not null"`
	GifLink     *string      `json:"gif_link,omitempty" db:"gif_link" gorm:"type:text"`
	Tags        []ProjectTag `json:"tags,omitempty" gorm:"foreignKey:ProjectID;references:ID;constraint:OnDelete:CASCADE"`
}
