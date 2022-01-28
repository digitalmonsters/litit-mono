package api

import (
	"github.com/digitalmonsters/go-common/router"
	"github.com/digitalmonsters/music/configs"
	"github.com/digitalmonsters/music/pkg/own_storage/uploader"
	"github.com/digitalmonsters/music/utils"
	"github.com/rs/zerolog/log"
	"github.com/valyala/fasthttp"
)

func InitUploadApi(httpRouter *router.HttpRouter, cfg *configs.Settings) {
	httpRouter.Router().OPTIONS("/upload", func(ctx *fasthttp.RequestCtx) {
		log.Info().Msg("[Upload Options] defer cors")
		utils.SetCors(ctx)
	})

	httpRouter.Router().POST("/upload", func(ctx *fasthttp.RequestCtx) {
		defer func() {
			log.Info().Msg("[Upload] defer cors")
			utils.SetCors(ctx)
		}()

		log.Info().Msg("[Upload] try to upload")
		resp, err := uploader.FileUpload(cfg, ctx)
		if err != nil {
			log.Info().Msgf("[Upload] an error occurred %s", err.Error())
			ctx.Error(err.Error(), fasthttp.StatusInternalServerError)
		} else {
			log.Info().Msg("[Upload] set content type")

			ctx.Response.Header.Set("Content-Type", "application/json")
			ctx.Response.SetBody(resp)

			log.Info().Msgf("[Upload] resp %s", string(resp))
			log.Info().Msg("[Upload] exite")
		}
	})
}
