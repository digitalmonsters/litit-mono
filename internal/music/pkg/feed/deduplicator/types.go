package deduplicator

import "time"

const (
	defaultTTL = 30 * time.Minute
)

type SongExpiration struct {
	SongId    int64 `json:"id"`
	ExpiresAt int64 `json:"e"`
}
