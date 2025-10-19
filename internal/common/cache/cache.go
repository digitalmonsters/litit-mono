package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/digitalmonsters/go-common/apm_helper"
	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/go-redis/redis/v8"
	"github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	"go.elastic.co/apm"
	"time"
)

type Service struct {
	dbNumber int
	cache    *cache.Cache
	redis    *redis.Client
}

func New(expiration time.Duration, redisConfig boilerplate.RedisConfig, ctx context.Context) *Service {
	s := &Service{
		cache: cache.New(expiration, expiration),
		redis: redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%v:%v", redisConfig.Host, redisConfig.Port),
			Password: redisConfig.Password,
			DB:       redisConfig.Db,
		}),
		dbNumber: redisConfig.Db,
	}

	s.ListenAsync(ctx)

	return s
}

func (s *Service) BulkGet(keys []string, ctx context.Context, apmTransaction *apm.Transaction) map[string]interface{} {
	result := make(map[string]interface{})
	var missingInCache []string

	for _, k := range keys {
		v, ok := s.getFromLocalCache(k)
		if ok {
			result[k] = v
			continue
		}

		missingInCache = append(missingInCache, k)
	}

	if len(missingInCache) == 0 {
		return result
	}

	redisResult, err := s.getFromRedisInternal(missingInCache, ctx)

	if err != nil {
		apm_helper.LogError(err, ctx)
	}

	for k, v := range redisResult {
		result[k] = v
	}

	return result
}

func (s *Service) Get(key string, ctx context.Context, apmTransaction *apm.Transaction) (interface{}, error) {
	res := s.BulkGet([]string{key}, ctx, apmTransaction)
	cached, ok := res[key]
	if !ok {
		return nil, errors.New("key is not found in cache")
	}

	return cached, nil
}

func (s *Service) Set(key string, value interface{}, ctx context.Context, expiration time.Duration, transaction *apm.Transaction) error {
	return s.BulkSet(map[string]interface{}{
		key: value,
	}, expiration, ctx, transaction)
}

func (s *Service) BulkSet(mapped map[string]interface{}, exp time.Duration, ctx context.Context, apmTransaction *apm.Transaction) error {
	var keys []string
	var values []interface{}

	for k, v := range mapped {
		keys = append(keys, k)
		values = append(values, v)

		s.cache.Set(k, v, exp)
	}

	return s.saveToRedisInternal(keys, values, exp, ctx, apmTransaction)
}

func (s *Service) saveToRedisInternal(keys []string, values []interface{}, exp time.Duration, ctx context.Context, apmTransaction *apm.Transaction) error {
	var toCache []interface{}
	pipe := s.redis.TxPipeline()

	for i := range keys {
		marshalled, err := json.Marshal(values[i])
		if err != nil {
			apm_helper.LogError(err, ctx)
			continue
		}

		toCache = append(toCache, keys[i], marshalled)
		pipe.Expire(ctx, keys[i], exp)
	}

	if err := s.redis.MSet(ctx, toCache...).Err(); err != nil {
		return err
	}
	if _, err := pipe.Exec(ctx); err != nil {
		return err
	}

	return nil
}

func (s *Service) getFromRedisInternal(keys []string, ctx context.Context) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	cacheData := s.redis.MGet(ctx, keys...)
	if cacheData.Err() != nil {
		return nil, cacheData.Err()
	}

	for index, value := range cacheData.Val() {
		if value != nil {
			result[keys[index]] = value
		}
	}

	return result, nil
}

func (s *Service) getFromLocalCache(key string) (interface{}, bool) {
	value, ok := s.cache.Get(key)
	if ok {
		return value, true
	}

	return nil, false
}

func (s *Service) ListenAsync(ctx context.Context) {
	subscriber := s.redis.PSubscribe(ctx, fmt.Sprintf("__keyevent@%v__:expired", s.dbNumber),
		fmt.Sprintf("__keyevent@%v__:del", s.dbNumber))
	go func() {
		for {
			msg, _ := subscriber.ReceiveMessage(ctx)
			s.cache.Delete(msg.Payload)
		}
	}()
}
