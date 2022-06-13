package utils

import "github.com/valyala/fasthttp"

func SetCors(ctx *fasthttp.RequestCtx) {
	ctx.Response.Header.Set("Access-Control-Allow-Credentials", "true")
	ctx.Response.Header.SetBytesV("Access-Control-Allow-Origin", ctx.Request.Header.Peek("Origin"))
	ctx.Response.Header.Set("Access-Control-Allow-Headers", "*")
	ctx.Response.Header.Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
}
