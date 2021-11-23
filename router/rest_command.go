package router

import "github.com/digitalmonsters/go-common/common"

type HttpMethodType string

const (
	MethodGet     = HttpMethodType("GET")     // RFC 7231, 4.3.1
	MethodHead    = HttpMethodType("HEAD")    // RFC 7231, 4.3.2
	MethodPost    = HttpMethodType("POST")    // RFC 7231, 4.3.3
	MethodPut     = HttpMethodType("PUT")     // RFC 7231, 4.3.4
	MethodPatch   = HttpMethodType("PATCH")   // RFC 5789
	MethodDelete  = HttpMethodType("DELETE")  // RFC 7231, 4.3.5
	MethodConnect = HttpMethodType("CONNECT") // RFC 7231, 4.3.6
	MethodOptions = HttpMethodType("OPTIONS") // RFC 7231, 4.3.7
	MethodTrace   = HttpMethodType("TRACE")   // RFC 7231, 4.3.8
)

type RestCommand struct {
	commandFn                 CommandFunc
	method                    string
	path                      string
	forceLog                  bool
	accessLevel               common.AccessLevel
	requireIdentityValidation bool
}

func (r RestCommand) RequireIdentityValidation() bool {
	return r.requireIdentityValidation
}

func (r RestCommand) AccessLevel() common.AccessLevel {
	return r.accessLevel
}

func NewRestCommand(commandFn CommandFunc, path string, httpMethod HttpMethodType, accessLevel common.AccessLevel, requiredIdentityValidation bool,
	forceLog bool) *RestCommand {
	return &RestCommand{commandFn: commandFn, path: path, method: string(httpMethod), forceLog: forceLog, accessLevel: accessLevel,
		requireIdentityValidation: requiredIdentityValidation}
}

func (r RestCommand) GetPath() string {
	return r.path
}

func (r RestCommand) GetHttpMethod() string {
	return r.method
}

type genericRestResponse struct {
	Data    interface{} `json:"data"`
	Success bool        `json:"success"`
	Error   string      `json:"error,omitempty"`
	Stack   string      `json:"stack,omitempty"`
}
