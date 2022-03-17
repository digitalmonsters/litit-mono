package router

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/digitalmonsters/go-common/apm_helper"
	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/digitalmonsters/go-common/common"
	"github.com/digitalmonsters/go-common/error_codes"
	"github.com/digitalmonsters/go-common/rpc"
	"github.com/digitalmonsters/go-common/swagger"
	auth_go "github.com/digitalmonsters/go-common/wrappers/auth_go"
	fastRouter "github.com/fasthttp/router"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"
	"go.elastic.co/apm"
	"go.elastic.co/apm/module/apmhttp"
	"net/http/pprof"
	"os"
	"strings"
	"sync"
	"time"
)

type HttpRouter struct {
	realRouter               *fastRouter.Router
	hostname                 string
	restCommands             map[string]*RestCommand
	isProd                   bool
	authGoWrapper            auth_go.IAuthGoWrapper
	srv                      *fasthttp.Server
	rpcEndpointPublic        IRpcEndpoint
	rpcEndpointAdmin         IRpcEndpoint
	rpcEndpointAdminLegacy   IRpcEndpoint
	rpcEndpointService       IRpcEndpoint
	endpointRegistratorMutex sync.Mutex
}

var hostName string

// user auth -> node js -> creates tokens using forward-auth (rpc public, rest)
// user admin (same user) -> node js -> admin or super admin
// admin (rbac) -> auth go ->creates tokens using forward-auth (additional api)
// service -> no auth

// /rpc (user token or without token, if require identity validation = false) (auth method 1, 1.5)
// rest api -> /sddfsdf_fdsfsd/dsfds (auth method 1, 1.5)
// /rpc-admin -> (rbac) (3)
// /rpc-admin-legacy -> legacy admin command (command should use auth method 1.5)
// /rpc-service - internal services -> (will not be available for external use)

func NewRouter(rpcEndpointPath string, auth auth_go.IAuthGoWrapper) *HttpRouter {
	h := &HttpRouter{
		realRouter:               fastRouter.New(),
		endpointRegistratorMutex: sync.Mutex{},
		authGoWrapper:            auth,
		restCommands:             map[string]*RestCommand{},
	}

	if hostname, _ := os.Hostname(); len(hostname) > 0 {
		h.hostname = hostname
		hostName = hostname
	}

	if boilerplate.GetCurrentEnvironment() == boilerplate.Prod {
		h.isProd = true
	}

	return h
}

func (r *HttpRouter) GetRpcAdminLegacyEndpoint() IRpcEndpoint {
	if r.rpcEndpointAdminLegacy == nil {
		r.rpcEndpointAdminLegacy = newRpcEndpointPublic()

		r.prepareRpcEndpoint("/rpc-admin-legacy", r.rpcEndpointAdminLegacy, "rpc-admin-legacy")
	}

	return r.rpcEndpointAdminLegacy
}

func (r *HttpRouter) GetRpcPublicEndpoint() IRpcEndpoint {
	if r.rpcEndpointPublic == nil {
		r.rpcEndpointPublic = newRpcEndpointPublic()

		r.prepareRpcEndpoint("/rpc", r.rpcEndpointPublic, "rpc")
	}

	return r.rpcEndpointPublic
}

func (r *HttpRouter) GetRpcAdminEndpoint() IRpcEndpoint {
	if r.rpcEndpointAdmin == nil {
		r.rpcEndpointAdmin = newRpcEndpointAdmin()

		r.prepareRpcEndpoint("/rpc-admin", r.rpcEndpointAdmin, "rpc-admin")
	}

	return r.rpcEndpointAdmin
}

func (r *HttpRouter) GetRpcServiceEndpoint() IRpcEndpoint {
	if r.rpcEndpointService == nil {
		r.rpcEndpointService = newRpcEndpointService()

		r.prepareRpcEndpoint("/rpc-service", r.rpcEndpointService, "rpc-service")
	}

	return r.rpcEndpointService
}

