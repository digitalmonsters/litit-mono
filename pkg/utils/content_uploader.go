package utils

import (
	"context"
	"fmt"
	"github.com/digitalmonsters/go-common/apm_helper"
	"net/url"
	"strings"
)

func RetrieveContentUploaderPath(ctx context.Context, inputUrl string) string {
	parsedUri, err := url.Parse(inputUrl)

	if err != nil {
		apm_helper.LogError(err, ctx)

		return inputUrl
	}

	repl := fmt.Sprintf("%v://%v", parsedUri.Scheme, parsedUri.Host)

	return strings.ReplaceAll(inputUrl, repl, "internal")
}
