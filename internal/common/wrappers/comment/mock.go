package comment

import (
	"go.elastic.co/apm"
)

type CommentWrapperMock struct {
	GetCommentsInfoByIdFn func(commentIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetCommentsInfoByIdResponseChan
}

func (w *CommentWrapperMock) GetCommentsInfoById(commentIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetCommentsInfoByIdResponseChan {
	return w.GetCommentsInfoByIdFn(commentIds, apmTransaction, forceLog)
}

func GetMock() ICommentWrapper { // for compiler errors
	return &CommentWrapperMock{}
}
