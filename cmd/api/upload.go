package api

import (
	"github.com/digitalmonsters/go-common/router"
	"github.com/digitalmonsters/music/configs"
	"github.com/digitalmonsters/music/pkg/own_storage/uploader"
	"github.com/valyala/fasthttp"
)

func InitUploadApi(httpRouter *router.HttpRouter, cfg *configs.Settings) {
	httpRouter.Router().POST("/upload", func(ctx *fasthttp.RequestCtx) {
		resp, err := uploader.FileUpload(cfg, ctx)
		if err != nil {
			ctx.Error(err.Error(), fasthttp.StatusInternalServerError)
		} else {
			ctx.Response.Header.Set("Access-Control-Allow-Credentials", "true")
			ctx.Response.Header.SetBytesV("Access-Control-Allow-Origin", ctx.Request.Header.Peek("Origin"))
			ctx.Response.Header.Set("Access-Control-Allow-Headers", "*")
			ctx.Response.Header.Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
			ctx.Response.Header.Set("Content-Type", "application/json")
			ctx.Response.SetBody(resp)
		}
	})
}
