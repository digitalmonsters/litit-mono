package filters

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/thoas/go-funk"
)

func GetFilterString(filter Filter) (string, error) {
	filterErr := errors.New("invalid filter settings")
	operator := FilterOperator(filter.Operator)
	possibleOperators := []FilterOperator{More, Less, Equal, NotEqual, MoreEqual, LessEqual, ILike}
	if !funk.Contains(possibleOperators, operator) {
		return "", filterErr
	}
	if operator == ILike && filter.ValueType != String {
		return "", filterErr
	}
	filterString := fmt.Sprintf("%v %v ", filter.Field, operator)
	switch filter.ValueType {
	case Integer, Decimal:
		filterString += filter.Value
	case String:
		formatString := "'%v'"
		if operator == ILike {
			formatString = "'%%%v%%'"
		}
		filterString += fmt.Sprintf(formatString, filter.Value)
	}
	return filterString, nil
}
