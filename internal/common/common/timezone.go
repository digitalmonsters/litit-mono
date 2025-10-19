package common

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ConvertTimezoneToGo expected GMT+3:00 -> +0300
func ConvertTimezoneToGo(timezone string) (string, int) {
	if len(timezone) == 0 {
		return "0000", 0
	}

	if len(timezone) > 9 {
		timezone = timezone[0:9]
	}

	timezoneWithoutPrefix := regexp.MustCompile(`(?i)(gmt)`).ReplaceAllString(timezone, "")

	if len(timezoneWithoutPrefix) == 0 {
		return "0000", 0
	}

	timezoneWithoutPrefixSplit := strings.Split(timezoneWithoutPrefix, ":")

	minutes := "00"

	if len(timezoneWithoutPrefixSplit) > 1 {
		minutes = timezoneWithoutPrefixSplit[1]

		if len(minutes) > 2 {
			minutes = minutes[0:2]
		}

		minutes = regexp.MustCompile(`\D`).ReplaceAllString(minutes, "")
		minutes = fmt.Sprintf("%02s", minutes)
	}

	hours := timezoneWithoutPrefixSplit[0]

	if len(hours) > 3 {
		hours = hours[0:3]
	}

	matched, _ := regexp.MatchString(`[-+]`, string(hours[0]))
	sign := "+"

	if matched {
		sign = string(hours[0])

		if len(hours) > 3 {
			hours = hours[1:3]
		}
	} else if len(hours) > 2 {
		hours = hours[0:2]
	}

	hours = regexp.MustCompile(`\D`).ReplaceAllString(hours, "")
	hours = fmt.Sprintf("%02s", hours)

	if strings.Contains(fmt.Sprintf("%v%v", hours, minutes), "0000") {
		return "0000", 0
	}

	signInt, _ := strconv.Atoi(fmt.Sprintf("%v1", sign))
	hoursInt, _ := strconv.Atoi(hours)
	minutesInt, _ := strconv.Atoi(minutes)

	if hoursInt >= 24 || hoursInt < 0 {
		return "0000", 0
	}

	if minutesInt >= 60 || minutesInt < 0 {
		minutesInt = 0
		minutes = "00"
	}

	if strings.Contains(fmt.Sprintf("%v%v", hours, minutes), "0000") {
		return "0000", 0
	}

	offset := (hoursInt*60 + minutesInt) * 60
	offset = signInt * offset

	timezoneConverted := fmt.Sprintf("%v%v", sign, fmt.Sprintf("%v%v", hours, minutes))

	return timezoneConverted, offset
}

func ShiftDateToTimezone(date time.Time, timezone string) time.Time {
	_, timezoneOffset := ConvertTimezoneToGo(timezone)

	return date.Add(time.Duration(timezoneOffset) * time.Second)
}
