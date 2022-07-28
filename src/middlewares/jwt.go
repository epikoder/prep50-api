package middlewares

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/Prep50mobileApp/prep50-api/src/models"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/cache"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/ijwt"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/repository"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/middleware/jwt"
)

func Protected(ctx iris.Context) {
	claims, ok := jwt.Get(ctx).(*ijwt.JWTClaims)
	if !ok {
		ctx.StatusCode(http.StatusUnauthorized)
		return
	}
	userInfo, ok := claims.Audience.(map[string]interface{})
	if !ok {
		ctx.StatusCode(http.StatusUnauthorized)
		return
	}
	var user = new(models.User)
	if ok := repository.NewRepository(user).
		FindOne("email = ?", userInfo["email"]); !ok {
		ctx.StatusCode(http.StatusUnauthorized)
		return
	}
	authToken := strings.TrimSpace(strings.ReplaceAll(ctx.GetHeader("authorization"), "Bearer ", ""))
	access, ok := cache.Get(user.UserName + ".access")
	if !ok || access != authToken {
		ctx.StatusCode(http.StatusUnauthorized)
		return
	}

	if err := ctx.SetUser(user); err != nil {
		ctx.JSON(err)
		ctx.StatusCode(http.StatusAccepted)
		return
	}
	ctx.Next()
}

func AdminUser(ctx iris.Context) {
	user, _ := getUser(ctx)
	if !user.HasRole("admin") {
		ctx.StatusCode(http.StatusUnauthorized)
		return
	}
	ctx.Next()
}

func SuperAdminUser(ctx iris.Context) {
	user, _ := getUser(ctx)
	if !user.HasRole("super-admin") {
		ctx.StatusCode(http.StatusUnauthorized)
		return
	}
	ctx.Next()
}

func ResourcePermission(ctx iris.Context) {
	perm, ok := func(arr []string) (s string, ok bool) {
		l := len(arr)
		if ok = l >= 3; !ok {
			return
		}
		s = fmt.Sprintf("%s.%s", arr[l-2], arr[l-1])
		return
	}(strings.Split(ctx.Request().URL.String(), "/"))
	if !ok {
		ctx.StatusCode(http.StatusPreconditionFailed)
		return
	}
	user, _ := getUser(ctx)
	if !user.HasPermission(perm) {
		ctx.StatusCode(http.StatusUnauthorized)
		return
	}
	ctx.Next()
}
