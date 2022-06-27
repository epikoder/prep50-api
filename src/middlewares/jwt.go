package middlewares

import (
	"fmt"
	"net/http"

	"github.com/Prep50mobileApp/prep50-api/src/pkg/ijwt"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/middleware/jwt"
)

func Protected(ctx iris.Context) {
	claims, ok := jwt.Get(ctx).(ijwt.JWTClaims)
	if !ok {
		ctx.StatusCode(http.StatusUnauthorized)
		return
	}
	fmt.Println(claims)
	ctx.Next()
}

func AdminUser(ctx iris.Context) {

}
