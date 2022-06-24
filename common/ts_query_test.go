package common

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTsQuery(t *testing.T) {
	cases := []struct {
		input  string
		output string
	}{
		{
			input:  "vidjo ekstreme",
			output: "vidjo|ekstreme",
		},
		{
			input:  "vidjo!@#SASD ekstreme",
			output: "vidjosasd|ekstreme",
		},
		{
			input:  "girls ",
			output: "girls",
		},
		{
			input:  "girls   ",
			output: "girls",
		},
		{
			input:  "girls   vasya xer    xerov ",
			output: "girls|vasya|xer|xerov",
		},
	}

	for i, c := range cases {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			assert.Equal(t, c.output, ToTsQuery(c.input))
		})
	}
}
