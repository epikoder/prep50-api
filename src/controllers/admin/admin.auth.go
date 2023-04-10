package admin

import (
	"net/http"

	"github.com/Prep50mobileApp/prep50-api/src/models"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/hash"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/ijwt"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/logger"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/repository"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/validation"
	"github.com/kataras/iris/v12"
)

func AdminLogin(ctx iris.Context) {
	data := models.UserLoginFormStruct{}
	if err := ctx.ReadJSON(&data); !logger.HandleError(err) {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(validation.Errors(err))
		return
	}

	var user = &models.User{}
	{
		if ok := repository.NewRepository(user).
			Preload("Roles.Permissions").
			Preload("Permissions").
			FindOne("username = ? OR email = ?", data.UserName, data.UserName); !ok {
			ctx.StatusCode(http.StatusUnauthorized)
			ctx.JSON(apiResponse{
				"status":  "failed",
				"message": "account not found",
			})
			return
		}
		if ok := hash.CheckHash(user.Password, data.Password); !ok {
			ctx.StatusCode(http.StatusUnauthorized)
			ctx.JSON(apiResponse{
				"status":  "failed",
				"message": "invalid username or password",
			})
			return
		}

		if !user.HasRole("admin") {
			ctx.StatusCode(http.StatusUnauthorized)
			ctx.JSON(apiResponse{
				"status":  "failed",
				"message": "unauthorized access",
			})
			return
		}
	}

	permissions := []string{}
	for _, p := range user.Permissions {
		permissions = append(permissions, p.Name)
	}
	roles := []string{}
	for _, r := range user.Roles {
		roles = append(roles, r.Name)
	}
	userAdmin := &models.AdminUser{User: *user, Permissions: permissions, Roles: roles}
	token, err := ijwt.GenerateToken(&models.AdminUser{User: *user, Permissions: permissions, Roles: roles}, user.UserName)
	if !logger.HandleError(err) {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(internalServerError)
		return
	}

	ctx.JSON(apiResponse{
		"status":  "success",
		"message": "logged in successfully",
		"data":    ijwt.LoginResponse{JwtToken: token, User: userAdmin},
	})
}
