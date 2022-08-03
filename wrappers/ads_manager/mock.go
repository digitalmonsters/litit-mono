package ads_manager

import (
	"context"
	"github.com/digitalmonsters/go-common/wrappers"
)

type AdsManagerWrapperMock struct {
	GetAdsContentForUserFn func(userId int64, contentIdsToMix []int64, contentIdsToIgnore []int64, ctx context.Context,
		forceLog bool) chan wrappers.GenericResponseChan[GetAdsContentForUserResponse]
}

func (w *AdsManagerWrapperMock) GetAdsContentForUser(userId int64, contentIdsToMix []int64, contentIdsToIgnore []int64,
	ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[GetAdsContentForUserResponse] {
	return w.GetAdsContentForUserFn(userId, contentIdsToMix, contentIdsToIgnore, ctx, forceLog)
}
