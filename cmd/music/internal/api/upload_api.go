package api

import (
	"encoding/json"
	"github.com/digitalmonsters/music/pkg/uploader"
	"github.com/digitalmonsters/music/utils"
	"github.com/valyala/fasthttp"
)

func (m *musicApp) InitUploadApi() {
	m.httpRouter.Router().OPTIONS("/upload", func(ctx *fasthttp.RequestCtx) {
		utils.SetCors(ctx)
	})

	m.httpRouter.Router().POST("/upload", func(ctx *fasthttp.RequestCtx) {
		defer func() {
			utils.SetCors(ctx)
		}()

		resp, err := uploader.FileUpload(m.cfg, uploader.UploadTypeAdminMusic, ctx)
		if err != nil {
			ctx.Error(err.Error(), fasthttp.StatusInternalServerError)
		} else {
			ctx.Response.Header.Set("Content-Type", "application/json")
			if respBytes, err := json.Marshal(resp); err != nil {
				ctx.Error(err.Error(), fasthttp.StatusInternalServerError)
			} else {
				ctx.Response.SetBody(respBytes)
			}
		}
	})
}
