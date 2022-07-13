package content_uploader

import (
	"github.com/digitalmonsters/go-common/rpc"
	"go.elastic.co/apm"
)

type ContentUploaderWrapperMock struct {
	UploadContentInternalFn func(url string, contentType string, data string, apmTransaction *apm.Transaction, forceLog bool) chan *rpc.RpcError
}

func (m ContentUploaderWrapperMock) UploadContentInternal(url string, contentType string, data string, apmTransaction *apm.Transaction, forceLog bool) chan *rpc.RpcError {
	return m.UploadContentInternalFn(url, contentType, data, apmTransaction, forceLog)
}

func GetMock() IContentUploaderWrapper { // for compiler errors
	return &ContentUploaderWrapperMock{}
}
