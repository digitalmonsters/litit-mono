package router

import (
	"encoding/json"
	"fmt"
	"github.com/digitalmonsters/go-common/apm_helper"
	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/digitalmonsters/go-common/error_codes"
	"github.com/digitalmonsters/go-common/rpc"
	fastRouter "github.com/fasthttp/router"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/valyala/fasthttp"
	"go.elastic.co/apm"
	"os"
	"strconv"
	"time"
)

type HttpRouter struct {
	realRouter   *fastRouter.Router
	executor     *CommandExecutor
	hostname     string
	restCommands map[string]*RestCommand
	isProd       bool
}

func NewRouter(rpcEndpointPath string) *HttpRouter {
	h := &HttpRouter{
		realRouter:   fastRouter.New(),
		executor:     NewCommandExecutor(),
		restCommands: map[string]*RestCommand{},
	}

	if hostname, _ := os.Hostname(); len(hostname) > 0 {
		h.hostname = hostname
	}

	if boilerplate.GetCurrentEnvironment() == boilerplate.Prod {
		h.isProd = true
	}

	h.prepareRpcEndpoint(rpcEndpointPath)

	return h
}

func (r *HttpRouter) RegisterRpcCommand(command *Command) error {
	return r.executor.AddCommand(command)
}

func (r *HttpRouter) RegisterRestCmd(targetCmd *RestCommand) error {
	key := fmt.Sprintf("%v_%v", targetCmd.method, targetCmd.path)

	if _, ok := r.restCommands[key]; ok {
		return errors.New(fmt.Sprintf("rest command [%v] already registered", key))
	}

	r.restCommands[key] = targetCmd

	r.realRouter.Handle(targetCmd.method, targetCmd.path, func(ctx *fasthttp.RequestCtx) {
		apmTransaction := apm_helper.StartNewApmTransaction(fmt.Sprintf("[%v] [%v]", targetCmd.method,
			targetCmd.path), "rpc", nil, nil)
		defer apmTransaction.End()

		requestBody := ctx.PostBody()

		rpcRequest := rpc.RpcRequest{
			Method:  targetCmd.path,
			Params:  requestBody,
			Id:      "1",
			JsonRpc: "2.0",
		}

		rpcResponse, shouldLog := r.executeAction(rpcRequest, targetCmd.commandFn, ctx, apmTransaction, targetCmd.forceLog,
			func(key string) interface{} {
				if v := ctx.UserValue(key); v != nil {
					return v
				}

				if ctx.QueryArgs() != nil {
					if v := ctx.QueryArgs().Peek(key); len(v) > 0 {
						return string(v)
					}
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
		}()

		finalStatusCode := int(error_codes.None)

		var err error

		var restResponse genericRestResponse

		restResponse.Success = true

		if rpcResponse.Result != nil {
			restResponse.Data = rpcResponse.Result
		}
		if rpcResponse.Error != nil {
			originalCode := int(rpcResponse.Error.Code)
			restResponse.Success = false
			restResponse.Error = rpcResponse.Error.Message
			restResponse.Stack = rpcResponse.Error.Stack

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

		if responseBody, err = json.Marshal(restResponse); err != nil {
			log.Err(err).Send()
		}

		ctx.Response.SetBodyRaw(responseBody)
		ctx.Response.SetStatusCode(finalStatusCode)
	})

	return nil
}

func (r *HttpRouter) executeAction(rpcRequest rpc.RpcRequest, cmdFn CommandFunc, ctx *fasthttp.RequestCtx,
	apmTransaction *apm.Transaction, forceLog bool, getUserValue func(key string) interface{}) (rpcResponse rpc.RpcResponse, shouldLog bool) {
	totalTiming := time.Now()
	apm.ContextWithTransaction(ctx, apmTransaction)

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
				Code:    error_codes.GenericPanicError,
				Message: panicErr.Error(),
				Data:    nil,
			}

			if !r.isProd {
				rpcResponse.Error.Stack = fmt.Sprintf("%+v", panicErr)
			}
		}
	}()

	rpcResponse.Id = rpcRequest.Id

	shouldLog = forceLog

	userId := int64(0)

	userIdFromHeader := ctx.Request.Header.Peek("UserId")

	if userIdFromHeader != nil {
		if id, err := strconv.ParseInt(string(userIdFromHeader), 10, 64); err != nil {
			// todo handle somehow
		} else {
			userId = id
		}
	}

	//if cmd.RequireIdentityValidation() && userId <= 0 { // some specific logic to properly handle authorization
	//	// todo handle somehow
	//}

	executionTiming := time.Now()

	if resp, err := cmdFn(rpcRequest.Params, MethodExecutionData{
		ApmTransaction: apmTransaction,
		Context:        ctx,
		UserId:         userId,
		getUserValueFn: getUserValue,
	}); err != nil {
		rpcResponse.Error = &rpc.RpcError{
			Code:    err.GetCode(),
			Message: err.GetMessage(),
			Data:    nil,
		}

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

func (r *HttpRouter) prepareRpcEndpoint(rpcEndpointPath string) {
	r.realRouter.POST(rpcEndpointPath, func(ctx *fasthttp.RequestCtx) {
		var rpcRequest rpc.RpcRequest
		var rpcResponse rpc.RpcResponse
		var shouldLog bool
		var requestBody []byte

		apmTransaction := apm_helper.StartNewApmTransaction(rpcRequest.Method, "rpc", nil, nil)
		defer apmTransaction.End()

		defer func() {
			var responseBody []byte

			if rpcResponse.Result != nil || rpcResponse.Error != nil {
				if respBody, err := json.Marshal(rpcResponse); err != nil {
					shouldLog = true
					rpcResponse.Result = nil
					rpcResponse.Error = &rpc.RpcError{
						Code:    error_codes.GenericMappingError,
						Message: errors.Wrap(err, "error during response serialization").Error(),
						Data:    nil,
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
			}
		}()

		requestBody = ctx.PostBody()

		if err := json.Unmarshal(requestBody, &rpcRequest); err != nil {
			rpcResponse.Error = &rpc.RpcError{
				Code:    error_codes.GenericMappingError,
				Message: err.Error(),
				Data:    nil,
			}

			if !r.isProd {
				rpcResponse.Error.Stack = fmt.Sprintf("%+v", err)
			}

			return
		}

		cmd, err := r.executor.GetCommand(rpcRequest.Method)

		if err != nil {
			rpcResponse.Error = &rpc.RpcError{
				Code:    error_codes.CommandNotFoundError,
				Message: err.Error(),
				Data:    nil,
			}

			if !r.isProd {
				rpcResponse.Error.Stack = fmt.Sprintf("%+v", err)
			}

			return
		}

		rpcResponse, shouldLog = r.executeAction(rpcRequest, cmd.fn, ctx, apmTransaction, cmd.forceLog, nil)
	})
}

func (r *HttpRouter) GET(path string, handler fasthttp.RequestHandler) {
	r.realRouter.GET(path, handler)
}

func (r *HttpRouter) logRequestBody(body []byte, apmTransaction *apm.Transaction) {
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
		keyStr := string(key)

		if keyStr == "cookies" || keyStr == "authorization" {
			return
		}

		valueStr := string(value)

		apm_helper.AddApmLabel(apmTransaction, keyStr, valueStr)
	})
}

func (r *HttpRouter) Router() *fastRouter.Router {
	return r.realRouter
}

func (r *HttpRouter) Handler() func(ctx *fasthttp.RequestCtx) {
	return r.realRouter.Handler
}

func (r *HttpRouter) GetRpcRegisteredCommands() []Command {
	var commands []Command

	if r.executor.commands != nil {
		for _, c := range r.executor.commands {
			commands = append(commands, *c)
		}
	}

	return commands
}

func (r *HttpRouter) GetRestRegisteredCommands() []RestCommand {
	var commands []RestCommand

	if r.restCommands != nil {
		for _, c := range r.restCommands {
			commands = append(commands, *c)
		}
	}

	return commands
}
