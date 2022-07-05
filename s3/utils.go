package s3

import (
	"context"
	"fmt"
	"github.com/digitalmonsters/go-common/apm_helper"
	"net/url"
	"strings"
)

func UseContentUploaderProxy(ctx context.Context, inputUrl string, proxyUrl string) string {
	proxyUrl = strings.TrimSuffix(proxyUrl, "/")

	if len(proxyUrl) == 0 {
		return inputUrl
	}

	if !strings.HasPrefix(proxyUrl, "http") {
		return inputUrl
	}

	parsedUri, err := url.Parse(inputUrl)

	if err != nil {
		apm_helper.LogError(err, ctx)

		return inputUrl
	}

	repl := fmt.Sprintf("%v://%v", parsedUri.Scheme, parsedUri.Host)

	return strings.ReplaceAll(inputUrl, repl, proxyUrl)
}
