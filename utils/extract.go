package utils

import "strconv"

func ExtractString(getFn func(key string) interface{}, key string, defaultValue string) string {
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

func ExtractInt64(getFn func(key string) interface{}, key string, defaultValue int64, maxValue int64) int64 {
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

