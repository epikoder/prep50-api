package middlewares

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/Prep50mobileApp/prep50-api/src/models"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/config"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/list"
	"github.com/iris-contrib/middleware/throttler"
	"github.com/kataras/iris/v12"

	"github.com/throttled/throttled/v2"
	"github.com/throttled/throttled/v2/store/memstore"
)

type apiResponse map[string]interface{}

const (
	xFrameOptions                = "X-Frame-Options"
	xFrameOptionsValue           = "DENY"
	xContentTypeOptions          = "X-Content-Type-Options"
	xContentTypeOptionsValue     = "nosniff"
	xssProtection                = "X-XSS-Protection"
	xssProtectionValue           = "1; mode=block"
	strictTransportSecurity      = "Strict-Transport-Security"
	strictTransportSecurityValue = "max-age=31536000; includeSubDomains; preload"
)

func Security(ctx iris.Context) {
	ctx.Header(xFrameOptions, xFrameOptionsValue)
	ctx.Header(xContentTypeOptions, xContentTypeOptionsValue)
	ctx.Header(xssProtection, xssProtectionValue)
	ctx.Header(strictTransportSecurity, strictTransportSecurityValue)
	ctx.Next()
}

var (
	internalServerError = apiResponse{
		"status":  "failed",
		"message": "error occcured",
	}

	getUser = func(ctx iris.Context) (u *models.User, err error) {
		i, err := ctx.User().GetRaw()
		if err != nil {
			return nil, err
		}
		var ok bool
		if u, ok = i.(*models.User); !ok {
			return nil, fmt.Errorf("user is nil")
		}
		return u, nil
	}
)

func init() {
	rand.Seed(time.Now().Unix())
}

func CORS(ctx iris.Context) {
	origin := func() string {
		i := config.Conf.App.Host
		env := os.Getenv("APP_ENV")
		protocol := func() string {
			if env != "" && env != "production" {
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
