package ad_campaign

import (
	"context"
	"fmt"
	"github.com/RichardKnop/machinery/v1"
	"github.com/digitalmonsters/ads-manager/configs"
	"github.com/digitalmonsters/ads-manager/pkg/database"
	"github.com/digitalmonsters/ads-manager/utils"
	"github.com/digitalmonsters/go-common/apm_helper"
	"github.com/digitalmonsters/go-common/wrappers/ads_manager"
	"github.com/digitalmonsters/go-common/wrappers/content"
	"github.com/digitalmonsters/go-common/wrappers/go_tokenomics"
	"github.com/digitalmonsters/go-common/wrappers/user_category"
	"github.com/digitalmonsters/go-common/wrappers/user_go"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"go.elastic.co/apm"
	"gopkg.in/guregu/null.v4"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"sort"
	"sync"
	"time"
)

type IService interface {
	CreateAdCampaign(req CreateAdCampaignRequest, userId int64, tx *gorm.DB, ctx context.Context) error
	GetAdsContentForUser(req ads_manager.GetAdsContentForUserRequest, db *gorm.DB, ctx context.Context) (*ads_manager.GetAdsContentForUserResponse, error)
	ClickLink(userId int64, req ClickLinkRequest, tx *gorm.DB) error
	StopAdCampaign(userId int64, req StopAdCampaignRequest, tx *gorm.DB) error
	StartAdCampaign(userId int64, req StartAdCampaignRequest, tx *gorm.DB, ctx context.Context) error
	ListAdCampaigns(userId int64, req ListAdCampaignsRequest, db *gorm.DB, ctx context.Context) (*ListAdCampaignsResponse, error)
	InitTasks() error
}

type service struct {
	contentWrapper      content.IContentWrapper
	userCategoryWrapper user_category.IUserCategoryWrapper
	userWrapper         user_go.IUserGoWrapper
	jobber              *machinery.Server
	goTokenomicsWrapper go_tokenomics.IGoTokenomicsWrapper
}

func NewService(
	contentWrapper content.IContentWrapper,
	userCategoryWrapper user_category.IUserCategoryWrapper,
	userWrapper user_go.IUserGoWrapper,
	jobber *machinery.Server,
	goTokenomicsWrapper go_tokenomics.IGoTokenomicsWrapper,
) IService {
	return &service{
		contentWrapper:      contentWrapper,
		userCategoryWrapper: userCategoryWrapper,
		userWrapper:         userWrapper,
		jobber:              jobber,
		goTokenomicsWrapper: goTokenomicsWrapper,
	}
}

