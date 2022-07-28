package controllers

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/Prep50mobileApp/prep50-api/src/models"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/hash"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/ijwt"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/logger"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/repository"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/sendmail"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/validation"
	"github.com/Prep50mobileApp/prep50-api/src/services/queue"
	"github.com/google/uuid"
	"github.com/kataras/iris/v12"
)

type (
	apiResponse    map[string]interface{}
	UserDeviceForm struct {
		DeviceId   string `json:"device_id"`
		DeviceName string `json:"device_name"`
	}
)

const googleOAuthUrl = "https://www.googleapis.com/userinfo/v2/me"

func QueryUsernameV1(ctx iris.Context) {
	type QueryUsername struct {
		UserName string `validate:"required,alphanum"`
	}
	data := QueryUsername{}
	if err := ctx.ReadJSON(&data); err != nil {
		ctx.Write([]byte("0"))
		return
	}
	if ok := repository.NewRepository(&models.User{}).FindOne("username = ?", data.UserName); !ok {
		ctx.Write([]byte("0"))
		return
	}
	ctx.Write([]byte("1"))
}

func RegisterV1(ctx iris.Context) {
	data := models.UserRegisterFormStruct{}
	if err := ctx.ReadJSON(&data); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(validation.Errors(err))
		return
	}

	p, err := hash.MakeHash(data.Password)
	if !logger.HandleError(err) {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(internalServerError)
		return
	}

	var user *models.User = &models.User{}
	{
		if ok := repository.NewRepository(user).FindOne("username = ?", data.UserName); ok {
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
		if ok := repository.NewRepository(user).FindOne("email = ?", data.Email); ok {
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
		if ok := repository.NewRepository(user).FindOne("phone = ?", data.Phone); ok {
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
	var referral string = ""
	if data.Referral != "" {
		u := models.User{}
		if repository.NewRepository(&u).FindOne("username = ?", data.Referral) {
			referral = u.UserName
			defer func() {
				{
					// TODO: credit referral
				}
			}()
		}
	}

	user = &models.User{
		Id:       uuid.New(),
		UserName: data.UserName,
		Email:    data.Email,
		Phone:    data.Phone,
		Password: p,
		Referral: referral,
	}
	if err := repository.NewRepository(user).Create(); !logger.HandleError(err) {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(internalServerError)
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
	type (
		UserDeviceLoginForm struct {
			models.UserLoginFormStruct
			UserDeviceForm
		}
		userWithExam struct {
			models.User
			Exam         []models.UserExam `json:"exams"`
			RegisterExam bool              `json:"register_exam"`
		}
	)

	data := UserDeviceLoginForm{}
	if err := ctx.ReadJSON(&data); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(validation.Errors(err))
		return
	}

	if data.DeviceId == "" || data.DeviceName == "" {
		ctx.StatusCode(http.StatusForbidden)
		ctx.JSON(apiResponse{
			"status":  "failed",
			"code":    400,
			"message": "missing/invalid device info",
		})
		return
	}
	var user = &models.User{}
	{
		if ok := repository.NewRepository(user).
			Preload("Device").
			Preload("Exams").
			FindOne("username = ? OR email = ?", data.UserName, data.UserName); !ok {
			ctx.StatusCode(http.StatusUnauthorized)
			ctx.JSON(apiResponse{
				"status":  "failed",
				"message": "account not found",
			})
			return
		}

		if user.IsProvider {
			ctx.StatusCode(http.StatusNotAcceptable)
			// TODO: Return provider
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

	userExams := []models.UserExam{}
	repository.NewRepository(&models.Exam{}).FindMany(&userExams, "user_id = ?", user.Id)
	token, err := ijwt.GenerateToken(&userWithExam{*user, userExams, len(userExams) == 0}, user.UserName)
	if !logger.HandleError(err) {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(internalServerError)
		return
	}

	{
		if user.Device.Id == uuid.Nil {
			device := &models.Device{}
			if ok := repository.NewRepository(device).FindOne("identifier = ?", data.DeviceId); ok && device.UserID != user.Id {
				ctx.StatusCode(http.StatusForbidden)
				ctx.JSON(apiResponse{
					"status":  "failed",
					"code":    401,
					"message": "this device is currently registered to another user",
				})
				return
			}
			user.Device = models.Device{
				Id:         uuid.New(),
				UserID:     user.Id,
				Identifier: data.DeviceId,
				Name:       data.DeviceName,
			}

			if err := repository.NewRepository(&user.Device).Save(); !logger.HandleError(err) {
				ctx.StatusCode(http.StatusInternalServerError)
				ctx.JSON(internalServerError)
				return
			}
		} else if user.Device.Identifier != data.DeviceId ||
			!strings.EqualFold(strings.TrimSpace(strings.ToLower(user.Device.Name)), strings.TrimSpace(strings.ToLower(data.DeviceName))) {
			queue.Dispatch(queue.Job{
				Type: queue.SendMail,
				Func: func() error {
					return sendmail.SendNewDeviceMail(user)
				},
			})
			ctx.StatusCode(http.StatusForbidden)
			ctx.JSON(apiResponse{
				"status":  "failed",
				"code":    402,
				"message": "new device detected",
			})
			return
		}
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
			UserDeviceForm
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
		if err != nil {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(internalServerError)
			return
		}
		fmt.Println(string(b))
	}
}
