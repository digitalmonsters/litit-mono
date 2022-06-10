package template

import (
	"fmt"
	"github.com/digitalmonsters/notification-handler/pkg/database"
	"github.com/pkg/errors"
	"gopkg.in/guregu/null.v4"
	"gorm.io/gorm"
	"time"
)

type IService interface {
	EditTemplate(req EditTemplateRequest, tx *gorm.DB) error
	ListTemplates(req ListTemplatesRequest, db *gorm.DB) (*ListTemplatesResponse, error)
}

type service struct {
}

func NewService() IService {
	return &service{}
}

func (s service) EditTemplate(req EditTemplateRequest, tx *gorm.DB) error {
	var template database.RenderTemplate

	if err := tx.Where("id = ?", req.Id).Find(&template).Error; err != nil {
		return errors.WithStack(err)
	}

	if len(template.Id) == 0 {
		return errors.WithStack(errors.New("template not found"))
	}

	template.Title = req.Title
	template.Body = req.Body
	template.Headline = req.Headline
	template.Kind = req.Kind
	template.Route = req.Route
	template.ImageUrl = req.ImageUrl
	template.UpdatedAt = time.Now().UTC()

	if err := tx.Save(&template).Error; err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (s service) ListTemplates(req ListTemplatesRequest, db *gorm.DB) (*ListTemplatesResponse, error) {
	var templates []database.RenderTemplate

	query := db.Model(&templates)

	if len(req.Id) > 0 {
		search := fmt.Sprintf("%%%v%%", req.Id)
		query = query.Where("render_templates.id ilike ?", search)
	}

	if len(req.Title) > 0 {
		search := fmt.Sprintf("%%%v%%", req.Title)
		query = query.Where("render_templates.title ilike ?", search)
	}

	if len(req.Body) > 0 {
		search := fmt.Sprintf("%%%v%%", req.Body)
		query = query.Where("render_templates.body ilike ?", search)
	}

	if len(req.Headline) > 0 {
		search := fmt.Sprintf("%%%v%%", req.Headline)
		query = query.Where("render_templates.headline ilike ?", search)
	}

	if len(req.Kind) > 0 {
		search := fmt.Sprintf("%%%v%%", req.Kind)
		query = query.Where("render_templates.kind ilike ?", search)
	}

	if len(req.Route) > 0 {
		search := fmt.Sprintf("%%%v%%", req.Route)
		query = query.Where("render_templates.route ilike ?", search)
	}

	if len(req.ImageUrl) > 0 {
		search := fmt.Sprintf("%%%v%%", req.ImageUrl)
		query = query.Where("render_templates.image_url ilike ?", search)
	}

	if req.CreatedAtFrom.Valid {
		query = query.Where("render_templates.created_at >= ?", req.CreatedAtFrom.ValueOrZero())
	}

	if req.CreatedAtTo.Valid {
		query = query.Where("render_templates.created_at <= ?", req.CreatedAtTo.ValueOrZero())
	}

	if req.UpdatedAtFrom.Valid {
		query = query.Where("render_templates.updated_at >= ?", req.UpdatedAtFrom.ValueOrZero())
	}

	if req.UpdatedAtTo.Valid {
		query = query.Where("render_templates.updated_at <= ?", req.UpdatedAtTo.ValueOrZero())
	}

	if sortingArr := req.Sorting; len(sortingArr) > 0 {
		for _, sorting := range sortingArr {
			sortOrder := " asc"
			if !sorting.IsAscending {
				sortOrder = " desc"
			}
			query = query.Order(string(sorting.Field) + sortOrder)
		}
	}

	var totalCount int64
	if req.Offset == 0 {
		if err := query.Count(&totalCount).Error; err != nil {
			return nil, errors.WithStack(err)
		}
	}

	if req.Limit > 0 {
		query = query.Limit(req.Limit)
	}

	if err := query.Offset(req.Offset).Find(&templates).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	respItems := make([]*ListTemplateItem, len(templates))
	for i, template := range templates {
		respItems[i] = &ListTemplateItem{
			Id:       template.Id,
			Title:    template.Title,
			Body:     template.Body,
			Headline: template.Headline,
			Kind:     template.Kind,
			Route:    template.Route,
			ImageUrl: template.ImageUrl,
		}
	}

	return &ListTemplatesResponse{
		Items:      respItems,
		TotalCount: null.IntFrom(totalCount),
	}, nil
}
