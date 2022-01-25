package music_source

import (
	"github.com/digitalmonsters/music/pkg/database"
	"github.com/digitalmonsters/music/pkg/music_source/internal"
	"gopkg.in/guregu/null.v4"
)

type ListMusicRequest struct {
	SearchKeyword null.String         `json:"search_keyword"`
	Source        database.SongSource `json:"source"`
	Page          int                 `json:"page"`
	Size          int                 `json:"size"`
}

type ListMusicResponse struct {
	Songs      []internal.SongModel `json:"songs"`
	TotalCount int64                `json:"total_count"`
}
