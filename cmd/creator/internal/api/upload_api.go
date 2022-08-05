package api

import (
	"encoding/json"
	"github.com/digitalmonsters/go-common/error_codes"
	"github.com/digitalmonsters/go-common/router"
	"github.com/digitalmonsters/music/pkg/uploader"
	"github.com/digitalmonsters/music/utils"
	"github.com/pkg/errors"
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

		var errWithCode *error_codes.ErrorWithCode

		m, err := ctx.Request.MultipartForm()
		if err != nil {
			errWithCode = error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
			setResponseBody(ctx, nil, errWithCode)
			return
		}

		t := m.Value["type"]
		if len(t) == 0 {
			errWithCode = error_codes.NewErrorWithCodeRef(errors.New("type is required"), error_codes.GenericServerError)
			setResponseBody(ctx, nil, errWithCode)
			return
		}

		ut, err := strconv.Atoi(t[0])
		if err != nil {
			errWithCode = error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
			setResponseBody(ctx, nil, errWithCode)
			return
		}

		uploadType := uploader.UploadType(ut)
		if uploadType < uploader.UploadTypeAdminMusic || uploadType > uploader.UploadTypeCreatorsSongImage {
			errWithCode = error_codes.NewErrorWithCodeRef(errors.New("wrong upload type"), error_codes.GenericServerError)
			setResponseBody(ctx, nil, errWithCode)
			return
		}

		resp, err := uploader.FileUpload(c.cfg, c.appConfig, uploadType, ctx)
		if err != nil {
			errWithCode = error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		setResponseBody(ctx, resp, errWithCode)
	})
}

func setResponseBody(ctx *fasthttp.RequestCtx, data interface{}, err *error_codes.ErrorWithCode) {
	ctx.Response.Header.Set("Content-Type", "application/json")
	if bytes, err := json.Marshal(router.ToRestResponse(data, err)); err != nil {
		ctx.Error(err.Error(), fasthttp.StatusInternalServerError)
	} else {
		ctx.Response.SetBody(bytes)
	}
}
