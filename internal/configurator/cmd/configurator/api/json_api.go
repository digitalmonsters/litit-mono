package api

import (
	"encoding/json"
	"github.com/digitalmonsters/configurator/pkg/configs"
	"github.com/digitalmonsters/configurator/pkg/database"
	"github.com/digitalmonsters/go-common/apm_helper"
	"github.com/digitalmonsters/go-common/application"
	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/digitalmonsters/go-common/callback"
	"github.com/rs/zerolog/log"
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

type migratorRequest struct {
	Configs map[string]configs.UpsertConfigRequest
}

func (a apiApp) jsonMigratorApi() (method, path string, handler fasthttp.RequestHandler) {
	return "POST", "/internal/json/migrator", func(ctx *fasthttp.RequestCtx) {
		apmTx := apm_helper.StartNewApmTransaction("publish config", "kafka", nil, nil)
		customCtx := boilerplate.CreateCustomContext(ctx, apmTx, log.Logger)
		var err error

		defer func() {
			if err != nil {
				apm_helper.LogError(err, customCtx)
				apmTx.End()
			} else {
				apmTx.Discard()
			}
		}()

		var cfgRequest application.MigratorRequest

		if err = json.Unmarshal(ctx.PostBody(), &cfgRequest); err != nil {
			ctx.Response.SetBodyString(err.Error())
			ctx.Response.SetStatusCode(500)

			return
		}
		tx := database.GetDb(database.DbTypeMaster).Begin()
		defer tx.Rollback()
		var newConfigs []application.ConfigModel
		var callbacks []callback.Callback
		newConfigs, callbacks, err = a.service.MigrateConfigs(tx, cfgRequest.Configs, a.publisher)

		if err != nil {
			ctx.Response.SetBodyString(err.Error())
			ctx.Response.SetStatusCode(500)

			return
		}
		if err = tx.Commit().Error; err != nil {
			ctx.Response.SetBodyString(err.Error())
			ctx.Response.SetStatusCode(500)

			return
		}

		for _, callbackFn := range callbacks {
			if err = callbackFn(ctx); err != nil {
				ctx.Response.SetBodyString(err.Error())
				ctx.Response.SetStatusCode(500)

				return
			}
		}

		res := map[string]application.ConfigModel{}

		for _, c := range newConfigs {
			res[c.Key] = c
		}
		var data []byte
		if data, err = json.Marshal(res); err != nil {
			ctx.Response.SetBodyString(err.Error())
			ctx.Response.SetStatusCode(500)

			return
		} else {
			ctx.Response.SetBodyRaw(data)
		}
	}
}
