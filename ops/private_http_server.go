package ops

import (
	"fmt"
	fastRouter "github.com/fasthttp/router"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"
)

type PrivateHttpServer struct {
	srv        *fasthttp.Server
	realRouter *fastRouter.Router
	ready      bool
	healthy    bool
}

func NewPrivateHttpServer() *PrivateHttpServer {
	h := &PrivateHttpServer{
		realRouter: fastRouter.New(),
		healthy:    true,
	}

	h.registerHttpReadinessCheck()
	h.registerHttpHealthCheck()
	h.registerMetrics()

	return h
}

func (r *PrivateHttpServer) registerMetrics() {
	r.realRouter.GET("/metrics", fasthttpadaptor.NewFastHTTPHandler(promhttp.Handler()))
}

func (r *PrivateHttpServer) registerHttpReadinessCheck() {
	r.realRouter.GET("/readiness", func(ctx *fasthttp.RequestCtx) {
		if r.ready {
			ctx.Response.SetStatusCode(200)
		} else {
			ctx.Response.SetStatusCode(500)
		}
	})
}

func (r *PrivateHttpServer) registerHttpHealthCheck() {
	r.realRouter.GET("/health", func(ctx *fasthttp.RequestCtx) {
		if r.healthy {
			ctx.Response.SetStatusCode(200)
		} else {
			ctx.Response.SetStatusCode(500)
		}
	})
}

func (r *PrivateHttpServer) Ready() {
	r.ready = true
}

func (r *PrivateHttpServer) UnHealthy() {
	r.healthy = false
}

func (r *PrivateHttpServer) StartAsync(port int) *PrivateHttpServer {
	if r.srv != nil {
		return r
	}

	r.srv = &fasthttp.Server{
		Handler: fasthttp.CompressHandlerBrotliLevel(r.realRouter.Handler,
			fasthttp.CompressDefaultCompression, fasthttp.CompressDefaultCompression),
	}

	go func() {
		log.Info().Msgf("Private http Server started on port [%v]", port)

		if err := r.srv.ListenAndServe(fmt.Sprintf("0.0.0.0:%v", port)); err != nil {
			panic(err)
		}
	}()

	return r
}