func (r *HttpRouter) setCors(ctx *fasthttp.RequestCtx) {
	ctx.Response.Header.Set("Access-Control-Allow-Credentials", "true")
	ctx.Response.Header.SetBytesV("Access-Control-Allow-Origin", ctx.Request.Header.Peek("Origin"))
	ctx.Response.Header.Set("Access-Control-Allow-Headers", "*")
	ctx.Response.Header.Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
}

func (r *HttpRouter) RegisterProfiler() {
	r.endpointRegistratorMutex.Lock()
	defer r.endpointRegistratorMutex.Unlock()

	r.realRouter.GET("/debug/pprof/cpu", fasthttpadaptor.NewFastHTTPHandlerFunc(pprof.Profile))
	r.realRouter.GET("/debug/pprof/{name}", fasthttpadaptor.NewFastHTTPHandlerFunc(pprof.Index))
}

func (r *HttpRouter) RegisterDocs(apiDef map[string]swagger.ApiDescription,
	constants []swagger.ConstantDescription) {
	routes := map[string][]swagger.IApiCommand{}

	for _, c := range r.GetRestRegisteredCommands() {
		routes["/swagger"] = append(routes["/swagger"], c)
	}

	if r.rpcEndpointPublic != nil {
		for _, c := range r.rpcEndpointPublic.GetRegisteredCommands() {
			routes["/swagger"] = append(routes["/swagger"], c)
		}
	}

	if r.rpcEndpointAdmin != nil {
		for _, c := range r.rpcEndpointAdmin.GetRegisteredCommands() {
			routes["/swagger-admin"] = append(routes["/swagger-admin"], c)
		}
	}

	if r.rpcEndpointService != nil {
		for _, c := range r.rpcEndpointService.GetRegisteredCommands() {
			routes["/swagger-service"] = append(routes["/swagger-service"], c)
		}
	}

	if r.rpcEndpointAdminLegacy != nil {
		for _, c := range r.rpcEndpointAdminLegacy.GetRegisteredCommands() {
			routes["/swagger-admin-legacy"] = append(routes["/swagger-admin-legacy"], c)
		}
	}

	r.endpointRegistratorMutex.Lock()
	defer r.endpointRegistratorMutex.Unlock()

	r.realRouter.GET("/swag", func(ctx *fasthttp.RequestCtx) {
		ctx.Response.Header.SetContentType("text/html; charset=utf-8")

		ctx.Response.SetBodyRaw([]byte("<ul>\n<li><a href=\"/swagger\">public</a></li>\n<li><a href=\"/swagger-admin\">admin</a></li>\n<li><a href=\"/swagger-admin-legacy\">admin-legacy</a></li>\n<li><a href=\"/swagger-service\">service</a></li>\n</ul>"))
	})

	for path, commands := range routes {
		cPath := path
		cCommands := commands

		r.realRouter.GET(cPath, func(ctx *fasthttp.RequestCtx) {
			res := swagger.GenerateDoc(cCommands, apiDef, constants)

			ctx.Response.Header.SetContentType("text/html; charset=utf-8")

			b, _ := json.Marshal(res)

			redoc := fmt.Sprintf("<!DOCTYPE html>\n<html>\n  <head>\n    <title>Doc</title>\n    <meta charset=\"utf-8\"/>\n    <meta name=\"viewport\" content=\"width=device-width, initial-scale=1\">\n    <link href=\"https://fonts.googleapis.com/css?family=Montserrat:300,400,700|Roboto:300,400,700\" rel=\"stylesheet\">\n\n    <style>\n      body {\n        margin: 0;\n        padding: 0;\n      }\n    </style>\n  </head>\n  <body>\n    <div id=\"redoc-container\">\n    <redoc spec-url='http://petstore.swagger.io/v2/swagger.json'></redoc>\n    <script src=\"https://cdn.jsdelivr.net/npm/redoc@next/bundles/redoc.standalone.js\"> </script>\n    <script>Redoc.init(JSON.parse('%v'), {\n  scrollYOffset: 50\n}, document.getElementById('redoc-container'))</script>\n  </body>\n</html>",
				string(b))

			ctx.Response.SetBody([]byte(redoc))
		})
	}
}

