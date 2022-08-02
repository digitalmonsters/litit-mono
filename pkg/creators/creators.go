package creators

import (
	"context"
	"fmt"
	"github.com/digitalmonsters/go-common/apm_helper"
	"github.com/digitalmonsters/go-common/error_codes"
	"github.com/digitalmonsters/go-common/eventsourcing"
	"github.com/digitalmonsters/go-common/router"
	"github.com/digitalmonsters/go-common/wrappers/content"
	"github.com/digitalmonsters/go-common/wrappers/music"
	"github.com/digitalmonsters/go-common/wrappers/user_go"
	"github.com/digitalmonsters/music/pkg/database"
	"github.com/digitalmonsters/music/pkg/feed/feed_converter"
	"github.com/digitalmonsters/music/pkg/global"
	"github.com/digitalmonsters/music/utils"
	"github.com/pilagod/gorm-cursor-paginator/v2/paginator"
	"github.com/pkg/errors"
	"github.com/thoas/go-funk"
	"gopkg.in/guregu/null.v4"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type Service struct {
	notifiers     []global.INotifier
	feedConverter *feed_converter.Service
}

func NewService(feedConverter *feed_converter.Service, notifiers []global.INotifier) *Service {
	return &Service{
		feedConverter: feedConverter,
		notifiers:     notifiers,
	}
}

func (s *Service) BecomeMusicCreator(req BecomeMusicCreatorRequest, db *gorm.DB, executionData router.MethodExecutionData, userGoWrapper user_go.IUserGoWrapper) error {
	tx := db.Begin()
	defer tx.Rollback()

	userResp := <-userGoWrapper.GetUsers([]int64{executionData.UserId}, executionData.Context, false)
	if userResp.Error != nil {
		apm_helper.LogError(userResp.Error.ToError(), executionData.Context)
	}

	var user user_go.UserRecord
	if val, ok := userResp.Response[executionData.UserId]; ok {
		user.UserId = val.UserId
		user.Username = val.Username
		user.Firstname = val.Firstname
		user.Lastname = val.Lastname
		user.Email = val.Email
	}

	var creator database.Creator
	if err := tx.Model(creator).Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("user_id = ?", executionData.UserId).Find(&creator).Error; err != nil {
		return errors.WithStack(err)
	}

	if creator.Id > 0 {
		if creator.Status == user_go.CreatorStatusApproved {
			return errors.New("request has been already approved")
		}

		if creator.Status == user_go.CreatorStatusPending {
			return errors.New("request is under consideration")
		}
	}

	creator.UserId = executionData.UserId
	creator.Status = user_go.CreatorStatusPending
	creator.LibraryUrl = req.LibraryLink
	creator.CreatedAt = time.Now()
	creator.Username = user.Username
	creator.Firstname = user.Firstname
	creator.Lastname = user.Lastname
	creator.Email = user.Email

	if err := tx.Save(&creator).Error; err != nil {
		return errors.WithStack(err)
	}

	if err := tx.Commit().Error; err != nil {
		return errors.WithStack(err)
	}

	for _, notifier := range s.notifiers {
		if notifier != nil {
			notifier.Enqueue(creator.Id, &creator)
		}
	}

	return nil
}

func (s *Service) CreatorRequestsList(req CreatorRequestsListRequest, db *gorm.DB, maxThreshold int, ctx context.Context, userGoWrapper user_go.IUserGoWrapper) (*CreatorRequestsListResponse, error) {
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

	if len(req.SearchQuery) > 0 {
		search := fmt.Sprintf("%%%v%%", req.SearchQuery)
		query = query.Where(db.
			Where(utils.AddSearchQuery(db.Table("creators"), []string{search}, []string{"username"})).
			Or(db.
				Where(
					utils.AddSearchQuery(db.Table("creators"), []string{search}, []string{"firstname", "lastname"}),
				),
			).Or(db.
			Where(
				utils.AddSearchQuery(db.Table("creators"), []string{search}, []string{"email"}),
			),
		),
		)
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

	userResp := <-userGoWrapper.GetUsers(userIds, ctx, false)
	if userResp.Error != nil {
		apm_helper.LogError(userResp.Error.ToError(), ctx)
	}

	var respItems []creatorListItem
	for _, c := range creators {
		respItem := creatorListItem{
			Id:         c.Id,
			Status:     c.Status,
			LibraryUrl: c.LibraryUrl,
			UserId:     c.UserId,
			SlaExpired: time.Now().After(c.CreatedAt.Add(time.Hour * time.Duration(maxThreshold))),
			CreatedAt:  c.CreatedAt,
			ApprovedAt: c.ApprovedAt,
			DeletedAt:  c.DeletedAt,
		}

		if c.Reason != nil && c.RejectReason.Valid {
			respItem.RejectReason = null.StringFrom(c.Reason.Reason)
		}

		userModel, ok := userResp.Response[c.UserId]
		if !ok {
			continue
		}

		respItem.UserId = userModel.UserId
		respItem.UserName = userModel.Username
		respItem.FirstName = userModel.Firstname
		respItem.LastName = userModel.Lastname
		respItem.Avatar = userModel.Avatar
		respItem.Email = userModel.Email

		respItems = append(respItems, respItem)
	}

	return &CreatorRequestsListResponse{
		Items:      respItems,
		TotalCount: totalCount,
	}, nil
}

func (s *Service) CreatorRequestApprove(req CreatorRequestApproveRequest, db *gorm.DB) ([]*database.Creator, error) {
	tx := db.Begin()
	defer tx.Rollback()

	var creators []*database.Creator
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id in ? and status = ?", req.Ids, user_go.CreatorStatusPending).
		Find(&creators).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	for _, r := range creators {
		r.Status = user_go.CreatorStatusApproved
		r.ApprovedAt = null.TimeFrom(time.Now())
	}

	if err := tx.Save(&creators).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, errors.WithStack(err)
	}

	for _, notifier := range s.notifiers {
		if notifier != nil {
			for _, creator := range creators {
				notifier.Enqueue(creator.Id, creator)
			}
		}
	}

	return creators, nil
}

func (s *Service) CreatorRequestReject(req CreatorRequestRejectRequest, db *gorm.DB) ([]*database.Creator, error) {
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
		Where("id in ? and status = ?", ids, user_go.CreatorStatusPending).
		Find(&creatorRequests).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	var creatorRequestsIds []int64
	for _, r := range creatorRequests {
		if reasonId, ok := creatorsMapped[r.Id]; ok {
			r.Status = user_go.CreatorStatusRejected
			r.RejectReason = null.IntFrom(reasonId)
			creatorRequestsIds = append(creatorRequestsIds, r.Id)
		}
	}

	if err := tx.Save(&creatorRequests).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	var creatorsWithReason []*database.Creator
	if err := tx.Preload("Reason").Find(&creatorsWithReason, creatorRequestsIds).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, errors.WithStack(err)
	}

	for _, notifier := range s.notifiers {
		if notifier != nil {
			for _, creator := range creatorsWithReason {
				notifier.Enqueue(creator.Id, creator)
			}
		}
	}

	return creatorRequests, nil
}

