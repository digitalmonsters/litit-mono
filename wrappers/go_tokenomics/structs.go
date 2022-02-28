package go_tokenomics

import (
	"fmt"
	"github.com/digitalmonsters/go-common/rpc"
	"github.com/shopspring/decimal"
)

type FilterField string
type FilterOperator string
type FilterValueType string

type GetUsersTokenomicsInfoRequest struct {
	UserIds []int64  `json:"user_ids"`
	Filters []Filter `json:"filters"`
}

type GetUsersTokenomicsInfoResponseChan struct {
	Items map[int64]UserTokenomicsInfo
	Error *rpc.RpcError
}

type Filter struct {
	Field     FilterField     `json:"field"`
	Operator  FilterOperator  `json:"operator"`
	ValueType FilterValueType `json:"value_type"`
	Value     interface{}     `json:"value"`
}

type UserTokenomicsInfo struct {
	TotalPoints        decimal.Decimal `json:"total_points"`
	CurrentPoints      decimal.Decimal `json:"current_points"`
	VaultPoints        decimal.Decimal `json:"vault_points"`
	AllTimeVaultPoints decimal.Decimal `json:"all_time_vault_points"`
	CurrentTokens      decimal.Decimal `json:"current_tokens"`
	CurrentRate        decimal.Decimal `json:"current_rate"`
	WithdrawnTokens    decimal.Decimal `json:"withdrawn_tokens"`
}

type SendMessageResponseChan struct {
	Error *rpc.RpcError `json:"error"`
}

const (
	TotalPoints        FilterField = "total_points"
	CurrentPoints      FilterField = "current_points"
	VaultPoints        FilterField = "vault_points"
	AllTimeVaultPoints FilterField = "all_time_vault_points"
	CurrentTokens      FilterField = "current_tokens"
	CurrentRate        FilterField = "current_rate"
	WithdrawnTokens    FilterField = "withdrawn_tokens"
)

const (
	More      FilterOperator = ">"
	Less      FilterOperator = "<"
	Equal     FilterOperator = "="
	MoreEqual FilterOperator = ">="
	LessEqual FilterOperator = "<="
)

const (
	Integer FilterValueType = "integer"
	Decimal FilterValueType = "decimal"
	String  FilterValueType = "string"
)

func GetFilterString(filter Filter) string {
	filterString := fmt.Sprintf("%v %v ", filter.Field, filter.Operator)
	switch filter.ValueType {
	case Integer:
		filterString += fmt.Sprint(filter.Value.(int))
	case Decimal:
		filterString += fmt.Sprint(filter.Value.(decimal.Decimal))
	case String:
		filterString += fmt.Sprintf("'%v'", filter.Value.(string))
	}
	return filterString
}

type GetWithdrawalsAmountsByAdminIdsRequest struct {
	AdminIds []int64 `json:"admin_ids"`
}

type GetWithdrawalsAmountsByAdminIdsResponseChan struct {
	Items map[int64]decimal.Decimal
	Error *rpc.RpcError
}

type GetContentEarningsTotalByContentIdsRequest struct {
	ContentIds []int64 `json:"content_ids"`
}

type GetContentEarningsTotalByContentIdsResponseChan struct {
	Items map[int64]decimal.Decimal
	Error *rpc.RpcError
}
