package database

import "time"

type RenderTemplate struct {
	Id        string
	Title     string
	Body      string
	Headline  string
	Kind      string
	IsGrouped bool
	Route     string
	ImageUrl  string

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (RenderTemplate) TableName() string {
	return "render_templates"
}
