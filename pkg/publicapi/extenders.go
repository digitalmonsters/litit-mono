package publicapi

import (
	"errors"
	"github.com/digitalmonsters/go-common/apm_helper"
	"github.com/digitalmonsters/go-common/wrappers/user"
	"github.com/thoas/go-funk"
	"go.elastic.co/apm"
)

func extendWithAuthor(userInfoWrapper user.IUserWrapper, apmTransaction *apm.Transaction, comments ...*Comment) chan error {
	ch := make(chan error)

	go func() {
		defer func() {
			close(ch)
		}()

		if len(comments) == 0 {
			return
		}

		var authors []int64

		for _, comment := range comments {
			if !funk.ContainsInt64(authors, comment.AuthorId) {
				authors = append(authors, comment.AuthorId)
			}
		}

		if len(authors) > 0 {
			responseData := <-userInfoWrapper.GetUsers(authors, apmTransaction, false)

			if responseData.Error != nil {
				apm_helper.CaptureApmError(errors.New(responseData.Error.Message), apmTransaction) // todo fancy error
				return
			}

			for _, comment := range comments {
				if v, ok := responseData.Items[comment.AuthorId]; ok {
					comment.Author = Author{
						Id:        v.UserId,
						Username:  v.Username,
						Avatar:    v.Avatar,
						Firstname: v.Firstname,
						Lastname:  v.Lastname,
					}
				}
			}
		}
	}()

	return ch
}
