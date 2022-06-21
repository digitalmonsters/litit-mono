package common

import (
	"bytes"
	"fmt"
	"github.com/pkg/errors"
	"github.com/valyala/fasthttp"
	"strings"
)

type ContentEncodingType string

const (
	ContentEncodingGzip    ContentEncodingType = "gzip"
	ContentEncodingBrotli  ContentEncodingType = "br"
	ContentEncodingDeflate ContentEncodingType = "deflate"
)

func UnpackFastHttpBody(response *fasthttp.Response) ([]byte, error) {
	encoding := response.Header.Peek("Content-Encoding")

	if len(encoding) == 0 {
		b := make([]byte, len(response.Body()))
		copy(b, response.Body())

		return b, nil
	}

	var err error
	var buf bytes.Buffer
	encodingStr := ContentEncodingType(strings.ToLower(string(encoding)))

	switch encodingStr {
	case ContentEncodingGzip:
		_, err = fasthttp.WriteGunzip(&buf, response.Body())
	case ContentEncodingBrotli:
		_, err = fasthttp.WriteUnbrotli(&buf, response.Body())
	case ContentEncodingDeflate:
		_, err = fasthttp.WriteInflate(&buf, response.Body())
	default:
		err = errors.New(fmt.Sprintf("Cannot decompress response. Unknown Content-Encoding value: %s", encodingStr))
	}

	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func GetRealIp(req *fasthttp.RequestCtx) string {
	ip := ""

	if ipHeader := req.Request.Header.Peek("X-Envoy-External-Address"); len(ipHeader) > 0 {
		ip = string(ipHeader)
	}

	if len(ip) > 0 {
		return ip
	}

	if ipHeader := req.Request.Header.Peek("X-Forwarded-For"); len(ipHeader) > 0 {
		ip = strings.TrimSpace(strings.Split(string(ipHeader), ",")[0])
	}

	if len(ip) == 0 {
		ip = "0.0.0.0"
	}

	return ip
}
func StripSlashFromUrl(input string) string {
	if len(input) == 0 {
		return input
	}

	input = strings.TrimSpace(strings.ToLower(input))

	if strings.HasSuffix(input, "/") {
		return input[:len(input)-1]
	}

	return input
}
