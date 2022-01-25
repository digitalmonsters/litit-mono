package soundstripe

import (
	"encoding/json"
	"fmt"
	"github.com/digitalmonsters/go-common/apm_helper"
	"github.com/digitalmonsters/music/configs"
	"github.com/digitalmonsters/music/pkg/database"
	"github.com/digitalmonsters/music/pkg/music_source/internal"
	"github.com/digitalmonsters/music/utils"
	"github.com/gammazero/workerpool"
	"github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	"github.com/thoas/go-funk"
	"github.com/valyala/fasthttp"
	"go.elastic.co/apm"
	"gorm.io/gorm"
	"time"
)

type Service struct {
	apiUrl     string
	apiToken   string
	timeout    time.Duration
	workerPool *workerpool.WorkerPool
	songsCache *cache.Cache
}

func NewService(cfg configs.SoundStripeConfig) internal.IMusicStorageAdapter {
	return &Service{
		apiUrl:     cfg.ApiUrl,
		apiToken:   cfg.ApiToken,
		timeout:    time.Second * time.Duration(cfg.MaxTimeout),
		workerPool: workerpool.New(cfg.MaxWorkers),
		songsCache: cache.New(60*time.Minute, 61*time.Minute),
	}
}

func (s *Service) SyncSongsList(externalSongsIds []string, tx *gorm.DB, apmTransaction *apm.Transaction) error {
	var missing []string

	var songsInDb []string
	if err := tx.Model(database.Song{}).Where("external_id in ? and source = ?", externalSongsIds, database.SongSourceSoundStripe).
		Pluck("external_id", &songsInDb).Error; err != nil {
		return errors.WithStack(err)
	}

	for _, songId := range externalSongsIds {
		if !funk.ContainsString(songsInDb, songId) {
			missing = append(missing, songId)
		}
	}

	if len(missing) == 0 {
		return nil
	}

	var songsToAdd []database.Song
	chans := make([]chan internal.GetSongResponseChan, 0)

	for _, songId := range missing {
		cachedItem, hasCachedItem := s.songsCache.Get(songId)
		if hasCachedItem {
			song := cachedItem.(internal.SongModel)
			songsToAdd = append(songsToAdd, database.Song{
				Source:     database.SongSourceSoundStripe,
				ExternalId: songId,
				Title:      song.Title,
				Artist:     song.Artist,
				ImageUrl:   song.ImageUrl,
				Genre:      song.Genre,
				Duration:   song.Duration,
			})
			continue
		}

		chans = append(chans, s.getSong(songId, apmTransaction))
	}

	var internalErrors []error
	if len(chans) > 0 {
		for _, ch := range chans {
			result := <-ch
			if result.Error != nil {
				internalErrors = append(internalErrors, result.Error)
				continue
			}

			songsToAdd = append(songsToAdd, database.Song{
				Source:     database.SongSourceSoundStripe,
				ExternalId: result.Song.ExternalId,
				Title:      result.Song.Title,
				Artist:     result.Song.Artist,
				ImageUrl:   result.Song.ImageUrl,
				Genre:      result.Song.Genre,
				Duration:   result.Song.Duration,
			})

			s.songsCache.Set(result.Song.ExternalId, result.Song, cache.DefaultExpiration)
		}
	}

	if len(songsToAdd) > 0 {
		if err := tx.Create(&songsToAdd).Error; err != nil {
			return errors.WithStack(err)
		}
	}

	if len(internalErrors) > 0 {
		for _, err := range internalErrors {
			apm_helper.CaptureApmError(err, apmTransaction)
		}

		return errors.Wrap(internalErrors[0], "sync error")
	}

	return nil
}

func (s *Service) GetSongsList(req internal.GetSongsListRequest, apmTransaction *apm.Transaction) chan internal.GetSongsListResponseChan {
	resChan := make(chan internal.GetSongsListResponseChan, 2)
	s.workerPool.Submit(func() {
		/*finalResponse := internal.GetSongsListResponseChan{}

		queryParams := fmt.Sprintf("?size=%v&page=%v", req.Size, req.Page)
		if req.SearchKeyword.Valid {
			queryParams = fmt.Sprintf("%v&q=?", req.SearchKeyword.Valid)
		}

		url := fmt.Sprintf("songs%v", queryParams)

		internalResp, err := s.makeApiRequestInternal(url, "GET", nil, apmTransaction)
		if err != nil {
			finalResponse.Error = err
			resChan <- finalResponse
			return
		}

		if finalResponse.Error == nil && len(internalResp) > 0 {
			var ssResp soundstripeSongsResp
			if err = json.Unmarshal(internalResp, &ssResp); err != nil {
				finalResponse.Error = err
				resChan <- finalResponse
				return
			}

			//todo: sounstripe models parsing

			var songs []internal.SongModel
			for _, song := range songs {
				s.songsCache.Set(song.ExternalId, song, cache.DefaultExpiration)
			}

			finalResponse.Response = internal.GetSongsListResponse{
				Songs:      songs,
				TotalCount: ssResp.Links.Meta.TotalCount,
			}
		}*/

		var songs []internal.SongModel
		for i := 1; i <= 10; i++ {
			song := internal.SongModel{
				ExternalId: fmt.Sprint(i),
				Title:      fmt.Sprintf("test_title%v", i),
				Artist:     fmt.Sprintf("test_artist%v", i),
				ImageUrl:   fmt.Sprintf("test_image_url%v", i),
				Genre:      fmt.Sprintf("test_genre%v", i),
				Duration:   float64(10 * i),
			}

			songs = append(songs, song)
			s.songsCache.Set(song.ExternalId, song, cache.DefaultExpiration)
		}

		finalResponse := internal.GetSongsListResponseChan{
			Error: nil,
			Response: internal.GetSongsListResponse{
				Songs:      songs,
				TotalCount: 10,
			},
		}

		resChan <- finalResponse
	})

	return resChan
}

func (s *Service) getSong(externalId string, apmTransaction *apm.Transaction) chan internal.GetSongResponseChan {
	resChan := make(chan internal.GetSongResponseChan, 2)
	s.workerPool.Submit(func() {
		finalResponse := internal.GetSongResponseChan{}
		url := fmt.Sprintf("songs/%v", externalId)
		internalResp, err := s.makeApiRequestInternal(url, "GET", nil, apmTransaction)
		if err != nil {
			finalResponse.Error = err
			resChan <- finalResponse
			return
		}

		if finalResponse.Error == nil && len(internalResp) > 0 {
			var song internal.SongModel
			if err = json.Unmarshal(internalResp, &song); err != nil {
				finalResponse.Error = err
				resChan <- finalResponse
				return
			}

			finalResponse.Song = song
		}

		resChan <- finalResponse
	})

	return resChan
}

func (s *Service) makeApiRequestInternal(apiMethod string, httpMethod string, body []byte, apmTransaction *apm.Transaction) ([]byte, error) {
	url := fmt.Sprintf("%v/%v", s.apiUrl, apiMethod)
	cl := &fasthttp.Client{}

	httpReq := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(httpReq)
	httpRes := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(httpRes)

	httpReq.SetRequestURI(url)
	httpReq.Header.SetMethod(httpMethod)
	if body != nil {
		httpReq.SetBody(body)
	}

	err := apm_helper.SendHttpRequest(cl, httpReq, httpRes, apmTransaction, s.timeout, true)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	if httpRes.StatusCode() != fasthttp.StatusOK {
		return nil, errors.New(fmt.Sprintf("Soundstipe HTTP CODE %v. URL %v", httpRes.StatusCode(), url))
	}

	return utils.UnpackFastHttpBody(httpRes)
}
