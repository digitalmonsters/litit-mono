package azure_blob

import (
	"context"
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

	parsedUrl := strings.Split(inputUrl, "/")
	parsedUrl = parsedUrl[4:]

	return proxyUrl + "/" + strings.Join(parsedUrl, "/")
}
