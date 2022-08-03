package ad_campaign

import (
	"context"
	"github.com/RichardKnop/machinery/v1/tasks"
	"github.com/digitalmonsters/ads-manager/pkg/database"
	"github.com/digitalmonsters/go-common/apm_helper"
	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

func (s *service) InitTasks() error {
	if err := s.jobber.RegisterTask("ad_campaigns:check_end",
		func(traceHeader string) error {
			apmTransaction := apm_helper.StartNewApmTransaction("ad_campaigns:check_end", "task", nil, nil)
			defer apmTransaction.End()
			ctx := boilerplate.CreateCustomContext(context.Background(), apmTransaction, log.Logger)

			if err := database.GetDbWithContext(database.DbTypeMaster, ctx).
				Exec("update ad_campaigns set status = ? where ended_at is not null and ended_at >= now() and status = ?",
					database.AdCampaignStatusCompleted, database.AdCampaignStatusActive).Error; err != nil {
				return errors.WithStack(err)
			}

			return nil
		},
	); err != nil {
		return errors.WithStack(err)
	}

	if err := s.jobber.RegisterPeriodicTask("0 * * * *", "ad_campaigns:check_end", &tasks.Signature{
		Name: "ad_campaigns:check_end",
	}); err != nil {
		return errors.WithStack(err)
	}

	return nil
}
