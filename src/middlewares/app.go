package middlewares

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/Prep50mobileApp/prep50-api/config"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/list"
	"github.com/iris-contrib/middleware/throttler"
	"github.com/kataras/iris/v12"

	"github.com/throttled/throttled/v2"
	"github.com/throttled/throttled/v2/store/memstore"
)

func init() {
	rand.Seed(time.Now().Unix())
}

func CORS(ctx iris.Context) {
	origin := func() string {
		i := config.Conf.App.Host
		protocol := func() string {
			if os.Getenv("APP_ENV") == "local" {
				return "http://"
			}
			return "https://"
		}()

		ref := func() string {
			arr := strings.Split(ctx.Request().Referer(), "//")
			if len(arr) < 2 {
				return config.Conf.App.Host
			}
			arr2 := strings.Split(arr[1], "/")
			if len(arr2) < 2 {
				return strings.Trim(arr[1], "/")
			}
			return strings.Trim(arr2[0], "/")
		}()
		if list.Contains(config.Conf.Cors.Origins, ref) {
			return fmt.Sprintf("%s%s", protocol, ref)
		}
		return fmt.Sprintf("%s%s:%d", protocol, i, config.Conf.App.Port)
	}
	ctx.Header("Vary", "Access-Control-Request-Method")
	ctx.Header("Access-Control-Allow-Origin", origin())
	ctx.Header("Access-Control-Request-Headers", config.Conf.Cors.Headers)
	ctx.Header("Access-Control-Allow-Headers", config.Conf.Cors.Headers)
	ctx.Header("Access-Control-Allow-Credentials", config.Conf.Cors.AllowCredentials)

	if ctx.Method() == iris.MethodOptions {
		ctx.Header("Access-Control-Methods",
			"POST, PUT, PATCH, DELETE")

		ctx.Header("Access-Control-Allow-Headers",
			"Access-Control-Allow-Origin,Content-Type")

		ctx.Header("Access-Control-Max-Age",
			"86400")

		ctx.StatusCode(iris.StatusNoContent)
		return
	}
	ctx.Next()
}

func RateLimiter() iris.Handler {
	store, err := memstore.New(65536)
	if err != nil {
		panic(err)
	}

	quota := throttled.RateQuota{
		MaxRate:  throttled.PerMin(config.Conf.Throttle.Limit),
		MaxBurst: config.Conf.Throttle.Burst,
	}

	rateLimiter, err := throttled.NewGCRARateLimiter(store, quota)
	if err != nil {
		panic(err)
	}

	httpRateLimiter := throttler.RateLimiter{
		RateLimiter: rateLimiter,
		VaryBy:      &throttled.VaryBy{Path: true},
	}

	return httpRateLimiter.RateLimit
}
