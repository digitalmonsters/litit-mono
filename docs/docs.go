package docs

import (
	"encoding/json"
	"fmt"
	"github.com/digitalmonsters/go-common/router"
	"github.com/digitalmonsters/go-common/swagger"
	"github.com/valyala/fasthttp"
)

func RegisterHttpDoc(httpRouter *router.HttpRouter, docPath string, apiCmd []swagger.IApiCommand, apiDef map[string]swagger.ApiDescription,
	constants []swagger.ConstantDescription) {
	httpRouter.GET(docPath, func(ctx *fasthttp.RequestCtx) {
		res := swagger.GenerateDoc(apiCmd, apiDef, constants)

		ctx.Response.Header.SetContentType("text/html; charset=utf-8")

		b, _ := json.Marshal(res)

		redoc := fmt.Sprintf("<!DOCTYPE html>\n<html>\n  <head>\n    <title>Doc</title>\n    <meta charset=\"utf-8\"/>\n    <meta name=\"viewport\" content=\"width=device-width, initial-scale=1\">\n    <link href=\"https://fonts.googleapis.com/css?family=Montserrat:300,400,700|Roboto:300,400,700\" rel=\"stylesheet\">\n\n    <style>\n      body {\n        margin: 0;\n        padding: 0;\n      }\n    </style>\n  </head>\n  <body>\n    <div id=\"redoc-container\">\n    <redoc spec-url='http://petstore.swagger.io/v2/swagger.json'></redoc>\n    <script src=\"https://cdn.jsdelivr.net/npm/redoc@next/bundles/redoc.standalone.js\"> </script>\n    <script>Redoc.init(JSON.parse('%v'), {\n  scrollYOffset: 50\n}, document.getElementById('redoc-container'))</script>\n  </body>\n</html>",
			string(b))

		ctx.Response.SetBody([]byte(redoc))
	})
}
