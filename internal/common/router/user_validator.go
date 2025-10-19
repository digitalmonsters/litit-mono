package router

import (
	"context"
	"fmt"
	"github.com/digitalmonsters/go-common/translation"
	"github.com/digitalmonsters/go-common/wrappers/auth_go"
	"github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	"gopkg.in/guregu/null.v4"
	"time"
)

type UserExecutorValidationResponse struct {
	Id         int64                `json:"id"`
	Deleted    bool                 `json:"deleted"`
	BannedTill null.Time            `json:"banned_till"`
	Guest      bool                 `json:"guest"`
	Verified   bool                 `json:"verified"`
	Language   translation.Language `json:"language"`
}

type UserExecutorValidator interface {
	Validate(userId int64, ctx context.Context) (*UserExecutorValidationResponse, error)
}

type DefaultUserExecutorValidator struct {
	wrapper auth_go.IAuthGoWrapper
	cache   *cache.Cache
}

func NewDefaultUserExecutorValidator(wrapper auth_go.IAuthGoWrapper) UserExecutorValidator {
	return &DefaultUserExecutorValidator{
		wrapper: wrapper,
		cache:   cache.New(4*time.Minute, 5*time.Minute),
	}
}

func (v DefaultUserExecutorValidator) Validate(userId int64, ctx context.Context) (*UserExecutorValidationResponse, error) {
	key := fmt.Sprint(userId)

	if val, ok := v.cache.Get(key); ok {
		m := val.(UserExecutorValidationResponse)

		return &m, nil
	}

	usersResp := <-v.wrapper.InternalGetUsersForValidation([]int64{userId}, ctx, false)

	if usersResp.Error != nil {
		err := errors.Wrap(usersResp.Error.ToError(), "can not get user info from auth service")

		return nil, err
	}

	user, ok := usersResp.Response[userId]

	if !ok {
		err := errors.WithStack(errors.New("have no such user info"))

		return nil, err
	}

	resp := UserExecutorValidationResponse{
		Id:         user.Id,
		Deleted:    user.Deleted,
		BannedTill: user.BannedTill,
		Guest:      user.Guest,
		Verified:   user.Verified,
		Language:   user.Language,
	}

	v.cache.Set(key, resp, 2*time.Minute)

	return &resp, nil
}