func (s *Service) UploadNewSong(req UploadNewSongRequest, contentWrapper content.IContentWrapper, db *gorm.DB, executionData router.MethodExecutionData) (*database.CreatorSong, error) {
	tx := db.Begin()
	defer tx.Rollback()

	var creator database.Creator
	if err := tx.Where("user_id = ?", executionData.UserId).Find(&creator).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	if creator.Id == 0 || creator.Status != user_go.CreatorStatusApproved {
		return nil, errors.New("only approved creators can upload music")
	}

	newContent := <-contentWrapper.InsertMusicContent(content.MusicContentRequest{
		ContentType: eventsourcing.ContentTypeMusic,
		Duration:    int(req.FullSongDuration),
		AuthorId:    executionData.UserId,
		Hashtags:    req.Hashtags,
	}, executionData.Context, true)

	if newContent.Error != nil {
		return nil, newContent.Error.ToError()
	}

	song := database.CreatorSong{
		UserId:            executionData.UserId,
		Name:              req.Name,
		Status:            music.CreatorSongStatusPublished,
		LyricAuthor:       req.LyricAuthor,
		MusicAuthor:       req.MusicAuthor,
		FullSongDuration:  req.FullSongDuration,
		ShortSongDuration: req.ShortSongDuration,
		CategoryId:        req.CategoryId,
		MoodId:            req.MoodId,
		FullSongUrl:       req.FullSongUrl,
		ShortSongUrl:      req.ShortSongUrl,
		ImageUrl:          req.ImageUrl,
		Hashtags:          req.Hashtags,
		CreatedAt:         time.Now(),
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

func (s *Service) CheckRequestStatus(userId int64, db *gorm.DB) (*CheckRequestStatusResponse, error) {
	var creator database.Creator

	if err := db.Where("user_id = ?", userId).Preload("Reason").First(&creator).Error; err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("creator request not found")
	} else if err != nil {
		return nil, errors.WithStack(err)
	}

	var reason null.String

	if creator.Reason != nil {
		reason = null.StringFrom(creator.Reason.Reason)
	}

	return &CheckRequestStatusResponse{
		Status:       creator.Status,
		RejectReason: reason,
	}, nil
}

func (s *Service) SongsList(req SongsListRequest, currentUserId int64, db *gorm.DB, executionData router.MethodExecutionData) (*SongsListResponse, *error_codes.ErrorWithCode) {
	var songs []*database.CreatorSong

	q := db.Where("user_id = ?", req.UserId)

	if req.UserId != currentUserId {
		q = q.Where("status != ?", music.CreatorSongStatusRejected)
	}

	paginatorRules := []paginator.Rule{
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

	result, cursor, err := p.Paginate(q, &songs)
	if err != nil {
		return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
	}

	if result.Error != nil {
		return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
	}

	resp := &SongsListResponse{
		Items: s.feedConverter.ConvertToSongModel(songs, currentUserId, executionData.ApmTransaction, executionData.Context),
	}

	if cursor.After != nil {
		resp.Cursor = *cursor.After
	}

	return resp, nil
}
