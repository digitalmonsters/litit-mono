package content

import (
	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/stretchr/testify/assert"
	"gopkg.in/guregu/null.v4"
	"testing"
)

func TestSendRequest(t *testing.T) {
	w := NewContentWrapper(boilerplate.WrapperConfig{
		ApiUrl:     "https://content.dev.digitalmonster.link",
		TimeoutSec: 0,
	})

	cc := <-w.GetAllCategories([]int64{}, true, nil, false)

	assert.Nil(t, cc.Error)
	assert.NotEqual(t, len(cc.Response), 0)
}

func TestSendRequestArr(t *testing.T) {
	w := NewContentWrapper(boilerplate.WrapperConfig{
		ApiUrl:     "https://content.dev.digitalmonster.link",
		TimeoutSec: 0,
	})

	cc := <-w.GetCategoryInternal([]int64{1, 23, 4, 5, 6}, nil, 10, 0, null.BoolFrom(false),
		null.BoolFrom(false), nil, false, false)

	assert.Nil(t, cc.Error)
	var ccResponse = cc.Response
	assert.NotEqual(t, len(ccResponse.Items), 0)
}
