package base_api

import "github.com/digitalmonsters/go-common/rpc"

type GetCountriesWithAgeLimitResponseChan struct {
	Error *rpc.RpcError
	Items []Country `json:"items"`
}

type Country struct {
	Name     string `json:"name"`
	Code     string `json:"code"`
	MinAge   int    `json:"min_age"`
	AdultAge int    `json:"adult_age"`
}
