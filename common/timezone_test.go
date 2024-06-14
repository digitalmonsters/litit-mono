package common

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestConvertTimezoneToGo(t *testing.T) {
	a := assert.New(t)

	timezone, offset := ConvertTimezoneToGo("Gmt+3:00")

	a.Equal("+0300", timezone)
	a.Equal(3*60*60, offset)

	timezone, offset = ConvertTimezoneToGo("Gmt-05")

	a.Equal("-0500", timezone)
	a.Equal(-5*60*60, offset)

	timezone, offset = ConvertTimezoneToGo("Gmt-05:45")

	a.Equal("-0545", timezone)
	a.Equal(-(5*60+45)*60, offset)

	timezone, offset = ConvertTimezoneToGo("Gmt+0:00")

	a.Equal("0000", timezone)
	a.Equal(0, offset)

	timezone, offset = ConvertTimezoneToGo("Gmt")

	a.Equal("0000", timezone)
	a.Equal(0, offset)

	timezone, offset = ConvertTimezoneToGo("")

	a.Equal("0000", timezone)
	a.Equal(0, offset)

	timezone, offset = ConvertTimezoneToGo("Gmt0")

	a.Equal("0000", timezone)
	a.Equal(0, offset)

	timezone, offset = ConvertTimezoneToGo("Gmt00:00")

	a.Equal("0000", timezone)
	a.Equal(0, offset)

	timezone, offset = ConvertTimezoneToGo("Gmt-0000:0000")

	a.Equal("0000", timezone)
	a.Equal(0, offset)
}

func TestConvertDateToTimezone(t *testing.T) {
	date := time.Date(2022, 7, 12, 18, 0, 0, 0, time.UTC)
	dateConverted := ShiftDateToTimezone(date, "GMT+3:00")

	a := assert.New(t)

	a.Equal("2022-07-12 21:00:00 +0000 UTC", dateConverted.String())

	dateConverted = ShiftDateToTimezone(date, "GMT+0:00")

	a.Equal("2022-07-12 18:00:00 +0000 UTC", dateConverted.String())
}
