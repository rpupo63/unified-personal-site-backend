package models

import "github.com/google/uuid"

// BlogTag represents a tag associated with a blog post
type BlogTag struct {
	ID         uuid.UUID `json:"id" db:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid();not null"`
	BlogPostID uuid.UUID `json:"blog_post_id" db:"blog_post_id" gorm:"type:uuid;not null;index:idx_blog_tag_blog_post_id;uniqueIndex:idx_blog_tag_unique"`
	Value      string    `json:"value" db:"value" gorm:"type:text;not null;uniqueIndex:idx_blog_tag_unique"`

	BlogPost BlogPost `json:"blog_post,omitempty" gorm:"foreignKey:BlogPostID;references:ID"`
}
