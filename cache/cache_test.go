package cache

import (
	"context"
	"fmt"
	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestGetSetSingle(t *testing.T) {
	ctx := context.Background()
	service := New(time.Minute*60, boilerplate.RedisConfig{
		Host: "127.0.0.1",
		Port: 6379,
		Db:   1,
	}, ctx)

	service.redis.FlushDB(ctx)

	key := "test"
	value := "test_value"

	err := service.Set(key, value, ctx, time.Minute*20, nil)
	assert.Nil(t, err)

	result, err := service.Get(key, ctx, nil)
	assert.Nil(t, err)
	assert.Equal(t, result.(string), value)

	local, ok := service.cache.Get(key)
	assert.True(t, ok)

	assert.Equal(t, local.(string), value)
}

func TestGetSetBulk(t *testing.T) {
	ctx := context.Background()
	service := New(time.Minute*60, boilerplate.RedisConfig{
		Host: "127.0.0.1",
		Port: 6379,
		Db:   1,
	}, ctx)

	service.redis.FlushDB(ctx)

	type testStruct struct {
		Id   int
		Name string
	}

	var toCache = map[string]interface{}{
		"test1": testStruct{
			Id:   1,
			Name: "test1_value",
		},
		"test2": testStruct{
			Id:   2,
			Name: "test2_value",
		},
	}

	err := service.BulkSet(toCache, time.Minute*60, ctx, nil)
	assert.Nil(t, err)

	res := service.BulkGet([]string{"test1", "test2", "test3"}, ctx, nil)
	assert.Len(t, res, 2)
	assert.Equal(t, toCache, res)

	for key, value := range toCache {
		local, ok := service.cache.Get(key)
		assert.True(t, ok)
		assert.Equal(t, value, local)
	}
}

func TestNew2(t *testing.T) {
	ctx := context.Background()
	service := New(time.Minute*60, boilerplate.RedisConfig{
		Host: "127.0.0.1",
		Port: 6379,
		Db:   1,
	}, ctx)

	service.redis.FlushDB(ctx)

	key := "test"
	value := "test_value"

	err := service.Set(key, value, ctx, time.Second*60, nil)
	assert.Nil(t, err)

	time.Sleep(time.Second * 6)
	v, ok := service.cache.Get(key)
	assert.True(t, ok)
	assert.NotNil(t, v)
}

func TestBenchmark(t *testing.T) {
	ctx := context.Background()
	service := New(time.Minute*60, boilerplate.RedisConfig{
		Host: "127.0.0.1",
		Port: 6379,
		Db:   1,
	}, ctx)

	service.redis.FlushDB(ctx)

	count := 100000
	result := make(map[string]interface{})

	for i := 0; i < count; i++ {
		result[fmt.Sprint(i)] = i
	}

	err := service.BulkSet(result, time.Minute*1, ctx, nil)
	assert.Nil(t, err)
}
