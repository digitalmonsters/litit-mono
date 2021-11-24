package common

import (
	"github.com/digitalmonsters/go-common/wrappers/auth"
	"github.com/valyala/fasthttp"
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

func Test() {
	a := auth.NewAuthWrapper()

	v := <- a.ParseToken()
	v.
}