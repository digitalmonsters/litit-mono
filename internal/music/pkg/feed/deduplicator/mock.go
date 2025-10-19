package deduplicator

import (
	"context"
	"github.com/digitalmonsters/music/pkg/database"
)

type Mock struct {
}

func GetMock() IDeDuplicator {
	return &Mock{}
}

func (m *Mock) SetIdsToIgnore(content []*database.CreatorSong, userId int64, expirationData []SongExpiration, ctx context.Context) {
}

func (m *Mock) GetIdsToIgnore(userId int64, ctx context.Context) ([]SongExpiration, []int64) {
	return []SongExpiration{}, []int64{}
}
