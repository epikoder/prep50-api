package cache

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Prep50mobileApp/prep50-api/src/pkg/config"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/logger"
	"github.com/go-redis/redis/v8"
)

var (
	_redis *redis.Client
)

func init() {
	_redis = connect(config.Conf.Redis.Database)
}

func connect(db int) *redis.Client {
	var option *redis.Options
	option = &redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.Conf.Redis.Host, config.Conf.Redis.Port),
		Password: config.Conf.Redis.Password,
		DB:       db,
	}
	if redisUrl := os.Getenv("REDIS_URL"); redisUrl != "" {
		conf := strings.Split(strings.Split(redisUrl, "://:")[1], "@")
		option = &redis.Options{
			Addr:     conf[1],
			Password: conf[0],
			DB:       db,
		}
	}
	return redis.NewClient(option)
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
		logger.HandleError(err)
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
	if err != nil {
		logger.HandleError(err)
	}
	return err != redis.Nil
}

func Forget(key string) error {
	ctx := context.Background()
	defer ctx.Done()
	_, err := _redis.Del(ctx, key).Result()
	if err != nil {
		logger.HandleError(err)
	}
	return err
}

func Forever() {}

func Duration(unix int64) time.Duration {
	return time.Until(time.Unix(unix, 0))
}
