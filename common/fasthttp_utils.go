package common

import (
	"github.com/valyala/fasthttp"
	"strings"
)

func UnpackFastHttpBody(response *fasthttp.Response) ([]byte, error) {
	encoding := response.Header.Peek("Content-Encoding")

	if encoding != nil && string(encoding) == "gzip" {
		return response.BodyGunzip()
	} else {
		b := make([]byte, len(response.Body()))
		copy(b, response.Body())

		return b, nil
	}
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
