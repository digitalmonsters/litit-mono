package api

import (
	"encoding/json"
	"github.com/digitalmonsters/go-common/error_codes"
	"github.com/digitalmonsters/go-common/router"
	"github.com/digitalmonsters/go-common/swagger"
	"github.com/digitalmonsters/go-common/wrappers/music"
	"github.com/digitalmonsters/music/pkg/database"
)

func (c *creatorApp) initServiceApi() error {
	serviceCommands := []router.ICommand{
		c.getMusicInternal(),
	}

	for _, command := range serviceCommands {
		if err := c.httpRouter.GetRpcServiceEndpoint().RegisterRpcCommand(command); err != nil {
			return err
		}
	}
	return nil
}

func (c *creatorApp) getMusicInternal() router.ICommand {
	method := "GetMusicInternal"

	c.apiDef[method] = swagger.ApiDescription{
		Request:           music.GetMusicInternalRequests{},
		Response:          map[int64]music.SimpleMusic{},
		MethodDescription: "get music internal",
		Tags:              []string{"music"},
	}

	return router.NewServiceCommand(method, func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		var req music.GetMusicInternalRequests

		if err := json.Unmarshal(request, &req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}

		if len(req.Ids) == 0 {
			return map[int64]music.SimpleMusic{}, nil
		}

		var songs []database.CreatorSong
		if err := database.GetDbWithContext(database.DbTypeReadonly, executionData.Context).Where("id in ?", req.Ids).Find(&songs).Error; err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		result := map[int64]music.SimpleMusic{}

		for _, s := range songs {
			result[s.Id] = music.SimpleMusic{
				Id:                s.Id,
				UserId:            s.UserId,
				Name:              s.Name,
				Status:            s.Status,
				LyricAuthor:       s.LyricAuthor,
				MusicAuthor:       s.MusicAuthor,
				CategoryId:        s.CategoryId,
				MoodId:            s.MoodId,
				FullSongUrl:       s.FullSongUrl,
				FullSongDuration:  s.FullSongDuration,
				ShortSongUrl:      s.ShortSongUrl,
				ShortSongDuration: s.ShortSongDuration,
				ImageUrl:          s.ImageUrl,
				Hashtags:          s.Hashtags,
				ShortListens:      s.ShortListens,
				FullListens:       s.FullListens,
				Likes:             s.Likes,
				Comments:          s.Comments,
				UsedInVideo:       s.UsedInVideo,
				CreatedAt:         s.CreatedAt,
			}
		}

		return result, nil
	}, false)
}
