package controllers

import (
	"net/http"
	"strings"
	"time"

	"github.com/Prep50mobileApp/prep50-api/src/models"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/cache"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/ijwt"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/logger"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/repository"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/middleware/jwt"
)

func Refresh(ctx iris.Context) {
	claims, ok := jwt.Get(ctx).(*ijwt.JWTClaims)
	if !ok {
		ctx.StatusCode(http.StatusUnauthorized)
		return
	}

	userInfo, ok := claims.Aud.(map[string]interface{})
	if !ok {
		ctx.StatusCode(401)
		return
	}
	var user = new(models.User)
	if ok := repository.NewRepository(user).
		FindOne("email = ?", userInfo["email"]); !ok {
		ctx.StatusCode(http.StatusUnauthorized)
		return
	}

	authToken := strings.TrimSpace(strings.TrimPrefix(ctx.GetHeader("authorization"), "Bearer "))
	access, ok := cache.Get(user.UserName + ".refresh")
	if !ok || access != authToken {
		ctx.JSON(apiResponse{
			"db": access,
			"a":  authToken,
		})
		ctx.StatusCode(http.StatusUnauthorized)
		return
	}

	shouldGenerateRefreshToken := time.Until(time.Unix(claims.Exp, 0)).Hours() < 60
	token, err := func() (*ijwt.JwtToken, error) {
		if shouldGenerateRefreshToken {
			return ijwt.GenerateToken(user, user.UserName)
		}
		return ijwt.RefreshToken(user, ijwt.JwtToken{Refresh: authToken, ExpiresRt: time.Unix(claims.Exp, 0)}, user.UserName)
	}()

	if !logger.HandleError(err) {
		ctx.StatusCode(401)
		return
	}

	ctx.JSON(apiResponse{
		"status": "success",
		"data":   token,
	})
}
