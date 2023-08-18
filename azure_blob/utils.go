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
	parsedUrl = append(parsedUrl[:3], parsedUrl[4:]...)

	return strings.Join(parsedUrl, "/")
}
