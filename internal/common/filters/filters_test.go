package filters

import (
	"fmt"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetFilterString(t *testing.T) {
	field := "test_field"
	intValue := 5
	filters := []Filter{
		{
			Field:     field,
			Operator:  string(LessEqual),
			ValueType: Integer,
			Value:     fmt.Sprint(intValue),
		},
		{
			Field:     field,
			Operator:  string(NotEqual),
			ValueType: String,
			Value:     "test",
		},
		{
			Field:     field,
			Operator:  string(ILike),
			ValueType: String,
			Value:     "test",
		},
		{
			Field:     field,
			Operator:  string(Equal),
			ValueType: Decimal,
			Value:     decimal.NewFromFloat(10.51).String(),
		},
	}

	filterStrings := []string{
		"test_field <= 5",
		"test_field != 'test'",
		"test_field ilike '%test%'",
		"test_field = 10.51",
	}

	a := assert.New(t)

	for i, filter := range filters {
		filterString, err := GetFilterString(filter)
		if err != nil {
			t.Fatal(err)
		}

		a.Equal(filterStrings[i], filterString)
	}

}
