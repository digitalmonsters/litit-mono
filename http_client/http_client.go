package http_client

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/digitalmonsters/go-common/apm_helper"
	"github.com/imroc/req/v3"
	"github.com/rs/zerolog/log"
	"go.elastic.co/apm"
	"io/ioutil"
	"net/url"
	"time"
)

type HttpClient struct {
	cl                *req.Client
	targetServiceName string
}

type HttpRequest struct {
	*req.Request
}

type AsyncChan struct {
	Resp *req.Response
	Err  error
}

var DefaultHttpClient = NewHttpClient()

type forceLogKey struct {
}

func NewHttpClient() *HttpClient {
	client := req.C().SetTimeout(30 * time.Second)

	h := &HttpClient{
		cl: client,
	}

	extractServiceName := func(request *req.Request) string {
		if len(h.targetServiceName) > 0 {
			return h.targetServiceName
		}

		if request.URL != nil {
			return request.URL.Hostname()
		}

		u, err := url.Parse(request.RawURL)

		if err != nil {
			log.Err(err).Send()
		}

		if u != nil {
			return u.Hostname()
		}

		return "http_external"
	}

	client.OnBeforeRequest(func(client *req.Client, request *req.Request) error {
		if parentApm := apm.TransactionFromContext(request.Context()); parentApm != nil {
			span := parentApm.StartSpan("<to_be_replaced>", "<to_be_replaced>", nil)

			request.SetContext(apm.ContextWithSpan(request.Context(), span))
		}

		return nil
	})

	client.OnAfterResponse(func(client *req.Client, response *req.Response) error {
		ctx := response.Request.Context()

		if span := apm.SpanFromContext(ctx); span != nil && span.SpanData != nil {
			defer func() {
				if err := recover(); err != nil {
					log.Ctx(response.Request.Context()).Error().Msgf(fmt.Sprintf("%v", err))
				}

				if span.SpanData != nil && !span.Dropped() {
					span.End()
				}
			}()

			forceLog, _ := response.Request.Context().Value(forceLogKey{}).(bool)

			if response.IsError() {
				forceLog = true
			}

			targetServiceName := extractServiceName(response.Request)

			span.Name = fmt.Sprintf("HTTP [%v] [%v]", response.Request.RawRequest.Method,
				response.Request.RawURL)
			span.Type = targetServiceName

			apm_helper.AddSpanApmLabel(span, "full_url", response.Request.RawURL)

			finalStatement := ""

			if forceLog {
				var rawBodyRequest []byte

				rawBodyResponse := response.Bytes()

				if r, err := ioutil.ReadAll(response.Request.RawRequest.Body); err != nil {
					log.Ctx(ctx).Err(err).Send()
				} else {
					rawBodyRequest = r
				}

				if data, err := json.Marshal(map[string]interface{}{
					"request":  rawBodyRequest,
					"response": rawBodyResponse,
				}); err != nil {
					log.Ctx(ctx).Err(err).Send()

					finalStatement = fmt.Sprintf("request [%v] || response [%v]", rawBodyRequest, rawBodyResponse)
				} else {
					finalStatement = string(data)
				}
			}

			span.Context.SetHTTPRequest(response.Request.RawRequest)
			span.Context.SetHTTPStatusCode(response.StatusCode)
			span.Context.SetDatabase(apm.DatabaseSpanContext{
				Instance:  targetServiceName,
				Type:      targetServiceName,
				Statement: finalStatement,
			})
		}
		return nil
	})

	return h
}

func (h *HttpClient) WithServiceName(serviceName string) *HttpClient {
	h.targetServiceName = serviceName

	return h
}

func (h *HttpClient) WithTimeout(duration time.Duration) *HttpClient {
	h.cl.SetTimeout(duration)

	return h
}

func (h HttpClient) NewRequest(ctx context.Context) *HttpRequest {
	return &HttpRequest{h.cl.R().SetContext(ctx)}
}

func (h HttpClient) NewRequestWithTimeout(ctx context.Context, timeout time.Duration) *HttpRequest {
	return &HttpRequest{h.cl.Clone().SetTimeout(timeout).R().SetContext(ctx)}
}

func (r *HttpRequest) WithForceLog() *HttpRequest {
	r.SetContext(context.WithValue(r.Context(), forceLogKey{}, true))

	return r
}

func (r *HttpRequest) GetAsync(url string) chan AsyncChan {
	return r.doAsync(func() (*req.Response, error) {
		return r.Get(url)
	})
}

func (r *HttpRequest) PostAsync(url string) chan AsyncChan {
	return r.doAsync(func() (*req.Response, error) {
		return r.Post(url)
	})
}

func (r HttpRequest) doAsync(fn func() (*req.Response, error)) chan AsyncChan {
	ch := make(chan AsyncChan, 2)

	go func() {
		defer func() {
			close(ch)
		}()

		r, e := fn()

		ch <- AsyncChan{
			Resp: r,
			Err:  e,
		}
	}()

	return ch
}
