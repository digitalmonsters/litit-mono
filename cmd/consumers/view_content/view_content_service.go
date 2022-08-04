package view_content

import (
	"context"
	"github.com/digitalmonsters/ads-manager/pkg/database"
	"github.com/digitalmonsters/go-common/apm_helper"
	"github.com/digitalmonsters/go-common/eventsourcing"
	"github.com/digitalmonsters/go-common/wrappers/go_tokenomics"
	"github.com/pkg/errors"
	"github.com/segmentio/kafka-go"
	"github.com/shopspring/decimal"
	"go.elastic.co/apm"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type IService interface {
	process(db *gorm.DB, event fullEvent, ctx context.Context) *kafka.Message
}

type service struct {
	goTokenomicsWrapper go_tokenomics.IGoTokenomicsWrapper
}

func NewViewContentService(goTokenomicsWrapper go_tokenomics.IGoTokenomicsWrapper) IService {
	return &service{
		goTokenomicsWrapper: goTokenomicsWrapper,
	}
}

func (s *service) process(db *gorm.DB, event fullEvent, ctx context.Context) *kafka.Message {
	if err := s.handleOne(db, event, ctx); err != nil {
		apm_helper.LogError(err, ctx)
		return nil
	} else {
		return &event.Messages
	}
}

func (s *service) handleOne(db *gorm.DB, event fullEvent, ctx context.Context) error {
	apm_helper.AddApmLabel(apm.TransactionFromContext(ctx), "user_id", event.UserId)
	apm_helper.AddApmLabel(apm.TransactionFromContext(ctx), "content_id", event.ContentId)
	apm_helper.AddApmLabel(apm.TransactionFromContext(ctx), "content_author_id", event.ContentAuthorId)
	apm_helper.AddApmLabel(apm.TransactionFromContext(ctx), "source_view", event.SourceView)

	if event.UserId <= 0 || event.UserId == event.ContentAuthorId || !isSourceViewSupportedForAd(event.SourceView) {
		return nil
	}

	tx := db.Begin()
	defer tx.Rollback()

	var adCampaign database.AdCampaign
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("content_id = ? and status = ?", event.ContentId, database.AdCampaignStatusActive).
		Find(&adCampaign).Error; err != nil {
		return errors.WithStack(err)
	}

	if adCampaign.Id == 0 {
		return nil
	}

	if (adCampaign.EndedAt.Valid && time.Now().UTC().After(adCampaign.EndedAt.Time)) || (adCampaign.Views > 0 && !adCampaign.Paid) {
		adCampaign.Status = database.AdCampaignStatusCompleted
		if err := tx.Model(&adCampaign).Update("status", adCampaign.Status).Error; err != nil {
			return errors.WithStack(err)
		}

		if err := tx.Commit().Error; err != nil {
			return errors.WithStack(err)
		}

		return nil
	}

	var adCampaignView database.AdCampaignView
	if err := tx.Where("ad_campaign_id = ? and user_id = ?", adCampaign.Id, event.UserId).Find(&adCampaignView).Error; err != nil {
		return errors.WithStack(err)
	}

	if adCampaignView.AdCampaignId != 0 {
		return nil
	}

	adCampaignView.AdCampaignId = adCampaign.Id
	adCampaignView.UserId = event.UserId
	adCampaignView.CreatedAt = time.Now().UTC()

	if err := tx.Create(&adCampaignView).Error; err != nil {
		return errors.WithStack(err)
	}

	adCampaign.Views++

	if err := tx.Exec("update ad_campaigns set views = views + 1 where id = ?", adCampaign.Id).Error; err != nil {
		return errors.WithStack(err)
	}

	needPay := false

	if (adCampaign.Views-1)%1000 == 0 {
		adCampaign.Paid = false
		if err := tx.Model(&adCampaign).Update("paid", adCampaign.Paid).Error; err != nil {
			return errors.WithStack(err)
		}

		needPay = true
	}

	if err := tx.Commit().Error; err != nil {
		return errors.WithStack(err)
	}

	if needPay {
		tx2 := db.Begin()
		defer tx2.Rollback()

		adCampaign = database.AdCampaign{}
		if err := tx2.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("content_id = ? and status = ?", event.ContentId, database.AdCampaignStatusActive).
			Find(&adCampaign).Error; err != nil {
			return errors.WithStack(err)
		}

		if adCampaign.Id == 0 {
			return nil
		}

		adCampaign.Budget = adCampaign.Budget.Sub(adCampaign.Price)

		if adCampaign.Budget.LessThanOrEqual(decimal.Zero) {
			adCampaign.Status = database.AdCampaignStatusCompleted
			if err := tx2.Model(&adCampaign).
				Update("status", adCampaign.Status).
				Update("budget", adCampaign.Budget).Error; err != nil {
				return errors.WithStack(err)
			}

			if err := tx2.Commit().Error; err != nil {
				return errors.WithStack(err)
			}

			return nil
		}

		writeOffUserTokensForAdResp := <-s.goTokenomicsWrapper.WriteOffUserTokensForAd(adCampaign.UserId, adCampaign.Id, adCampaign.Price, ctx, false)
		if writeOffUserTokensForAdResp.Error != nil {
			return errors.WithStack(writeOffUserTokensForAdResp.Error.ToError())
		}

		adCampaign.Paid = true
		if err := tx2.Model(&adCampaign).
			Update("paid", adCampaign.Paid).
			Update("budget", adCampaign.Budget).Error; err != nil {
			return errors.WithStack(err)
		}

		if err := tx2.Commit().Error; err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}

func isSourceViewSupportedForAd(sourceView eventsourcing.SourceView) bool {
	switch sourceView {
	case eventsourcing.SourceViewFeedCountry:
		fallthrough
	case eventsourcing.SourceViewFeedTop:
		fallthrough
	case eventsourcing.SourceViewFeedFollowing:
		fallthrough
	case eventsourcing.SourceViewFeedInterests:
		return true
	default:
		return false
	}
}
