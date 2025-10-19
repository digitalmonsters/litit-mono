package geolocation_service

import (
	"bytes"
	"encoding/json"
	"github.com/valyala/fasthttp"
)

type Service struct {
	url string
}

func NewService(url string) *Service {
	return &Service{
		url: url,
	}
}

func (s *Service) GetLocationInfo(ip string) (*LocationInfo, error) {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)
	req.SetRequestURI(s.url + ip)
	req.Header.Set("Accept-Encoding", "gzip")
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)
	err := fasthttp.Do(req, resp)
	if err != nil {
		return nil, err
	}

	contentEncoding := resp.Header.Peek("Content-Encoding")
	var body []byte
	if bytes.EqualFold(contentEncoding, []byte("gzip")) {
		body, _ = resp.BodyGunzip()
	} else {
		body = resp.Body()
	}

	var respInfo LocationInfo

	if err := json.Unmarshal(body, &respInfo); err != nil {
		return nil, err
	}

	return &respInfo, nil
}
