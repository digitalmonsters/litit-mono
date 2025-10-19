package wrappers

import (
	"github.com/digitalmonsters/go-common/error_codes"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestGetValidFromNodeJs(t *testing.T) {
	b := GetBaseWrapper()

	resp := <-b.SendRequestWithRpcResponseFromNodeJsService("https://user-info.dev.digitalmonster.link/mobile/v1/profile/2/getProfile",
		"GET", "application/json", "getProfile", nil, map[string]string{}, 10*time.Second, nil,
		"userService", false)

	assert.True(t, len(resp.Result) > 0)
	assert.Nil(t, resp.Error)
}

func TestGetInvalidFromNodeJs(t *testing.T) {
	b := GetBaseWrapper()

	resp := <-b.SendRequestWithRpcResponseFromNodeJsService("https://user-info.dev.digitalmonster.link/mobile/v1/profile/-2/getProfile",
		"GET", "application/json", "getProfile", nil, map[string]string{}, 10*time.Second, nil,
		"userService", false)

	assert.True(t, len(resp.Result) == 0)
	assert.NotNil(t, resp.Error)
	assert.Equal(t, error_codes.ErrorCode(404), resp.Error.Code)
	assert.Equal(t, "remote server [userService] replied with status: [404] and error: [User not found]", resp.Error.Message)
}

func TestGetInValidFromRpc(t *testing.T) {
	b := GetBaseWrapper()

	resp := <-b.SendRpcRequest("https://content.dev.digitalmonster.link/rpc-service",
		"ContentGetInternal", "POST", map[string]string{}, 3*time.Second, nil, "content", false)

	assert.True(t, len(resp.Result) > 0)
	assert.Nil(t, resp.Error)
}

func TestGetValidFromRpc(t *testing.T) {
	b := GetBaseWrapper()

	resp := <-b.SendRpcRequest("https://content.dev.digitalmonster.link/rpc-service",
		"ContentGetInternal", map[string]interface{}{
			"content_ids": []int{1, 2, 3, 4, 5, 6, 100, 55},
		}, map[string]string{}, 3*time.Second, nil, "content", false)

	assert.True(t, len(resp.Result) > 0)
	assert.Nil(t, resp.Error)
}
