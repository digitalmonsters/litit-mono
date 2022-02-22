package categories

import (
	"fmt"
	"github.com/digitalmonsters/music/pkg/database"
	"github.com/pkg/errors"
	"gopkg.in/guregu/null.v4"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"strings"
	"time"
)

func Upsert(req UpsertRequest, db *gorm.DB) ([]database.Category, error) {
	tx := db.Begin()
	defer tx.Rollback()

	var categories []database.Category
	for _, item := range req.Items {
		r := database.Category{
			Name:      item.Name,
			SortOrder: item.SortOrder,
			CreatedAt: time.Now(),
		}

		if item.Id.Valid {
			r.Id = item.Id.Int64
			r.UpdatedAt = null.TimeFrom(time.Now())
		}

		categories = append(categories, r)
	}

	if err := tx.Model(&categories).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			DoUpdates: clause.AssignmentColumns([]string{"name", "sort_order"}),
		}).Create(&categories).Error; err != nil {
		if contain := strings.Contains(err.Error(), "duplicate key value violates unique constraint"); contain {
			return nil, errors.New("category with the given name has been already created")
		}

		return nil, errors.WithStack(err)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, errors.WithStack(err)
	}

	return categories, nil
}

func List(req ListRequest, db *gorm.DB) (*ListResponse, error) {
	var records []database.Category
	query := db.Model(records)

	if req.Name.Valid {
		query = query.Where("name ilike ?", fmt.Sprintf("%%%v%%", req.Name.String))
	}

	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	switch req.SortOption {
	case SortOptionSortOrderDesc:
		query = query.Order("sort_order desc")
	case SortOptionSortOrderAsc:
		query = query.Order("sort_order asc")
	case SortOptionSongsCountDesc:
		query = query.Order("songs_count desc")
	case SortOptionSongsCountAsc:
		query = query.Order("songs_count asc")
	}

	if err := query.Limit(req.Limit).Offset(req.Offset).Order("id desc").Find(&records).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	return &ListResponse{
		Items:      records,
		TotalCount: totalCount,
	}, nil
}

func Delete(req DeleteRequest, db *gorm.DB) error {
	tx := db.Begin()
	defer tx.Rollback()

	//todo: before deletion check songs which used these categories
	if err := tx.Where("id in ?", req.Ids).Delete(&database.Category{}).Error; err != nil {
		return errors.WithStack(err)
	}

	return tx.Commit().Error
}