func (r *HttpRouter) RegisterRestCmd(targetCmd *RestCommand) error {
	key := fmt.Sprintf("%v_%v", targetCmd.method, targetCmd.path)

	if _, ok := r.restCommands[key]; ok {
		return errors.New(fmt.Sprintf("rest command [%v] already registered", key))
	}

	r.restCommands[key] = targetCmd

	go func() {
		defer func() {
			r.endpointRegistratorMutex.Unlock()

			_ = recover()
		}()

		r.endpointRegistratorMutex.Lock()

		r.realRouter.OPTIONS(targetCmd.path, func(ctx *fasthttp.RequestCtx) {
			r.setCors(ctx)
		})
	}()

	r.endpointRegistratorMutex.Lock()
	defer r.endpointRegistratorMutex.Unlock()

	r.realRouter.Handle(targetCmd.method, targetCmd.path, func(ctx *fasthttp.RequestCtx) {
		var apmTransaction *apm.Transaction

		if traceHeader := ctx.Request.Header.Peek(apmhttp.W3CTraceparentHeader); len(traceHeader) > 0 {
			traceContext, _ := apmhttp.ParseTraceparentHeader(string(traceHeader))
			apmTransaction = apm_helper.StartNewApmTransactionWithTraceData(fmt.Sprintf("[%v] [%v]", targetCmd.method,
				targetCmd.path), "rest", nil, traceContext)
		} else {
			apmTransaction = apm_helper.StartNewApmTransaction(fmt.Sprintf("[%v] [%v]", targetCmd.method,
				targetCmd.path), "rest", nil, nil)
		}

		defer apmTransaction.End()

		requestBody := ctx.PostBody()

		rpcRequest := rpc.RpcRequest{
			Method:  targetCmd.path,
			Params:  requestBody,
			Id:      "1",
			JsonRpc: "2.0",
		}

		defer func() {
			r.setCors(ctx)
		}()

		apm_helper.AddApmLabel(apmTransaction, "full_url", string(ctx.URI().FullURI()))

		rpcResponse, shouldLog := r.executeAction(rpcRequest, targetCmd, ctx, apmTransaction, targetCmd.forceLog,
			func(key string) interface{} {
				if v := ctx.UserValue(key); v != nil {
					return v
				}

				if ctx.QueryArgs() != nil {
					if v := ctx.QueryArgs().Peek(key); len(v) > 0 {
						return string(v)
					}
				}

				if v := ctx.Request.Header.Peek(key); len(v) > 0 {
					return string(v)
				}

				return nil
			})

		var responseBody []byte

		defer func() {
			if !shouldLog {
				return
			}

			r.logRequestBody(requestBody, apmTransaction)
			r.logResponseBody(responseBody, apmTransaction)
			r.logRpcResponseError(rpcResponse, apmTransaction)
		}()

		finalStatusCode := int(error_codes.None)

		var err error

		var restResponse genericRestResponse

		restResponse.Success = true
		restResponse.ExecutionTimingMs = rpcResponse.ExecutionTimingMs
		restResponse.Hostname = rpcResponse.Hostname
		restResponse.Code = -1

		if rpcResponse.Result != nil {
			restResponse.Data = rpcResponse.Result
		}
		if rpcResponse.Error != nil {
			originalCode := int(rpcResponse.Error.Code)
			restResponse.Success = false
			restResponse.Error = rpcResponse.Error.Message
			restResponse.Stack = rpcResponse.Error.Stack

			if strings.EqualFold(restResponse.Error, "max threshold without kyc exceeded") {
				restResponse.Code = 2 // todo find a better way
			}
			if strings.EqualFold(restResponse.Error, error_codes.TokenomicsNotEnoughBalanceError.Error()) {
				restResponse.Code = int(error_codes.TokenomicsNotEnoughBalance)
			}
			if originalCode > 0 {
				finalStatusCode = originalCode
			} else {
				switch rpcResponse.Error.Code {
				case error_codes.GenericMappingError:
					finalStatusCode = int(error_codes.GenericValidationError)
				case error_codes.CommandNotFoundError:
					finalStatusCode = int(error_codes.GenericNotFoundError)
				default:
					finalStatusCode = int(error_codes.GenericServerError)
				}
			}
		}

		if d, ok := restResponse.Data.(*rpc.CustomFile); ok {
			responseBody = d.Data
			ctx.Response.Header.SetContentType(d.MimeType)
			contentDispositionFirstParam := d.ContentDispositionFirstParam

			if len(contentDispositionFirstParam) == 0 {
				contentDispositionFirstParam = "attachment"
			}

			ctx.Response.Header.Set(fasthttp.HeaderContentDisposition,
				fmt.Sprintf("%v; filename=\"%v\"", contentDispositionFirstParam, d.Filename))
			ctx.Response.Header.Set(fasthttp.HeaderAcceptRanges, "bytes")
		} else {
			if responseBody, err = json.Marshal(restResponse); err != nil {
				log.Err(err).Send()
			}
		}

		ctx.Response.SetBodyRaw(responseBody)
		ctx.Response.SetStatusCode(finalStatusCode)
	})

	return nil
}

