package filters

type FilterOperator string
type FilterValueType string

const (
	More      FilterOperator = ">"
	Less      FilterOperator = "<"
	Equal     FilterOperator = "="
	NotEqual  FilterOperator = "!="
	MoreEqual FilterOperator = ">="
	LessEqual FilterOperator = "<="
	ILike     FilterOperator = "ilike"
)

const (
	Integer FilterValueType = "integer"
	Decimal FilterValueType = "decimal"
	String  FilterValueType = "string"
)

type Filter struct {
	Field     string          `json:"field"`
	Operator  string          `json:"operator"`
	ValueType FilterValueType `json:"value_type"`
	Value     string          `json:"value"`
}
