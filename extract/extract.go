package extract

import (
	"fmt"
	"github.com/shopspring/decimal"
	"gopkg.in/guregu/null.v4"
	"strconv"
	"strings"
)

func String(getFn func(key string) interface{}, key string, defaultValue string) string {
	val := getFn(key)

	if val == nil {
		return defaultValue
	}

	if v, ok := val.(string); !ok {
		return defaultValue
	} else {
		return v
	}
}

func Bool(getFn func(key string) interface{}, key string, defaultValue null.Bool) null.Bool {
	val := getFn(key)

	if val == nil {
		return defaultValue
	}

	if v, err := strconv.ParseBool(fmt.Sprint(val)); err != nil {
		return defaultValue
	} else {
		return null.NewBool(v, true)
	}
}

func Int64(getFn func(key string) interface{}, key string, defaultValue int64, maxValue int64) int64 {
	val := getFn(key)

	if val == nil {
		return defaultValue
	}

	if v, ok := val.(string); !ok {
		return defaultValue
	} else {
		if finalVal, err := strconv.ParseInt(v, 10, 64); err != nil {
			return defaultValue
		} else {
			if maxValue > 0 && finalVal > maxValue {
				return maxValue
			}

			return finalVal
		}
	}
}

func ArrayInt64(getFn func(key string) interface{}, key string, defaultValue int64, maxValue int64, separator string,
	skipDefault bool, duplicatesAllowed bool) []int64 {
	val := getFn(key)

	if val == nil {
		return []int64{}
	}

	if v, ok := val.(string); !ok {
		return []int64{}
	} else {
		if len(v) == 0 {
			return []int64{}
		}

		splitV := strings.Split(v, separator)

		resultArr := make([]int64, 0)

		for _, item := range splitV {
			parsedItem := Int64(func(key string) interface{} {
				return item
			}, "", defaultValue, maxValue)

			if skipDefault && parsedItem == defaultValue {
				continue
			}

			if duplicatesAllowed {
				resultArr = append(resultArr, parsedItem)
				continue
			}

			hasItem := false
			for _, rv := range resultArr {
				if rv == parsedItem {
					hasItem = true
					break
				}
			}

			if !hasItem {
				resultArr = append(resultArr, parsedItem)
			}
		}

		return resultArr
	}
}

func Decimal(getFn func(key string) interface{}, key string, defaultValue decimal.Decimal) decimal.Decimal {
	val := getFn(key)

	if val == nil {
		return defaultValue
	}

	if v, ok := val.(string); !ok {
		return defaultValue
	} else {
		if n, err := decimal.NewFromString(v); err != nil {
			return defaultValue
		} else {
			return n
		}
	}
}