func (r *HttpRouter) executeAction(rpcRequest rpc.RpcRequest, cmd ICommand, ctx *fasthttp.RequestCtx,
	apmTransaction *apm.Transaction, forceLog bool, getUserValue func(key string) interface{}) (rpcResponse rpc.RpcResponse, shouldLog bool) {
	totalTiming := time.Now()

	newCtx, cancel := context.WithCancel(ctx)
	newCtx = apm.ContextWithTransaction(newCtx, apmTransaction)

	defer cancel()

	r.logRequestHeaders(ctx, apmTransaction) // in future filter for specific routes
	r.logUserValues(ctx, apmTransaction)

	var panicErr error

	var executionMs int64

	rpcResponse = rpc.RpcResponse{
		JsonRpc: "2.0",
	}

	defer func() {
		ctx.Response.Header.SetContentType("application/json")
		rpcResponse.ExecutionTimingMs = executionMs
		rpcResponse.TotalTimingMs = time.Since(totalTiming).Milliseconds()
		rpcResponse.Hostname = r.hostname
	}()

	defer func() {
		if rec := recover(); rec != nil {
			shouldLog = true

			switch val := rec.(type) {
			case error:
				panicErr = errors.Wrap(val, fmt.Sprintf("panic! %v", val))
			default:
				panicErr = errors.New(fmt.Sprintf("panic! : %v", val))
			}

			if panicErr == nil {
				panicErr = errors.New("panic! and that is really bad")
			}

			rpcResponse.Result = nil
			rpcResponse.Error = &rpc.RpcError{
				Code:     error_codes.GenericPanicError,
				Message:  panicErr.Error(),
				Data:     nil,
				Hostname: r.hostname,
			}

			if !r.isProd {
				rpcResponse.Error.Stack = fmt.Sprintf("%+v", panicErr)
			}
		}
	}()

	rpcResponse.Id = rpcRequest.Id

	shouldLog = forceLog

	userId, isGuest, rpcError := cmd.CanExecute(ctx, apmTransaction, r.authGoWrapper)

	if rpcError != nil {
		rpcResponse.Error = rpcError

		return
	}

	if userId <= 0 && (cmd.RequireIdentityValidation() || cmd.AccessLevel() > common.AccessLevelPublic) {
		err := errors.New("missing jwt token for auth")

		rpcError = &rpc.RpcError{
			Code:     error_codes.MissingJwtToken,
			Message:  "missing jwt token for auth",
			Hostname: r.hostname,
		}

		if !r.isProd {
			rpcError.Stack = fmt.Sprintf("%+v", err)
		}

		rpcResponse.Error = rpcError

		return
	}

	if userId > 0 {
		if apmTransaction != nil {
			apmTransaction.Context.SetUserID(fmt.Sprint(userId))
		}
	}

	executionTiming := time.Now()

	if resp, err := cmd.GetFn()(rpcRequest.Params, MethodExecutionData{
		ApmTransaction: apmTransaction,
		Context:        newCtx,
		UserId:         userId,
		IsGuest:        isGuest,
		UserIp:         common.GetRealIp(ctx),
		getUserValueFn: getUserValue,
	}); err != nil {
		rpcResponse.Error = &rpc.RpcError{
			Code:     err.GetCode(),
			Message:  err.GetMessage(),
			Data:     nil,
			Hostname: r.hostname,
		}

		shouldLog = true

		if !r.isProd {
			rpcResponse.Error.Stack = err.GetStack()
		}
	} else {
		if resp == nil {
			resp = "ok"
		}

		rpcResponse.Result = resp
	}

	executionMs = time.Since(executionTiming).Milliseconds()
	return
}

