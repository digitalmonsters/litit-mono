package template

import "gopkg.in/guregu/null.v4"

type EditTemplateRequest struct {
	Id string `json:"id"`
	//Title    string `json:"title"`
	//Body     string `json:"body"`
	//Headline string `json:"headline"`
	Kind     string `json:"kind"`
	Route    string `json:"route"`
	ImageUrl string `json:"image_url"`
	Muted    bool   `json:"muted"`
}

type SortField string

type Sorting struct {
	Field       SortField `json:"field"`
	IsAscending bool      `json:"is_ascending"`
}

type ListTemplatesRequest struct {
	Id            string    `json:"id"`
	Title         string    `json:"title"`
	Body          string    `json:"body"`
	Headline      string    `json:"headline"`
	Kind          string    `json:"kind"`
	Route         string    `json:"route"`
	ImageUrl      string    `json:"image_url"`
	CreatedAtFrom null.Time `json:"created_at_from"`
	CreatedAtTo   null.Time `json:"created_at_to"`
	UpdatedAtFrom null.Time `json:"updated_at_from"`
	UpdatedAtTo   null.Time `json:"updated_at_to"`
	Muted         null.Bool `json:"muted"`
	Sorting       []Sorting `json:"sorting"`
	Limit         int       `json:"limit"`
	Offset        int       `json:"offset"`
}

type ListTemplatesResponse struct {
	Items      []*ListTemplateItem `json:"items"`
	TotalCount null.Int            `json:"total_count"`
}

type ListTemplateItem struct {
	Id       string `json:"id"`
	Title    string `json:"title"`
	Body     string `json:"body"`
	Headline string `json:"headline"`
	Kind     string `json:"kind"`
	Route    string `json:"route"`
	ImageUrl string `json:"image_url"`
	Muted    bool   `json:"muted"`
}
