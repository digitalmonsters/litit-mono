package template

import (
	"encoding/json"

	"github.com/digitalmonsters/go-common/apm_helper"
	"github.com/digitalmonsters/go-common/error_codes"
	"github.com/digitalmonsters/go-common/router"
	"github.com/digitalmonsters/notification-handler/pkg/template/uploader"
	"github.com/digitalmonsters/notification-handler/pkg/utils"
	"github.com/valyala/fasthttp"
)

func (a templateApp) initUploaderApi() error {
	path := "/upload_notification_image"

	a.httpRouter.Router().OPTIONS(path, func(ctx *fasthttp.RequestCtx) {
		utils.SetCors(ctx)
	})

	a.httpRouter.Router().POST(path, func(ctx *fasthttp.RequestCtx) {
		defer func() {
			utils.SetCors(ctx)
		}()

		ctx.Response.Header.Set("Content-Type", "application/json")

		var errWithCode *error_codes.ErrorWithCode
		apmTx := apm_helper.StartNewApmTransaction("notification_image_upload", "request", nil, nil)
		resp, err := uploader.FileUpload(ctx, a.uploaderWrapper, apmTx, a.appCtx)
		if err != nil {
			errWithCode = error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
			ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		} else {
			ctx.SetStatusCode(fasthttp.StatusOK)
		}

		if bytes, err := json.Marshal(router.ToRestResponse(resp, errWithCode)); err != nil {
			ctx.Error(err.Error(), fasthttp.StatusInternalServerError)
		} else {
			ctx.Response.SetBody(bytes)
		}
	})

	return nil
}
