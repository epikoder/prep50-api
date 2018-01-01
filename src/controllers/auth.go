package controllers

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/Prep50mobileApp/prep50-api/src/models"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/hash"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/ijwt"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/repository"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/validation"
	"github.com/google/uuid"
	"github.com/kataras/iris/v12"
)

type (
	apiResponse map[string]interface{}
)

const googleOAuthUrl = "https://www.googleapis.com/userinfo/v2/me"

func QueryUsernameV1(ctx iris.Context) {

}

func RegisterV1(ctx iris.Context) {
	data := models.UserRegisterFormStruct{}
	if err := ctx.ReadJSON(&data); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(validation.Errors(err))
		return
	}

	p, err := hash.MakeHash(data.Password)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(apiResponse{
			"status":  "failed",
			"message": "error occcured",
		})
		return
	}

	var user *models.User = &models.User{}
	{
		if ok := repository.NewRepository(user).FindByField("username = ?", data.UserName); ok {
			ctx.StatusCode(http.StatusBadRequest)
			ctx.JSON(apiResponse{
				"status":  "failed",
				"message": "validation failed",
				"error": apiResponse{
					"username": "username already exist",
				},
			})
			return
		}
		if ok := repository.NewRepository(user).FindByField("email = ?", data.Email); ok {
			ctx.StatusCode(http.StatusBadRequest)
			ctx.JSON(apiResponse{
				"status":  "failed",
				"message": "validation failed",
				"error": apiResponse{
					"email": "email already exist",
				},
			})
			return
		}
		if ok := repository.NewRepository(user).FindByField("phone = ?", data.Phone); ok {
			ctx.StatusCode(http.StatusBadRequest)
			ctx.JSON(apiResponse{
				"status":  "failed",
				"message": "validation failed",
				"error": map[string]string{
					"phone": "phone already exist",
				},
			})
			return
		}
	}
	user = &models.User{
		Id:       uuid.New(),
		UserName: data.UserName,
		Email:    data.Email,
		Phone:    data.Phone,
		Password: p,
	}
	if err := repository.NewRepository(user).Create(); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(apiResponse{
			"status":  "failed",
			"message": "error occcured",
		})
		return
	}
	ctx.JSON(apiResponse{
		"status":  "success",
		"message": "Account created successfully",
		"data": apiResponse{
			"user": user,
		},
	})
}

func LoginV1(ctx iris.Context) {
	data := models.UserLoginFormStruct{}
	if err := ctx.ReadJSON(&data); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(validation.Errors(err))
		return
	}

	user := &models.User{}
	{
		if ok := repository.NewRepository(user).FindByField("username = ?", data.UserName); !ok {
			ctx.StatusCode(http.StatusUnauthorized)
			ctx.JSON(apiResponse{
				"status":  "failed",
				"message": "account not found",
			})
			return
		}

		if user.IsProvider {
			ctx.StatusCode(http.StatusNotAcceptable)
			// ctx.JSON(apiResponse{
			// 	"status":  "failed",
			// 	"message": "",
			// })
			ctx.JSON(user.Providers)
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
	}
	token, err := ijwt.GenerateToken(user)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(apiResponse{
			"status":  "failed",
			"message": "error occcured",
		})
		return
	}
	ctx.JSON(apiResponse{
		"status":  "success",
		"message": "logged in successfully",
		"data":    token,
	})
}

func SocialV1(ctx iris.Context) {
	provider := ctx.Params().Get("provider")
	var providers []*models.Provider = []*models.Provider{}
	repository.NewRepository(&models.Provider{}).Get(&providers)
	fmt.Println(providers)
	switch provider {
	case "google":
		type GoogleAuth struct {
			Code string `validate:"required"`
		}
		gA := GoogleAuth{}
		if err := ctx.ReadJSON(&gA); err != nil {
			ctx.StatusCode(http.StatusBadRequest)
			ctx.JSON(apiResponse{
				"status":  "failed",
				"message": validation.Errors(err),
			})
			return
		}

		// config := oauth2.Config{
		// 	ClientID:     "795485668384-n4hh34ibnc42cnrefmlih0ugerklt46v.apps.googleusercontent.com",
		// 	ClientSecret: "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",
		// 	Scopes: []string{
		// 		"https://www.googleapis.com/auth/userinfo.email",
		// 		"https://www.googleapis.com/auth/userinfo.profile",
		// 	},
		// 	Endpoint: oauth2.Endpoint{
		// 		AuthURL:  "https://accounts.google.com/o/oauth2/v2/auth",
		// 		TokenURL: "https://oauth2.googleapis.com/token",
		// 	},
		// }
		// token, err := config.Exchange(ctx.Request().Context(), gA.Code)
		// if err != nil {
		// 	ctx.StatusCode(http.StatusUnauthorized)
		// 	ctx.JSON(apiResponse{
		// 		"status":  "failed",
		// 		"message": err.Error(),
		// 	})
		// 	return
		// }
		res, err := http.Get(fmt.Sprintf("%s?access_token=%s", googleOAuthUrl, gA.Code))
		if err != nil {
			ctx.StatusCode(http.StatusUnauthorized)
			ctx.JSON(apiResponse{
				"status":  "failed",
				"message": err.Error(),
			})
			return
		}
		defer res.Body.Close()
		b, err := ioutil.ReadAll(res.Body)
		fmt.Println(string(b))
	}
}
