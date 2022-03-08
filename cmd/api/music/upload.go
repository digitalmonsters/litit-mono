package music

import (
	"github.com/digitalmonsters/go-common/router"
	"github.com/digitalmonsters/music/configs"
	"github.com/digitalmonsters/music/pkg/uploader"
	"github.com/digitalmonsters/music/utils"
	"github.com/rs/zerolog/log"
	"github.com/valyala/fasthttp"
)

func InitUploadApi(httpRouter *router.HttpRouter, cfg *configs.Settings) {
	httpRouter.Router().OPTIONS("/upload", func(ctx *fasthttp.RequestCtx) {
		utils.SetCors(ctx)
	})

	httpRouter.Router().POST("/upload", func(ctx *fasthttp.RequestCtx) {
		defer func() {
			utils.SetCors(ctx)
		}()

		resp, err := uploader.FileUpload(cfg, uploader.UploadTypeMusic, ctx)
		if err != nil {
			log.Info().Msgf("[Upload] an error occurred %s", err.Error())
			ctx.Error(err.Error(), fasthttp.StatusInternalServerError)
		} else {
			ctx.Response.Header.Set("Content-Type", "application/json")
			ctx.Response.SetBody(resp)
		}
	})
}
