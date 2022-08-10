package cache

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Prep50mobileApp/prep50-api/config"
	"github.com/go-redis/redis/v8"
)

var (
	_redis *redis.Client
)

func init() {
	_redis = connect(config.Conf.Redis.Database)
}

func connect(db int) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.Conf.Redis.Host, config.Conf.Redis.Port),
		Password: config.Conf.Redis.Password,
		DB:       db,
	})
}

func Set(key string, val interface{}, expires time.Duration) error {
	ctx := context.Background()
	defer ctx.Done()
	_, err := _redis.Set(ctx, key, val, expires).Result()
	return err
}

func Get(key string) (string, bool) {
	ctx := context.Background()
	defer ctx.Done()
	s, err := _redis.Get(ctx, key).Result()
	return s, err == nil
}

func Pull(key string) (string, bool) {
	ctx := context.Background()
	defer ctx.Done()
	s, err := _redis.GetDel(ctx, key).Result()
	if err != redis.Nil {
		if strings.Contains(err.Error(), "getdel") {
			v, ok := Get(key)
			Forget(key)
			return v, ok
		}
		return "", err == nil
	}
	return s, err == nil
}

func Exist(key string) bool {
	ctx := context.Background()
	defer ctx.Done()
	_, err := _redis.Get(ctx, key).Result()
	return err != redis.Nil
}

func Forget(key string) error {
	ctx := context.Background()
	defer ctx.Done()
	_, err := _redis.Del(ctx, key).Result()
	return err
}

func Forever() {}

func Duration(unix int64) time.Duration {
	return time.Until(time.Unix(unix, 0))
}
