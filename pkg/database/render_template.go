package database

import "time"

type RenderTemplate struct {
	Id        string `json:"id"`
	Kind      string `json:"kind"`
	IsGrouped bool   `json:"is_grouped"`
	Route     string `json:"route"`
	ImageUrl  string `json:"image_url"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (RenderTemplate) TableName() string {
	return "render_templates"
}
