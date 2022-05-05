package eventsourcing

import (
	"fmt"
	"github.com/digitalmonsters/go-common/common"
)

type ReferrerVerifiedEvent struct {
	UserId                  int64                 `json:"user_id"`
	ReferrerId              int64                 `json:"referrer_id"`
	ReferredByType          common.VerifiedByType `json:"referred_by_type"`
	IsVerifiedReferrer      bool                  `json:"is_verified_referrer"`
	GrandReferrerId         int64                 `json:"grand_referrer_id"`
	IsVerifiedGrandReferrer bool                  `json:"is_verified_grand_referrer"`
	BaseChangeEvent
}

func (c ReferrerVerifiedEvent) GetPublishKey() string {
	return fmt.Sprintf("%v_%v", c.UserId, c.ReferrerId)
}
