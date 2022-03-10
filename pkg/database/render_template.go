package database

import "time"

type RenderTemplate struct {
	Id    string
	Title string
	Body  string

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (RenderTemplate) TableName() string {
	return "render_templates"
}