func (r *HttpRouter) prepareRpcEndpoint(rpcEndpointPath string, endpoint IRpcEndpoint, apmTxType string) {
	r.endpointRegistratorMutex.Lock()
	defer r.endpointRegistratorMutex.Unlock()

	r.realRouter.OPTIONS(rpcEndpointPath, func(ctx *fasthttp.RequestCtx) {
		r.setCors(ctx)
	})

	r.realRouter.POST(rpcEndpointPath, func(ctx *fasthttp.RequestCtx) {
		var rpcRequest rpc.RpcRequest
		var rpcResponse rpc.RpcResponse
		var shouldLog bool
		var requestBody []byte
		var apmTransaction *apm.Transaction

		defer func() {
			r.setCors(ctx)
		}()

		defer func() {
			var responseBody []byte

			if rpcResponse.Result != nil || rpcResponse.Error != nil {
				if respBody, err := json.Marshal(rpcResponse); err != nil {
					shouldLog = true
					rpcResponse.Result = nil
					rpcResponse.Error = &rpc.RpcError{
						Code:     error_codes.GenericMappingError,
						Message:  errors.Wrap(err, "error during response serialization").Error(),
						Data:     nil,
						Hostname: r.hostname,
					}
					if !r.isProd {
						rpcResponse.Error.Stack = fmt.Sprintf("%+v", err)
					}

					if respBody, err1 := json.Marshal(rpcResponse); err1 != nil {
						responseBody = []byte(fmt.Sprintf("that`s really not good || %v", err1.Error()))
					} else {
						responseBody = respBody
					}
				} else {
					responseBody = respBody
				}

				ctx.Response.SetBodyRaw(responseBody)
			}

			if rpcResponse.Error != nil {
				shouldLog = true
			}

			if shouldLog {
				r.logRequestBody(requestBody, apmTransaction)
				r.logResponseBody(responseBody, apmTransaction)
				r.logRpcResponseError(rpcResponse, apmTransaction)
			}
		}()

		requestBody = ctx.PostBody()

		if err := json.Unmarshal(requestBody, &rpcRequest); err != nil {
			rpcResponse.Error = &rpc.RpcError{
				Code:     error_codes.GenericMappingError,
				Message:  err.Error(),
				Data:     nil,
				Hostname: r.hostname,
			}

			if !r.isProd {
				rpcResponse.Error.Stack = fmt.Sprintf("%+v", err)
			}

			return
		}

		if traceHeader := ctx.Request.Header.Peek(apmhttp.W3CTraceparentHeader); len(traceHeader) > 0 {
			traceContext, _ := apmhttp.ParseTraceparentHeader(string(traceHeader))
			apmTransaction = apm_helper.StartNewApmTransactionWithTraceData(rpcRequest.Method, apmTxType, nil, traceContext)
		} else {
			apmTransaction = apm_helper.StartNewApmTransaction(rpcRequest.Method, apmTxType, nil, nil)
		}

		defer apmTransaction.End()

		cmd, err := endpoint.GetCommand(rpcRequest.Method)

		if err != nil {
			rpcResponse.Error = &rpc.RpcError{
				Code:     error_codes.CommandNotFoundError,
				Message:  err.Error(),
				Data:     nil,
				Hostname: r.hostname,
			}

			if !r.isProd {
				rpcResponse.Error.Stack = fmt.Sprintf("%+v", err)
			}

			return
		}

		rpcResponse, shouldLog = r.executeAction(rpcRequest, cmd, ctx, apmTransaction, cmd.ForceLog(), nil)
	})
}

