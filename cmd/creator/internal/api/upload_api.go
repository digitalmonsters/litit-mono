package api

import (
	"github.com/digitalmonsters/music/pkg/uploader"
	"github.com/digitalmonsters/music/utils"
	"github.com/valyala/fasthttp"
	"strconv"
)

func (c *creatorApp) initUploadApi() {
	c.httpRouter.Router().OPTIONS("/creator/upload", func(ctx *fasthttp.RequestCtx) {
		utils.SetCors(ctx)
	})

	c.httpRouter.Router().POST("/creator/upload", func(ctx *fasthttp.RequestCtx) {
		defer func() {
			utils.SetCors(ctx)
		}()

		m, err := ctx.Request.MultipartForm()
		if err != nil {
			ctx.Error(err.Error(), fasthttp.StatusInternalServerError)
			return
		}

		t := m.Value["type"]
		if len(t) == 0 {
			ctx.Error("type is required", fasthttp.StatusInternalServerError)
			return
		}

		ut, err := strconv.Atoi(t[0])
		if err != nil {
			ctx.Error(err.Error(), fasthttp.StatusInternalServerError)
			return
		}

		uploadType := uploader.UploadType(ut)
		if uploadType < uploader.UploadTypeAdminMusic || uploadType > uploader.UploadTypeCreatorsSongImage {
			ctx.Error("wrong upload type", fasthttp.StatusInternalServerError)
			return
		}

		resp, err := uploader.FileUpload(c.cfg, uploadType, ctx)
		if err != nil {
			ctx.Error(err.Error(), fasthttp.StatusInternalServerError)
		} else {
			ctx.Response.Header.Set("Content-Type", "application/json")
			ctx.Response.SetBody(resp)
		}
	})
}
