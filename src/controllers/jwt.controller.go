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
		Preload("Roles.Permissions").
		Preload("Permissions").
		FindOne("email = ? ", userInfo["email"]); !ok {
		ctx.StatusCode(http.StatusUnauthorized)
		return
	}

	authToken := strings.TrimSpace(strings.TrimPrefix(ctx.GetHeader("authorization"), "Bearer "))
	access, ok := cache.Get(user.UserName + ".refresh")
	if !ok || access != authToken {
		ctx.StatusCode(http.StatusUnauthorized)
		return
	}

	shouldGenerateRefreshToken := time.Until(time.Unix(claims.Exp, 0)).Hours() < 60
	var userGeneric interface{}
	if user.IsAdmin || user.HasRole("admin") {
		permissions := []string{}
		for _, p := range user.Permissions {
			permissions = append(permissions, p.Name)
		}
		roles := []string{}
		for _, r := range user.Roles {
			roles = append(roles, r.Name)
		}
		userGeneric = &AdminUser{*user, permissions, roles}
	} else {
		userExams := []models.UserExam{}
		repository.NewRepository(&models.Exam{}).FindMany(&userExams, "user_id = ?", user.Id)
		userGeneric = &UserWithExam{*user, userExams, len(userExams) != 0}
	}

	token, err := func() (*ijwt.JwtToken, error) {
		if shouldGenerateRefreshToken {
			return ijwt.GenerateToken(userGeneric, user.UserName)
		}
		return ijwt.RefreshToken(userGeneric, ijwt.JwtToken{Refresh: authToken, ExpiresRt: time.Unix(claims.Exp, 0)}, user.UserName)
	}()

	if !logger.HandleError(err) {
		ctx.StatusCode(401)
		return
	}

	response := LoginResponse{token, userGeneric}
	ctx.JSON(apiResponse{
		"status": "success",
		"data":   response,
	})
}
