package api

import (
	"github.com/digitalmonsters/go-common/router"
	"github.com/digitalmonsters/music/configs"
	"github.com/digitalmonsters/music/pkg/own_storage/uploader"
	"github.com/digitalmonsters/music/utils"
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
		resp, err := uploader.FileUpload(cfg, ctx)
		if err != nil {
			ctx.Error(err.Error(), fasthttp.StatusInternalServerError)
		} else {
			ctx.Response.Header.Set("Content-Type", "application/json")
			ctx.Response.SetBody(resp)
		}
	})
}
