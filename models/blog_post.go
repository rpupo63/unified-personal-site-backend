package models

import (
	"time"

	"github.com/google/uuid"
)

// BlogPost represents a complete blog post with metadata
type BlogPost struct {
	ID         uuid.UUID  `json:"id" db:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid();not null"`
	Title      string     `json:"title" db:"title" gorm:"type:text;not null;unique"`
	Summary    *string    `json:"summary,omitempty" db:"summary" gorm:"type:text"`
	Content    string     `json:"content" db:"content" gorm:"type:text;not null"`
	DateAdded  time.Time  `json:"dateAdded" db:"date_added" gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP"`
	DateEdited *time.Time `json:"dateEdited,omitempty" db:"date_edited" gorm:"type:timestamp"`
	Length     int        `json:"length" db:"length" gorm:"type:integer;not null;default:0"`
	URL        *string    `json:"url,omitempty" db:"url" gorm:"type:text"`
	Tags       []BlogTag  `json:"tags,omitempty" gorm:"foreignKey:BlogPostID;references:ID;constraint:OnDelete:CASCADE"`
}
