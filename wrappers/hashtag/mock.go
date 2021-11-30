package hashtag

import "go.elastic.co/apm"

type HashtagWrapperMock struct {
	GetHashtagsInternalFn func(hashtags []string, omitHashtags []string, limit int, offset int, apmTransaction *apm.Transaction, forceLog bool) chan HashtagsGetInternalResponseChan
}

func (w *HashtagWrapperMock) GetHashtagsInternal(hashtags []string, omitHashtags []string, limit int, offset int, apmTransaction *apm.Transaction, forceLog bool) chan HashtagsGetInternalResponseChan {
	return w.GetHashtagsInternalFn(hashtags, omitHashtags, limit, offset, apmTransaction, forceLog)
}

func GetMock() IHashtagWrapper {
	return &HashtagWrapperMock{}
}
