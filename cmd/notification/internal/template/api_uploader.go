package template

import (
	"encoding/json"
	"github.com/digitalmonsters/go-common/error_codes"
	"github.com/digitalmonsters/go-common/router"
	"github.com/digitalmonsters/go-common/swagger"
	"github.com/digitalmonsters/notification-handler/pkg/template/uploader"
	"github.com/digitalmonsters/notification-handler/pkg/utils"
	"github.com/valyala/fasthttp"
)

func (a templateApp) initUploaderApi(httpRouter *router.HttpRouter) error {
	path := "/upload_notification_image"

	a.apiDef[path] = swagger.ApiDescription{
		Response:          "image_url",
		MethodDescription: "Upload image. Multipart, key: File",
		Tags:              []string{"notification"},
	}

	httpRouter.Router().OPTIONS(path, func(ctx *fasthttp.RequestCtx) {
		utils.SetCors(ctx)
	})

	httpRouter.Router().POST(path, func(ctx *fasthttp.RequestCtx) {
		defer func() {
			utils.SetCors(ctx)
		}()

		ctx.Response.Header.Set("Content-Type", "application/json")

		var errWithCode *error_codes.ErrorWithCode
		resp, err := uploader.FileUpload(ctx)
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
