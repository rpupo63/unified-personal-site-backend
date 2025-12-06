package models

import "github.com/google/uuid"

// ProjectTag represents a tag associated with a project
type ProjectTag struct {
	ID        uuid.UUID `json:"id" db:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid();not null"`
	ProjectID uuid.UUID `json:"project_id" db:"project_id" gorm:"type:uuid;not null;index:idx_project_tag_project_id;uniqueIndex:idx_project_tag_unique;constraint:OnDelete:CASCADE"`
	Value     string    `json:"value" db:"value" gorm:"type:text;not null;uniqueIndex:idx_project_tag_unique"`

	Project Project `json:"project,omitempty" gorm:"foreignKey:ProjectID;references:ID"`
}
