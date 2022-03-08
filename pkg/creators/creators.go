package creators

import (
	"fmt"
	"github.com/digitalmonsters/go-common/apm_helper"
	"github.com/digitalmonsters/go-common/router"
	"github.com/digitalmonsters/go-common/wrappers/user"
	"github.com/digitalmonsters/music/pkg/database"
	"github.com/pkg/errors"
	"github.com/thoas/go-funk"
	"go.elastic.co/apm"
	"gopkg.in/guregu/null.v4"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

func BecomeMusicCreator(req BecomeMusicCreatorRequest, db *gorm.DB, executionData router.MethodExecutionData) error {
	tx := db.Begin()
	defer tx.Rollback()

	var creator database.Creator
	if err := tx.Model(creator).Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("user_id = ?", executionData.UserId).Find(&creator).Error; err != nil {
		return errors.WithStack(err)
	}

	if creator.Id > 0 {
		if creator.Status == database.CreatorStatusApproved {
			return errors.New("request has been already approved")
		}

		if creator.Status == database.CreatorStatusPending {
			return errors.New("request is under consideration")
		}
	}

	creator.UserId = executionData.UserId
	creator.Status = database.CreatorStatusPending
	creator.LibraryUrl = req.LibraryLink
	creator.CreatedAt = time.Now()

	if err := tx.Save(&creator).Error; err != nil {
		return errors.WithStack(err)
	}

	return tx.Commit().Error
}

func CreatorRequestsList(req CreatorRequestsListRequest, db *gorm.DB, maxThreshold int, apmTransaction *apm.Transaction, userWrapper user.IUserWrapper) (*CreatorRequestsListResponse, error) {
	query := db.Model(database.Creator{}).Preload("Reason")

	if req.UserId.Valid {
		query = query.Where("user_id = ?", req.UserId.Int64)
	}

	if len(req.Statuses) > 0 {
		query = query.Where("status in ?", req.Statuses)
	}

	if req.MaxThresholdExceeded {
		query = query.Where("created_at <= NOW() - INTERVAL ?", gorm.Expr(fmt.Sprintf("'%v HOURS'", maxThreshold)))
	}

	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	if req.OrderOption > 0 {
		switch req.OrderOption {
		case OrderOptionCreatedAtDesc:
			query = query.Order("created_at desc")
		case OrderOptionCreatedAtAsc:
			query = query.Order("created_at asc")
		}
	}

	var creators []database.Creator
	if err := query.Order("id desc").Limit(req.Limit).Offset(req.Offset).Find(&creators).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	var userIds []int64

	for _, c := range creators {
		if !funk.ContainsInt64(userIds, c.UserId) {
			userIds = append(userIds, c.UserId)
		}
	}

	userResp := <-userWrapper.GetUsers(userIds, apmTransaction, false)
	if userResp.Error != nil {
		apm_helper.CaptureApmError(userResp.Error.ToError(), apmTransaction)
	}

	var respItems []creatorListItem
	for _, c := range creators {
		respItem := creatorListItem{
			Id:         c.Id,
			Status:     c.Status,
			LibraryUrl: c.LibraryUrl,
			UserId:     c.UserId,
			CreatedAt:  c.CreatedAt,
			ApprovedAt: c.ApprovedAt,
			DeletedAt:  c.DeletedAt,
		}

		if c.Reason != nil && c.RejectReason.Valid {
			respItem.RejectReason = null.StringFrom(c.Reason.Reason)
		}

		userModel, ok := userResp.Items[c.UserId]
		if !ok {
			continue
		}

		respItem.UserId = userModel.UserId
		respItem.UserName = userModel.Username
		respItem.FirstName = userModel.Firstname
		respItem.LastName = userModel.Lastname
		respItem.Avatar = userModel.Avatar

		respItems = append(respItems, respItem)
	}

	return &CreatorRequestsListResponse{
		Items:      respItems,
		TotalCount: totalCount,
	}, nil
}

func CreatorRequestApprove(req CreatorRequestApproveRequest, db *gorm.DB) ([]*database.Creator, error) {
	tx := db.Begin()
	defer tx.Rollback()

	var creators []*database.Creator
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id in ? and status = ?", req.Ids, database.CreatorStatusPending).
		Find(&creators).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	for _, r := range creators {
		r.Status = database.CreatorStatusApproved
		r.ApprovedAt = null.TimeFrom(time.Now())
	}

	//todo: add user to creators table

	if err := tx.Save(&creators).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, errors.WithStack(err)
	}

	//todo: kafka notifier
	//todo: k8s

	return creators, nil
}

func CreatorRequestReject(req CreatorRequestRejectRequest, db *gorm.DB) ([]*database.Creator, error) {
	tx := db.Begin()
	defer tx.Rollback()

	var ids []int64
	creatorsMapped := map[int64]int64{}

	for _, item := range req.Items {
		ids = append(ids, item.Id)
		creatorsMapped[item.Id] = item.Reason
	}

	var creatorRequests []*database.Creator
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id in ? and status = ?", ids, database.CreatorStatusPending).
		Find(&creatorRequests).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	for _, r := range creatorRequests {
		if reasonId, ok := creatorsMapped[r.Id]; ok {
			r.Status = database.CreatorStatusRejected
			r.RejectReason = null.IntFrom(reasonId)
		}
	}

	if err := tx.Save(&creatorRequests).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, errors.WithStack(err)
	}

	//todo: kafka notifier

	return creatorRequests, nil
}

func UploadNewSong(req UploadNewSongRequest, db *gorm.DB, executionData router.MethodExecutionData) (*database.CreatorSong, error) {
	tx := db.Begin()
	defer tx.Rollback()

	var creator database.Creator
	if err := tx.Where("user_id = ?", executionData.UserId).Find(&creator).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	if creator.Id == 0 || creator.Status != database.CreatorStatusApproved {
		return nil, errors.New("only approved creators can upload music")
	}

	song := database.CreatorSong{
		UserId:       executionData.UserId,
		Name:         req.Name,
		Status:       database.CreatorSongStatusPending,
		LyricAuthor:  req.LyricAuthor,
		MusicAuthor:  req.MusicAuthor,
		CategoryId:   req.CategoryId,
		FullSongUrl:  req.FullSongUrl,
		ShortSongUrl: req.ShortSongUrl,
		ImageUrl:     req.ImageUrl,
		Hashtags:     req.Hashtags,
		CreatedAt:    time.Now(),
	}

	if req.Id.Valid {
		song.Id = req.Id.Int64
		song.UpdatedAt = null.TimeFrom(time.Now())
	}

	if err := tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"name", "lyric_author", "music_author", "category_id", "hashtags"}),
	}).Create(&song).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, errors.WithStack(err)
	}

	return &song, nil
}