func (s *service) CreateAdCampaign(req CreateAdCampaignRequest, userId int64, tx *gorm.DB, ctx context.Context) error {
	usersTokenomicsInfoResp := <-s.goTokenomicsWrapper.GetUsersTokenomicsInfo([]int64{userId}, nil, ctx, false)
	if usersTokenomicsInfoResp.Error != nil {
		return errors.WithStack(usersTokenomicsInfoResp.Error.ToError())
	}

	userTokenomicsInfo, ok := usersTokenomicsInfoResp.Response[userId]
	if !ok {
		return errors.WithStack(errors.New("user not found"))
	}

	if userTokenomicsInfo.CurrentTokens.LessThan(req.Budget) {
		return errors.WithStack(errors.New("user does not have enough tokens"))
	}

	contentResp := <-s.contentWrapper.GetInternal([]int64{req.ContentId}, false, apm.TransactionFromContext(ctx), false)
	if contentResp.Error != nil {
		return errors.WithStack(contentResp.Error.ToError())
	}

	if _, ok := contentResp.Response[req.ContentId]; !ok {
		return errors.WithStack(errors.New("content not found"))
	}

	var oldAdCampaign database.AdCampaign

	if err := tx.Where("content_id = ? and status in ?", req.ContentId,
		[]database.AdCampaignStatus{database.AdCampaignStatusPending, database.AdCampaignStatusModerated, database.AdCampaignStatusActive}).
		Find(&oldAdCampaign).Error; err != nil {
		return errors.WithStack(err)
	}

	if oldAdCampaign.Id != 0 {
		return errors.WithStack(errors.New("content can participate only in one ad campaign in the same time"))
	}

	price := configs.GetAppConfig().ADS_CAMPAIGN_GLOBAL_PRICE

	if req.Country.Valid {
		var adCampaignCountriesPrice database.AdCampaignCountriesPrice
		if err := tx.Where("country_code = ?", req.Country.String).Find(&adCampaignCountriesPrice).Error; err != nil {
			return errors.WithStack(err)
		}

		if len(adCampaignCountriesPrice.CountryCode) != 0 && !adCampaignCountriesPrice.IsGlobalPrice {
			price = adCampaignCountriesPrice.Price
		}
	}

	adCampaignCategories := make([]*database.AdCampaignCategory, len(req.CategoriesIds))

	if len(req.CategoriesIds) > 0 {
		categoryResp := <-s.contentWrapper.GetCategoryInternal(req.CategoriesIds, nil, 10000, 0,
			null.BoolFrom(false), null.Bool{}, apm.TransactionFromContext(ctx), false, false)
		if categoryResp.Error != nil {
			return errors.WithStack(contentResp.Error.ToError())
		}

		for i, categoryId := range req.CategoriesIds {
			category, ok := lo.Find(categoryResp.Response.Items, func(item content.SimpleCategoryModel) bool {
				return item.Id == categoryId
			})
			if !ok {
				return errors.WithStack(errors.New("content not found"))
			}

			adCampaignCategories[i] = &database.AdCampaignCategory{
				CategoryId:   categoryId,
				CategoryName: category.Name,
			}
		}
	}

	adCampaign := database.AdCampaign{
		UserId:         userId,
		Name:           req.Name,
		AdType:         req.AdType,
		Status:         database.AdCampaignStatusPending,
		ContentId:      req.ContentId,
		Link:           req.Link,
		LinkButtonId:   req.LinkButtonId,
		Country:        req.Country,
		CreatedAt:      time.Now().UTC(),
		DurationMin:    req.DurationMin,
		OriginalBudget: req.Budget,
		Budget:         req.Budget,
		Gender:         req.Gender,
		AgeFrom:        req.AgeFrom,
		AgeTo:          req.AgeTo,
		Price:          price,
	}

	if err := tx.Create(&adCampaign).Error; err != nil {
		return errors.WithStack(err)
	}

	for _, item := range adCampaignCategories {
		item.AdCampaignId = adCampaign.Id

		if err := tx.Create(item).Error; err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}

func (s *service) GetAdsContentForUser(req ads_manager.GetAdsContentForUserRequest, db *gorm.DB, ctx context.Context) (*ads_manager.GetAdsContentForUserResponse, error) {
	adsPerVideos := configs.GetAppConfig().ADS_CAMPAIGN_VIDEOS_PER_CONTENT_VIDEOS

	if adsPerVideos <= 0 {
		return &ads_manager.GetAdsContentForUserResponse{MixedContentIdsWithAd: req.ContentIdsToMix}, nil
	}

	userResp := <-s.userWrapper.GetUserDetails(req.UserId, ctx, false)
	if userResp.Error != nil {
		return nil, errors.WithStack(userResp.Error.ToError())
	}

	user := userResp.Response

	if user.AdDisabled {
		return &ads_manager.GetAdsContentForUserResponse{MixedContentIdsWithAd: req.ContentIdsToMix}, nil
	}

	query := db.Table("ad_campaigns").
		Select("ad_campaigns.id").
		Where("status = ?", database.AdCampaignStatusActive).
		Where("content_id not in ?", append(req.ContentIdsToMix, req.ContentIdsToIgnore...))

	if len(user.CountryCode) == 0 {
		query = query.Where("country is null")
	} else {
		query = query.Where("country = ? or country is null", user.CountryCode)
	}

	if user.Gender.Valid {
		query = query.Where("gender = ? or gender is null", user.Gender.String)
	}

	if user.Birthdate.Valid {
		query = query.Where("(age_from is not null or ? <= now() - interval '1 years' * age_from) and (age_to is not null or ? >= now() - interval '1 years' * age_to)",
			user.Birthdate.Time, user.Birthdate.Time)
	}

	var adCampaignIds []int64
	if err := query.Scan(&adCampaignIds).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	contentIdsToMixLen := len(req.ContentIdsToMix)

	respDataLen := contentIdsToMixLen

	if adsPerVideos > 0 {
		respDataLen = contentIdsToMixLen + contentIdsToMixLen/adsPerVideos
	}

	pageState := ""
	isFirst := true
	limit := 500

	allCategoryAdCampaignIds := make([]int64, 0)
	for {
		if len(pageState) == 0 && !isFirst {
			break
		}

		if isFirst {
			isFirst = false
		}

		userCategoryResp := <-s.userCategoryWrapper.GetInternalUserCategorySubscriptions(req.UserId, limit, pageState, ctx, false)
		if userCategoryResp.Error != nil {
			apm_helper.LogError(errors.WithStack(userCategoryResp.Error.ToError()), ctx)
			break
		}

		pageState = userCategoryResp.Response.PageState

		categoryIdsLen := len(userCategoryResp.Response.CategoryIds)

		if categoryIdsLen < limit {
			pageState = ""
		}

		var categoryAdCampaignIds []int64
		catQuery := db.Table("ad_campaigns").
			Select("distinct ad_campaigns.id").
			Joins("left join ad_campaign_categories on ad_campaign_categories.ad_campaign_id = ad_campaigns.id").
			Where("ad_campaign_categories.category_id is null or ad_campaign_categories.category_id in ?", userCategoryResp.Response.CategoryIds).
			Limit(respDataLen)

		if len(adCampaignIds) > 0 {
			catQuery = catQuery.Where("ad_campaigns.id in ?", adCampaignIds)
		}

		if err := catQuery.Scan(&categoryAdCampaignIds).Error; err != nil {
			apm_helper.LogError(errors.WithStack(err), ctx)
			break
		}

		if len(categoryAdCampaignIds) == 0 {
			continue
		}

		allCategoryAdCampaignIds = append(allCategoryAdCampaignIds, categoryAdCampaignIds...)

		if len(allCategoryAdCampaignIds) >= contentIdsToMixLen {
			break
		}
	}

	if len(allCategoryAdCampaignIds) == 0 {
		return &ads_manager.GetAdsContentForUserResponse{MixedContentIdsWithAd: req.ContentIdsToMix}, nil
	}

	var adCampaignsData []*ads_manager.ContentAd
	if err := db.Table("ad_campaigns").
		Select("ad_campaigns.content_id, ad_campaigns.link, ad_campaigns.link_button_id, action_buttons.name as link_button_name").
		Joins("left join action_buttons on action_buttons.id = ad_campaigns.link_button_id").
		Where("ad_campaigns.status = ?", database.AdCampaignStatusActive).
		Where("ad_campaigns.id in ? or ad_campaigns.content_id in ?", allCategoryAdCampaignIds, req.ContentIdsToMix).
		Find(&adCampaignsData).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	newAd := lo.Filter(adCampaignsData, func(item *ads_manager.ContentAd, _ int) bool {
		_, ok := lo.Find(req.ContentIdsToMix, func(id int64) bool {
			return item.ContentId == id
		})

		return !ok
	})
	sort.Slice(newAd, func(i, j int) bool {
		return newAd[i].ContentId < newAd[j].ContentId
	})

	adCampaignsDataMap := make(map[int64]*ads_manager.ContentAd, len(adCampaignsData))
	for _, v := range adCampaignsData {
		adCampaignsDataMap[v.ContentId] = v
	}

	adCampaignsDataMapLen := len(adCampaignsDataMap)
	adContentIds := make([]int64, len(newAd))
	adContentIdsIter := 0
	for _, item := range newAd {
		adContentIds[adContentIdsIter] = item.ContentId
		adContentIdsIter++
	}
	adContentIdsIter = 0

	respData := make([]int64, respDataLen)
	contentIdsToMixIter := 0

	adIter := 0
	for i := 0; i < respDataLen; i++ {
		if i != 0 && adIter == adsPerVideos && adContentIdsIter < adCampaignsDataMapLen {
			respData[i] = adContentIds[adContentIdsIter]
			adContentIdsIter++
			adIter = 0

			continue
		}

		adIter++

		if contentIdsToMixIter >= contentIdsToMixLen {
			break
		}

		respData[i] = req.ContentIdsToMix[contentIdsToMixIter]
		contentIdsToMixIter++
	}

	contentAds := make(map[int64]*ads_manager.ContentAd, adCampaignsDataMapLen)
	for _, item := range respData {
		val, ok := adCampaignsDataMap[item]
		if !ok {
			continue
		}

		contentAds[val.ContentId] = val
	}

	respData = lo.Filter(respData, func(item int64, _ int) bool {
		return item != 0
	})

	return &ads_manager.GetAdsContentForUserResponse{
		MixedContentIdsWithAd: respData,
		ContentAds:            contentAds,
	}, nil
}

func (s *service) ClickLink(userId int64, req ClickLinkRequest, tx *gorm.DB) error {
	var adCampaign database.AdCampaign
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("content_id = ? and status = ?", req.ContentId, database.AdCampaignStatusActive).
		Find(&adCampaign).Error; err != nil {
		return errors.WithStack(err)
	}

	if adCampaign.Id == 0 || adCampaign.UserId == userId || (adCampaign.EndedAt.Valid && time.Now().UTC().After(adCampaign.EndedAt.Time)) {
		return nil
	}

	var adCampaignClick database.AdCampaignClick
	if err := tx.Where("ad_campaign_id = ? and user_id = ?", adCampaign.Id, userId).Find(&adCampaignClick).Error; err != nil {
		return errors.WithStack(err)
	}

	if adCampaignClick.AdCampaignId != 0 {
		return nil
	}

	adCampaignClick.AdCampaignId = adCampaign.Id
	adCampaignClick.UserId = userId
	adCampaignClick.CreatedAt = time.Now().UTC()

	if err := tx.Create(&adCampaignClick).Error; err != nil {
		return errors.WithStack(err)
	}

	adCampaign.Clicks++

	if err := tx.Exec("update ad_campaigns set clicks = clicks + 1 where id = ?", adCampaign.Id).Error; err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (s *service) StopAdCampaign(userId int64, req StopAdCampaignRequest, tx *gorm.DB) error {
	var adCampaign database.AdCampaign
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ?", req.AdCampaignId).
		Find(&adCampaign).Error; err != nil {
		return errors.WithStack(err)
	}

	if adCampaign.Id == 0 || adCampaign.UserId != userId {
		return errors.WithStack(errors.New("ad campaign not found"))
	}

	if adCampaign.Status != database.AdCampaignStatusActive {
		return errors.WithStack(errors.New("ad campaign is not active"))
	}

	adCampaign.Status = database.AdCampaignStatusCompleted
	if err := tx.Model(&adCampaign).Update("status", adCampaign.Status).Error; err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (s *service) StartAdCampaign(userId int64, req StartAdCampaignRequest, tx *gorm.DB, ctx context.Context) error {
	var adCampaign database.AdCampaign
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ?", req.AdCampaignId).
		Find(&adCampaign).Error; err != nil {
		return errors.WithStack(err)
	}

	if adCampaign.Id == 0 || adCampaign.UserId != userId {
		return errors.WithStack(errors.New("ad campaign not found"))
	}

	if adCampaign.Status != database.AdCampaignStatusModerated {
		return errors.WithStack(errors.New("ad campaign is not moderated"))
	}

	usersTokenomicsInfoResp := <-s.goTokenomicsWrapper.GetUsersTokenomicsInfo([]int64{userId}, nil, ctx, false)
	if usersTokenomicsInfoResp.Error != nil {
		return errors.WithStack(usersTokenomicsInfoResp.Error.ToError())
	}

	userTokenomicsInfo, ok := usersTokenomicsInfoResp.Response[userId]
	if !ok {
		return errors.WithStack(errors.New("user not found"))
	}

	if userTokenomicsInfo.CurrentTokens.LessThan(adCampaign.Budget) {
		return errors.WithStack(errors.New("user does not have enough tokens"))
	}

	adCampaign.Status = database.AdCampaignStatusActive
	adCampaign.StartedAt = null.TimeFrom(time.Now().UTC())

	if err := tx.Model(&adCampaign).
		Update("started_at", adCampaign.StartedAt.Time).
		Update("status", adCampaign.Status).Error; err != nil {
		return errors.WithStack(err)
	}

	if adCampaign.DurationMin > 0 {
		adCampaign.EndedAt = null.TimeFrom(adCampaign.StartedAt.Time.Add(time.Duration(adCampaign.DurationMin) * time.Minute))

		if err := tx.Model(&adCampaign).
			Update("ended_at", adCampaign.EndedAt.Time).Error; err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}

func (s *service) ListAdCampaigns(userId int64, req ListAdCampaignsRequest, db *gorm.DB, ctx context.Context) (*ListAdCampaignsResponse, error) {
	adCampaigns := make([]database.AdCampaign, 0)

	query := db.Model(adCampaigns).Where("user_id = ?", userId).Order("created_at desc")

	if req.Status != nil {
		query = query.Where("ad_campaigns.status = ?", req.Status)
	}

	if req.Name.Valid {
		search := fmt.Sprintf("%%%v%%", req.Name.String)
		query = query.Where("ad_campaigns.name ilike ?", search)
	}

	if req.Age.Valid {
		query = query.Where("ad_campaigns.age_from >= ? and ad_campaigns.age_to <= ?", req.Age.Int64, req.Age.Int64)
	}

	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, err
	}
	if req.Limit > 0 {
		query = query.Limit(req.Limit)
	}

	if err := query.Offset(req.Offset).Find(&adCampaigns).Error; err != nil {
		return nil, err
	}

	adCampaignsLen := len(adCampaigns)
	items := make([]*ListAdCampaignsResponseItem, adCampaignsLen)

	if adCampaignsLen == 0 {
		return &ListAdCampaignsResponse{
			Items:      items,
			TotalCount: null.IntFrom(totalCount),
		}, nil
	}

	contentIds := make([]int64, adCampaignsLen)
	for i, adCampaign := range adCampaigns {
		items[i] = &ListAdCampaignsResponseItem{
			AdCampaignId: adCampaign.Id,
			Content: ListAdCampaignsResponseItemContent{
				Content: content.SimpleContent{Id: adCampaign.ContentId},
			},
			Views:          adCampaign.Views,
			Clicks:         adCampaign.Clicks,
			Status:         adCampaign.Status,
			Budget:         adCampaign.Budget,
			OriginalBudget: adCampaign.OriginalBudget,
		}
		contentIds[i] = adCampaign.ContentId
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer func() {
			wg.Done()
		}()

		contentsResp := <-s.contentWrapper.GetInternal(contentIds, false, apm.TransactionFromContext(ctx), false)
		if contentsResp.Error != nil {
			apm_helper.LogError(errors.WithStack(contentsResp.Error.ToError()), ctx)
			return
		}

		for _, item := range items {
			c, ok := contentsResp.Response[item.Content.Content.Id]
			if !ok {
				continue
			}

			item.Content = ListAdCampaignsResponseItemContent{
				Content:    c,
				AnimUrl:    utils.GetAnimUrl(c.VideoId),
				Thumbnail:  utils.GetThumbnailUrl(c.VideoId),
				VideoUrl:   utils.GetVideoUrl(c.VideoId),
				IsVertical: utils.GetIsVertical(c.Width, c.Height),
			}
		}
	}()

	go func() {
		defer func() {
			wg.Done()
		}()

		if !req.DateFrom.Valid && !req.DateTo.Valid {
			return
		}

		var wgStats sync.WaitGroup

		for _, item := range items {
			wgStats.Add(2)

			itemCopy := item

			go func() {
				defer func() {
					wgStats.Done()
				}()

				var views int64
				subQuery := db.Table("ad_campaign_views").Where("ad_campaign_id = ?", itemCopy.AdCampaignId)

				if req.DateFrom.Valid {
					subQuery = subQuery.Where("created_at >= ?", req.DateFrom.Time)
				}

				if req.DateTo.Valid {
					subQuery = subQuery.Where("created_at <= ?", req.DateTo.Time)
				}

				if err := subQuery.Count(&views).Error; err != nil {
					apm_helper.LogError(errors.WithStack(err), ctx)
				}

				itemCopy.Views = int(views)
			}()

			go func() {
				defer func() {
					wgStats.Done()
				}()

				var clicks int64
				subQuery := db.Table("ad_campaign_clicks").Where("ad_campaign_id = ?", itemCopy.AdCampaignId)

				if req.DateFrom.Valid {
					subQuery = subQuery.Where("created_at >= ?", req.DateFrom.Time)
				}

				if req.DateTo.Valid {
					subQuery = subQuery.Where("created_at <= ?", req.DateTo.Time)
				}

				if err := subQuery.Count(&clicks).Error; err != nil {
					apm_helper.LogError(errors.WithStack(err), ctx)
				}

				itemCopy.Clicks = int(clicks)
			}()
		}

		wgStats.Wait()
	}()

	wg.Wait()

	return &ListAdCampaignsResponse{
		Items:      items,
		TotalCount: null.IntFrom(totalCount),
	}, nil
}