func (r *HttpRouter) GET(path string, handler fasthttp.RequestHandler) {
	r.endpointRegistratorMutex.Lock()
	defer r.endpointRegistratorMutex.Unlock()

	r.realRouter.GET(path, handler)
}

func (r *HttpRouter) logRequestBody(body []byte, apmTransaction *apm.Transaction) {
	if apmTransaction == nil {
		return
	}

	if len(body) > 0 {
		apm_helper.AddApmData(apmTransaction, "request_body", body)
	}
}

func (r *HttpRouter) logResponseBody(responseBody []byte,
	apmTransaction *apm.Transaction) {
	if body := responseBody; len(body) > 0 {
		apm_helper.AddApmData(apmTransaction, "response_body", body)
	}
}

func (r *HttpRouter) logRpcResponseError(rpcResponse rpc.RpcResponse, apmTransaction *apm.Transaction) {
	if rpcResponse.Error != nil {
		apm_helper.CaptureApmError(rpcResponse.Error.ToError(), apmTransaction)
	}
}

func (r *HttpRouter) logUserValues(ctx *fasthttp.RequestCtx,
	apmTransaction *apm.Transaction) string {
	var realMethodName string

	ctx.VisitUserValues(func(key []byte, i interface{}) {
		keyStr := string(key)
		valueStr, ok := i.(string)

		if !ok { // not supported cast
			return
		}

		apm_helper.AddApmLabel(apmTransaction, keyStr, valueStr)
	})

	return realMethodName
}

func (r *HttpRouter) logRequestHeaders(ctx *fasthttp.RequestCtx,
	apmTransaction *apm.Transaction) {
	ctx.Request.Header.VisitAll(func(key, value []byte) {
		keyStr := strings.ToLower(string(key))

		if keyStr == "cookies" || keyStr == "authorization" || keyStr == "x-forwarded-client-cert" ||
			keyStr == "x-envoy-peer-metadata" || keyStr == "x-envoy-peer-metadata-id" {
			return
		}

		valueStr := string(value)

		if keyStr == "user-id" {
			keyStr = "user_id"
		} else if keyStr == "is-guest" {
			keyStr = "is_guest"
		}

		apm_helper.AddApmLabel(apmTransaction, keyStr, valueStr)
	})
}

func (r *HttpRouter) Router() *fastRouter.Router {
	return r.realRouter
}

func (r *HttpRouter) Handler() func(ctx *fasthttp.RequestCtx) {
	return fasthttp.CompressHandlerBrotliLevel(r.realRouter.Handler, 11, 9)
}

//func (r *HttpRouter) GetRpcRegisteredCommands() []Command {
//	var commands []Command
//
//	if r.executor.commands != nil {
//		for _, c := range r.executor.commands {
//			commands = append(commands, *c)
//		}
//	}
//
//	return commands
//}

func (r *HttpRouter) GetRestRegisteredCommands() []RestCommand {
	var commands []RestCommand

	if r.restCommands != nil {
		for _, c := range r.restCommands {
			commands = append(commands, *c)
		}
	}

	return commands
}

func (r *HttpRouter) StartAsync(port int) *HttpRouter {
	if r.srv != nil {
		return r
	}

	r.srv = &fasthttp.Server{
		Handler: fasthttp.CompressHandlerBrotliLevel(r.Handler(),
			fasthttp.CompressDefaultCompression, fasthttp.CompressDefaultCompression),
		MaxRequestBodySize: 4 * 1024 * 1024 * 100,
	}

	go func() {
		log.Info().Msgf("Http Server started on port [%v]", port)

		if err := r.srv.ListenAndServe(fmt.Sprintf("0.0.0.0:%v", port)); err != nil {
			panic(err)
		}
	}()

	return r
}
