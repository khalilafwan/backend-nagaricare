package entity

import (
	"encoding/json"
	"time"
)

// Post represents a post made by a user
type Post struct {
	ID_Posts     int       `json:"id_posts"` // Primary key of the post
	Title        string    `json:"title"`    // Title of the forum post
	Content      string    `json:"content"`  // Content of the forum post
	ID_user      int       `json:"id_user"`
	CreatedAt    time.Time `json:"-"`
	CreatedAtStr string    `json:"created_at"` // Will hold the formatted date
}

// MarshalJSON formats the CreatedAt field
func (p *Post) MarshalJSON() ([]byte, error) {
	type Alias Post
	return json.Marshal(&struct {
		*Alias
		CreatedAtStr string `json:"created_at"`
	}{
		Alias: (*Alias)(p),
		CreatedAtStr: func() string {
			if p.CreatedAt.IsZero() {
				return ""
			}
			return p.CreatedAt.Format("2006-01-02 15:04:05")
		}(),
	})
}
