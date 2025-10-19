package music

import (
	"context"
	"fmt"
	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/digitalmonsters/go-common/common"
	"github.com/digitalmonsters/go-common/wrappers"
	"github.com/rs/zerolog/log"
	"go.elastic.co/apm"
	"time"
)

type IMusicWrapper interface {
	GetMusicInternal(ids []int64, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[map[int64]SimpleMusic]
}

//goland:noinspection GoNameStartsWithPackageName
type MusicWrapper struct {
	baseWrapper    *wrappers.BaseWrapper
	defaultTimeout time.Duration
	apiUrl         string
	serviceName    string
}

func NewMusicWrapper(config boilerplate.WrapperConfig) IMusicWrapper {
	timeout := 5 * time.Second

	if config.TimeoutSec > 0 {
		timeout = time.Duration(config.TimeoutSec) * time.Second
	}

	if len(config.ApiUrl) == 0 {
		config.ApiUrl = "http://music"

		log.Warn().Msgf("Api Url is missing for Likes. Setting as default : %v", config.ApiUrl)
	}

	return &MusicWrapper{
		baseWrapper:    wrappers.GetBaseWrapper(),
		defaultTimeout: timeout,
		apiUrl:         fmt.Sprintf("%v/rpc-service", common.StripSlashFromUrl(config.ApiUrl)),
		serviceName:    "music",
	}
}

func (w MusicWrapper) GetMusicInternal(ids []int64, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[map[int64]SimpleMusic] {
	return wrappers.ExecuteRpcRequestAsync[map[int64]SimpleMusic](w.baseWrapper, w.apiUrl,
		"GetMusicInternal", GetMusicInternalRequests{
			Ids: ids,
		}, map[string]string{}, w.defaultTimeout, apm.TransactionFromContext(ctx), w.serviceName, forceLog)
}
