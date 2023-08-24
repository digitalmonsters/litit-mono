package application

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/digitalmonsters/go-common/http_client"
	"github.com/pkg/errors"
)

const (
	HttpRetrieverDefaultUrl = "http://configurator/internal/json"
)

type Retriever interface {
	Retrieve(keys []string, ctx context.Context) (map[string]string, error)
}

func NewHttpRetriever(apiUrl string) Retriever {
	return &HttpRetriever{apiUrl: apiUrl}
}

type HttpRetriever struct {
	apiUrl string
}

type getConfigRequest struct {
	Items []string `json:"items"`
}

func (h *HttpRetriever) Retrieve(keys []string, ctx context.Context) (map[string]string, error) {
	resp, err := http_client.DefaultHttpClient.NewRequest(ctx).
		SetBody(getConfigRequest{
			Items: keys,
		}).
		Post(h.apiUrl)

	if err != nil {
		return nil, errors.WithStack(fmt.Errorf("request err: %s", err.Error()))
	}

	result := map[string]string{}

	if err = json.Unmarshal(resp.Bytes(), &result); err != nil {
		return nil, errors.WithStack(fmt.Errorf("resp:%v, unmarshal err: %s", string(resp.Bytes()), err.Error()))
	}

	return result, nil
}

type FileRetriever struct {
	filePath string
}

func NewFileRetriever(filePath string) *FileRetriever {
	return &FileRetriever{
		filePath: filePath,
	}
}

func (f *FileRetriever) Retrieve(keys []string, ctx context.Context) (map[string]string, error) {
	data, err := os.ReadFile(f.filePath)

	if err != nil {
		return nil, errors.WithStack(fmt.Errorf("readFile err: %s", err.Error()))
	}

	mapped := map[string]string{}

	if err = json.Unmarshal(data, &mapped); err != nil {
		return nil, errors.WithStack(fmt.Errorf("resp:%v, unmarshal err: %s", string(data), err.Error()))
	}

	return mapped, nil
}
