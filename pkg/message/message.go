package message

import (
	"fmt"
	"github.com/digitalmonsters/ads-manager/pkg/database"
	"github.com/digitalmonsters/go-common/wrappers/user_go"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"go.elastic.co/apm"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"math"
	"strings"
	"time"
)

func UpsertMessageBulkAdmin(req UpsertMessageAdminRequest, db *gorm.DB) ([]database.Message, error) {
	tx := db.Begin()
	defer tx.Rollback()

	for _, item := range req.Items {
		var duplicates []int64

		query := tx.Model(&database.Message{}).
			Where("countries && ?::text[]", item.Countries).
			Where("verification_status = ?", item.VerificationStatus)

		if item.Id.Valid {
			query = query.Where("id <> ?", item.Id.Int64)
		}

		if item.AgeFrom > 0 && item.AgeTo > 0 {
			query = query.Where(db.Where("int4range(messages.age_from, messages.age_to) && int4range(?,?)", item.AgeFrom, item.AgeTo).
				Or("messages.age_from = ?", item.AgeTo).Or("messages.age_to = ?", item.AgeFrom))
		}

		if item.PointsFrom > 0 && item.PointsTo > 0 {
			query = query.Where(db.Where("numrange(messages.points_from, messages.points_to) && numrange(?,?)", item.PointsFrom, item.PointsTo).
				Or("messages.points_from = ?", item.PointsTo).Or("messages.points_to = ?", item.PointsFrom))
		}

		if err := query.Pluck("id", &duplicates).Error; err != nil {
			return nil, errors.WithStack(err)
		}

		if len(duplicates) > 0 {
			return nil, fmt.Errorf("there are overlaps with other messages %v", duplicates)
		}
	}

	var records []database.Message
	var messagesIds []int64
	for _, item := range req.Items {
		if item.Id.Valid {
			messagesIds = append(messagesIds, item.Id.Int64)
		}
	}

	if err := tx.Where("id in ?", messagesIds).Find(&records).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	exitingRecordsMapped := map[int64]database.Message{}
	for _, r := range records {
		exitingRecordsMapped[r.Id] = r
	}

	records = []database.Message{}
	for _, item := range req.Items {
		r := database.Message{
			Title:              item.Title,
			Description:        item.Description,
			Countries:          item.Countries,
			VerificationStatus: item.VerificationStatus,
			AgeFrom:            item.AgeFrom,
			AgeTo:              item.AgeTo,
			PointsFrom:         item.PointsFrom,
			PointsTo:           item.PointsTo,
			IsActive:           item.IsActive,
		}

		if item.Id.Valid {
			r.Id = item.Id.Int64
			r.UpdatedAt = time.Now()

			if exRecord, ok := exitingRecordsMapped[item.Id.Int64]; ok && exRecord.IsActive != item.IsActive && !item.IsActive {
				t := time.Now()
				r.DeactivatedAt = &t
			}
		}

		records = append(records, r)
	}

	if err := tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		UpdateAll: true,
	}).Create(&records).Error; err != nil {
		if contain := strings.Contains(err.Error(), "duplicate key value violates unique constraint"); contain {
			return nil, errors.New("message with the given name has been already created")
		}

		return nil, errors.WithStack(err)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, errors.WithStack(err)
	}

	return records, nil
}

func DeleteMessagesBulkAdmin(req DeleteMessagesBulkAdminRequest, db *gorm.DB) error {
	tx := db.Begin()
	defer tx.Rollback()

	if err := tx.Where("id in ?", req.Ids).Delete(&database.Message{}).Error; err != nil {
		return errors.WithStack(err)
	}

	return tx.Commit().Error
}

func MessagesListAdmin(req MessagesListAdminRequest, db *gorm.DB) (*MessagesListAdminResponse, error) {
	var records []database.Message
	query := db.Model(&database.Message{})

	if req.Keyword.Valid {
		query = query.Where("title ilike ?", fmt.Sprintf("%%%v%%", req.Keyword.String)).
			Or("description ilike ?", fmt.Sprintf("%%%v%%", req.Keyword.String))
	}

	if req.VerificationStatus > 0 {
		query = query.Where("verification_status = ?", req.VerificationStatus)
	}

	if len(req.Countries) > 0 {
		query = query.Where("countries && ARRAY[?]", req.Countries)
	}

	if req.AgeFrom > 0 {
		query = query.Where("age_from >= ?", req.AgeFrom)
	}

	if req.AgeTo > 0 {
		query = query.Where("age_to <= ?", req.AgeTo)
	}

	if req.PointsFrom > 0 {
		query = query.Where("age_from >= ?", req.PointsFrom)
	}

	if req.PointsTo > 0 {
		query = query.Where("age_to <= ?", req.PointsTo)
	}

	var totalCount int64

	if err := query.Count(&totalCount).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	if err := query.Limit(req.Limit).Offset(req.Offset).
		Order("id desc").Find(&records).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	return &MessagesListAdminResponse{
		Items:      records,
		TotalCount: totalCount,
	}, nil
}

func GetMessageForUser(userId int64, db *gorm.DB, userGoWrapper user_go.IUserGoWrapper, apmTransaction *apm.Transaction) (*NotificationMessage, error) {
	userRespCh := <-userGoWrapper.GetUsersDetails([]int64{userId}, apmTransaction, true)
	if userRespCh.Error != nil {
		return nil, userRespCh.Error.ToError()
	}

	userInfo, ok := userRespCh.Items[userId]
	if !ok {
		return nil, errors.New("user info not found")
	}

	q := fmt.Sprintf("select * from messages where countries && ARRAY['%v'] and verification_status = %v "+
		"and int4range(messages.age_from, messages.age_to) @> %v "+
		"and (numrange(messages.points_from, messages.points_to) @> %.2f OR (points_from = 0 and points_to = 0)) "+
		"and is_active is true and deleted_at is null", userInfo.CountryCode, database.VerificationStatusFromString(userInfo.KycStatus),
		getAge(userInfo.Birthdate.ValueOrZero()), userInfo.VaultPoints.InexactFloat64())
	var records []database.Message
	if err := db.Raw(q).Scan(&records).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	switch len(records) {
	case 0:
		return nil, nil
	case 1:
		return &NotificationMessage{
			Title:       records[0].Title,
			Description: records[0].Description,
		}, nil
	default:
		r := getProperMessage(records, userInfo.VaultPoints)
		return &NotificationMessage{
			Title:       r.Title,
			Description: r.Description,
		}, nil
	}
}

func getProperMessage(records []database.Message, points decimal.Decimal) database.Message {
	for _, r := range records {
		if (r.PointsFrom >= points.InexactFloat64()) && (r.PointsTo <= points.InexactFloat64()) {
			return r
		}
	}

	return records[0]
}

func getAge(birthdate time.Time) int {
	return int(math.Floor(time.Since(birthdate).Hours() / 24 / 365))
}
