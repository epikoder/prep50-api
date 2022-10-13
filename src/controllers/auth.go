package controllers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/Prep50mobileApp/prep50-api/src/models"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/hash"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/ijwt"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/list"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/logger"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/repository"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/sendmail"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/validation"
	"github.com/Prep50mobileApp/prep50-api/src/services/database"
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

func QueryUsernameV1(ctx iris.Context) {
	if ok := repository.NewRepository(&models.User{}).FindOne("username = ? OR email = ?", ctx.URLParam("query"), ctx.URLParam("query")); ok {
		ctx.JSON(apiResponse{
			"status": "failed",
		})
		return
	}
	ctx.JSON(apiResponse{"status": "success"})
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

type (
	UserDeviceLoginForm struct {
		models.UserLoginFormStruct
		UserDeviceForm
	}
	UserWithExam struct {
		models.User
		Exam              []models.UserExam `json:"exams"`
		HasRegisteredExam bool              `json:"has_registered_exam"`
	}
	LoginResponse struct {
		*ijwt.JwtToken
		User interface{} `json:"user"`
	}
)

func LoginV1(ctx iris.Context) {
	data := UserDeviceLoginForm{}
	if err := ctx.ReadJSON(&data); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(validation.Errors(err))
		return
	}

	if data.UserName == "" {
		if data.User == "" {
			ctx.StatusCode(http.StatusBadRequest)
			ctx.JSON(apiResponse{
				"status":  "success",
				"message": "username or email is required",
			})
			return
		}
		data.UserName = data.User
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
			provider := &struct {
				Name string
			}{}
			if ok := database.UseDB("app").Table("user_providers as up").
				Select("up.*, p.name").Joins("LEFT JOIN providers as p ON up.provider_id = p.id").
				First(provider, "up.user_id = ?", user.Id).Error != nil; !ok {
				ctx.StatusCode(http.StatusNotAcceptable)
				ctx.JSON(apiResponse{
					"status":  "failed",
					"message": "Please use different login method",
				})
				return
			}

			ctx.StatusCode(http.StatusNotAcceptable)
			ctx.JSON(apiResponse{
				"status":  "failed",
				"message": fmt.Sprintf("Please use %s login", provider.Name),
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
	}

	{
		if user.Device.Id == uuid.Nil {
			device := &models.Device{}
			if ok := repository.NewRepository(device).FindOne("identifier = ?", data.DeviceId); ok && device.UserID != user.Id {
				queue.Dispatch(queue.Job{
					Type: queue.SendMail,
					Func: func() error {
						return sendmail.SendNewLoginMail(user)
					},
				})

				ctx.StatusCode(http.StatusForbidden)
				ctx.JSON(apiResponse{
					"status":  "failed",
					"code":    401,
					"message": "This device is currently registered to another user",
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

	userExams := []models.UserExam{}
	repository.NewRepository(&models.Exam{}).FindMany(&userExams, "user_id = ?", user.Id)

	userWithExam := &UserWithExam{*user, userExams, len(userExams) != 0}
	token, err := ijwt.GenerateToken(userWithExam, user.UserName)
	if !logger.HandleError(err) {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(internalServerError)
		return
	}

	response := LoginResponse{
		token, userWithExam,
	}

	ctx.JSON(apiResponse{
		"status":  "success",
		"message": "logged in successfully",
		"data":    response,
	})
}

type ProviderLogin struct {
	UserName string `validate:"alphanum"`
	Email    string `validate:"required,email"`
	Phone    string `validate:"required,numeric,min=8"`
	UserDeviceForm
}

func SocialV1(ctx iris.Context) {
	data := ProviderLogin{}
	if err := ctx.ReadJSON(&data); err != nil {
		ctx.StatusCode(400)
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

	_method := ctx.Params().Get("provider")
	var provider *models.Provider = &models.Provider{}
	if ok := repository.NewRepository(provider).FindOne("name = ?", _method); !ok {
		ctx.JSON(apiResponse{
			"status":  "failed",
			"message": "invalid authentication provider",
		})
		return
	}

	user := &models.User{}
	var token *ijwt.JwtToken
	userExist := repository.NewRepository(user).
		Preload("Device").
		Preload("Exams").
		FindOne("email = ?", data.Email)
	if userExist {
		if !user.IsProvider {
			ctx.StatusCode(http.StatusNotAcceptable)
			ctx.JSON(apiResponse{
				"status":  "failed",
				"message": "Please use your username and password to login",
			})
			return
		}

		if err := database.UseDB("app").
			First(&models.UserProvider{}, "user_id = ? AND provider_id = ?", user.Id, provider.Id).
			Error; err != nil {
			_provider := &struct {
				Name string
			}{}
			if notFound := database.UseDB("app").Table("providers as p").
				Select("up.*, p.name").Joins("LEFT JOIN user_providers as up ON up.provider_id = p.id").
				First(_provider, "up.user_id = ?", user.Id).Error != nil; notFound {
				ctx.StatusCode(http.StatusNotAcceptable)
				ctx.JSON(apiResponse{
					"status":  "failed",
					"message": "Please use different login method",
				})
				return
			}

			name := _provider.Name
			if name == "" {
				name = "different"
			}
			ctx.StatusCode(http.StatusNotAcceptable)
			ctx.JSON(apiResponse{
				"status":  "failed",
				"message": fmt.Sprintf("Please use %s login", name),
			})
			return
		}

		{
			if user.Device.Id == uuid.Nil {
				device := &models.Device{}
				if ok := repository.NewRepository(device).FindOne("identifier = ?", data.DeviceId); ok && device.UserID != user.Id {
					queue.Dispatch(queue.Job{
						Type: queue.SendMail,
						Func: func() error {
							return sendmail.SendNewLoginMail(user)
						},
					})

					ctx.StatusCode(http.StatusForbidden)
					ctx.JSON(apiResponse{
						"status":  "failed",
						"code":    401,
						"message": "This device is currently registered to another user",
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
	} else {
		device := &models.Device{}
		if ok := repository.NewRepository(device).FindOne("identifier = ?", data.DeviceId); ok {
			u := &models.User{}
			if ok := repository.NewRepository(u).FindOne("id = ?", device.UserID); ok {
				queue.Dispatch(queue.Job{
					Type: queue.SendMail,
					Func: func() error {
						return sendmail.SendNewLoginMail(u)
					},
				})
			}

			ctx.StatusCode(http.StatusForbidden)
			ctx.JSON(apiResponse{
				"status":  "failed",
				"code":    401,
				"message": "This device is currently registered to another user",
			})
			return
		}

		if data.UserName == "" {
			data.UserName = strings.Split(data.Email, "@")[0]
		}
		data.UserName = strings.ToLower(data.UserName)
		if ok := repository.NewRepository(user).FindOne("username = ?", data.UserName); ok {
			username, err := list.UniqueByField(user, data.UserName, "username")
			if err != nil {
				ctx.JSON(apiResponse{
					"status":  "failed",
					"message": "Unable to assign username",
				})
				return
			}
			data.UserName = username
		}
		user = &models.User{
			Id:       uuid.New(),
			UserName: data.UserName,
			Email:    data.Email,
			Phone:    data.Phone,
			Device: models.Device{
				Id:         uuid.New(),
				UserID:     user.Id,
				Identifier: data.DeviceId,
				Name:       data.DeviceName,
			},
			IsProvider: true,
		}
		if err := repository.NewRepository(user).Create(); !logger.HandleError(err) {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(internalServerError)
			return
		}

		if err := database.UseDB("app").Create(&models.UserProvider{
			UserId:     user.Id,
			ProviderId: provider.Id,
			IsLoggedIn: true,
		}).Error; err != nil {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(internalServerError)
			return
		}
	}

	userExams := []models.UserExam{}
	repository.NewRepository(&models.Exam{}).FindMany(&userExams, "user_id = ?", user.Id)
	userWithExam := &UserWithExam{*user, userExams, len(userExams) != 0}

	token, err := ijwt.GenerateToken(userWithExam, user.UserName)
	if !logger.HandleError(err) {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(internalServerError)
		return
	}

	response := LoginResponse{
		token, userWithExam,
	}
	ctx.JSON(apiResponse{
		"status":  "success",
		"message": "Logged in successfully",
		"data":    response,
	})
}
