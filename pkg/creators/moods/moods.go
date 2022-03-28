package moods

import (
	"fmt"
	"github.com/digitalmonsters/music/pkg/database"
	"github.com/digitalmonsters/music/pkg/frontend"
	"github.com/pilagod/gorm-cursor-paginator/v2/paginator"
	"github.com/pkg/errors"
	"gopkg.in/guregu/null.v4"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"strings"
	"time"
)

func Upsert(req UpsertRequest, db *gorm.DB) ([]database.Mood, error) {
	tx := db.Begin()
	defer tx.Rollback()

	var moods []database.Mood
	for _, item := range req.Items {
		r := database.Mood{
			Name:      item.Name,
			SortOrder: item.SortOrder,
			CreatedAt: time.Now(),
		}

		if item.Id.Valid {
			r.Id = item.Id.Int64
			r.UpdatedAt = null.TimeFrom(time.Now())
		}

		moods = append(moods, r)
	}

	if err := tx.Model(&moods).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			DoUpdates: clause.AssignmentColumns([]string{"name", "sort_order", "is_active"}),
		}).Create(&moods).Error; err != nil {
		if contain := strings.Contains(err.Error(), "duplicate key value violates unique constraint"); contain {
			return nil, errors.New("mood with the given name has been already created")
		}

		return nil, errors.WithStack(err)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, errors.WithStack(err)
	}

	return moods, nil
}

func AdminList(req ListRequest, db *gorm.DB) (*ListResponse, error) {
	var records []database.Mood
	query := db.Model(records)

	if req.Name.Valid {
		query = query.Where("name ilike ?", fmt.Sprintf("%%%v%%", req.Name.String))
	}

	if req.IsActive.Valid {
		query = query.Where("is_active is ?", req.IsActive.Bool)
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

	var creatorSongsIds []int64
	if err := tx.Model(&database.CreatorSong{}).Where("mood_id in ?", req.Ids).Pluck("id", &creatorSongsIds).Error; err != nil {
		return errors.WithStack(err)
	}

	if len(creatorSongsIds) > 0 {
		return fmt.Errorf("can not delete selected moods, used in songs %v", creatorSongsIds)
	}

	if err := tx.Where("id in ?", req.Ids).Delete(&database.Mood{}).Error; err != nil {
		return errors.WithStack(err)
	}

	return tx.Commit().Error
}

func PublicList(req PublicListRequest, db *gorm.DB) (*PublicListResponse, error) {
	var moods []database.Mood

	query := db.Model(moods).Where("is_active = true")

	if req.Name.Valid && len(req.Name.String) > 0 {
		query = query.Where("name ilike ?", fmt.Sprintf("%%%v%%", req.Name.String))
	}

	paginatorRules := []paginator.Rule{
		{
			Key:   "SortOrder",
			Order: paginator.ASC,
		},
		{
			Key:   "Id",
			Order: paginator.DESC,
		},
	}

	p := paginator.New(
		&paginator.Config{
			Rules: paginatorRules,
			Limit: req.Count,
		},
	)

	if len(req.Cursor) > 0 {
		p.SetAfterCursor(req.Cursor)
	}

	result, cursor, err := p.Paginate(query, &moods)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	if result.Error != nil {
		return nil, errors.WithStack(result.Error)
	}

	resp := &PublicListResponse{
		Items: make([]frontend.Mood, 0),
	}

	if len(moods) == 0 {
		return resp, nil
	}

	resp.Items = frontend.ConvertMoodsToFrontendModel(moods)
	if cursor.After != nil {
		resp.Cursor = *cursor.After
	}

	return resp, nil
}
