package utils

import "github.com/valyala/fasthttp"

func AppendBrowserHeaders(request *fasthttp.Request) {
	request.Header.Set("Accept-Encoding", "gzip")
	request.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")
	request.Header.Set("Connection", "keep-alive")
	request.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.88 Safari/537.36")
}

func UnpackFastHttpBody(response *fasthttp.Response) ([]byte, error) {
	encoding := response.Header.Peek("Content-Encoding")

	if encoding != nil && string(encoding) == "gzip" {
		return response.BodyGunzip()
	} else {
		return response.Body(), nil
	}
}

func SetCors(ctx *fasthttp.RequestCtx) {
	ctx.Response.Header.Set("Access-Control-Allow-Credentials", "true")
	ctx.Response.Header.SetBytesV("Access-Control-Allow-Origin", ctx.Request.Header.Peek("Origin"))
	ctx.Response.Header.Set("Access-Control-Allow-Headers", "*")
	ctx.Response.Header.Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
}
