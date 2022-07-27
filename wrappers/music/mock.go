package music

import (
	"context"
	"github.com/digitalmonsters/go-common/wrappers"
)

//goland:noinspection ALL
type MusicWrapperMock struct {
	GetMusicInternalFn func(ids []int64, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[map[int64]SimpleMusic]
}

func (w *MusicWrapperMock) GetMusicInternal(ids []int64, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[map[int64]SimpleMusic] {
	return w.GetMusicInternalFn(ids, ctx, forceLog)
}

func GetMock() ILikeWrapper { // for compiler errors
	return &MusicWrapperMock{}
}
