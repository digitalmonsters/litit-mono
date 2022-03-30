package reject_reasons

import (
	"github.com/digitalmonsters/music/pkg/database"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

func Upsert(req UpsertRequest, db *gorm.DB) ([]database.CreatorRejectReasons, error) {
	tx := db.Begin()
	defer tx.Rollback()

	var records []database.CreatorRejectReasons
	for _, item := range req.Items {
		r := database.CreatorRejectReasons{
			Reason:    item.Reason,
			Type:      item.Type,
			CreatedAt: time.Now(),
		}

		if item.Id.Valid {
			r.Id = item.Id.Int64
		}

		records = append(records, r)
	}

	if err := tx.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(&records).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, errors.WithStack(err)
	}

	return records, nil
}

func List(req ListRequest, db *gorm.DB) (*ListResponse, error) {
	var records []database.CreatorRejectReasons
	query := db.Model(records)

	if req.Type > 0 {
		query = query.Where("type = ?", req.Type)
	}

	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, errors.WithStack(err)
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

	if err := tx.Where("id in ?", req.Ids).Delete(&database.CreatorRejectReasons{}).Error; err != nil {
		return errors.WithStack(err)
	}

	return tx.Commit().Error
}
