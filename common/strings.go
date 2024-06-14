package common

import "strings"

func ToValidUtf8(input string) string {
	return strings.ToValidUTF8(strings.ReplaceAll(input, "\x00", ""), "")
}
