package api

import (
	"encoding/json"
	"github.com/digitalmonsters/configurator/pkg/database"
	"github.com/valyala/fasthttp"
)

type configRequest struct {
	Items []string `json:"items"`
}

func (a apiApp) jsonApi() (method, path string, handler fasthttp.RequestHandler) {
	return "POST", "/internal/json", func(ctx *fasthttp.RequestCtx) {
		var cfgRequest configRequest

		if err := json.Unmarshal(ctx.PostBody(), &cfgRequest); err != nil {
			ctx.Response.SetBodyString(err.Error())
			ctx.Response.SetStatusCode(500)

			return
		}

		configs, err := a.service.GetConfigsByIds(database.GetDb(database.DbTypeReadonly),
			cfgRequest.Items)

		if err != nil {
			ctx.Response.SetBodyString(err.Error())
			ctx.Response.SetStatusCode(500)

			return
		}

		res := map[string]string{}

		for _, c := range configs {
			res[c.Key] = c.Value
		}

		if data, err := json.Marshal(res); err != nil {
			ctx.Response.SetBodyString(err.Error())
			ctx.Response.SetStatusCode(500)

			return
		} else {
			ctx.Response.SetBodyRaw(data)
		}
	}
}
