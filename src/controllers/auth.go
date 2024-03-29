package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Prep50mobileApp/prep50-api/src/models"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/cache"
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
	q := ctx.URLParam("query")
	if ok := repository.NewRepository(&models.User{}).FindOne("username = ? OR email = ?", q, q); ok && len(q) > 3 {
		ctx.JSON(apiResponse{
			"status":  "failed",
			"message": "username or email is invalid",
		})
		return
	}
	ctx.JSON(apiResponse{"status": "success"})
}

func RegisterV1(ctx iris.Context) {
	data := models.UserRegisterFormStruct{}
	if err := ctx.ReadJSON(&data); !logger.HandleError(err) {
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
	token, err := ijwt.GenerateToken(user, user.UserName)
	if !logger.HandleError(err) {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(internalServerError)
		return
	}

	user.Exams = []models.Exam{}
	ctx.JSON(apiResponse{
		"status":  "success",
		"message": "Account created successfully",
		"data": ijwt.LoginResponse{
			JwtToken: token, User: user,
		},
	})
}

type (
	UserDeviceLoginForm struct {
		models.UserLoginFormStruct
		UserDeviceForm
	}
	UserExamWithName struct {
		Id            uuid.UUID            `json:"-"`
		Name          string               `json:"exam"`
		Session       uint                 `json:"session"`
		PaymentStatus models.PaymentStatus `json:"payment_status"`
		CreatedAt     time.Time            `json:"created_at"`
		ExpiresAt     *time.Time           `json:"expires_at"`
	}
	UserWithExam struct {
		models.User
		Exam              []UserExamWithName `json:"exams"`
		HasRegisteredExam bool               `json:"has_registered_exam"`
	}
)

type NewDeviceMail struct {
	DeviceId   string
	DeviceName string
	Username   string
	UserId     string
	Expires    time.Time
}

func creeteDeviceMail(userId, username, dId, dName string, expires time.Time) (id string, err error) {
	b, err := json.Marshal(NewDeviceMail{
		DeviceId:   dId,
		DeviceName: dName,
		Username:   username,
		Expires:    expires,
		UserId:     userId,
	})
	if err != nil {
		return
	}
	id = uuid.New().String()
	err = cache.Set(id, string(b), cache.Duration(expires.Unix()))
	if err != nil {
		return
	}
	return
}

var deviceExistError = apiResponse{
	"status":  "failed",
	"code":    401,
	"message": "This device is currently registered to another user",
}

func deviceExist(ctx iris.Context, userId uuid.UUID, deviceId string) (err error) {
	dev := &struct {
		Identifier string
		Name       string
		models.User
	}{}
	if err = database.DB().
		Table("users as u").
		Select("u.*, d.name, d.identifier").
		Joins("LEFT JOIN devices as d ON u.id = d.user_id").
		First(&dev, "identifier = ?", deviceId).Error; err == nil && dev.Id != userId {

		var i string
		var err error
		if i, err = creeteDeviceMail(dev.Id.String(),
			dev.UserName,
			dev.Identifier,
			dev.Name,
			time.Now().Add(time.Minute*10)); err != nil {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(internalServerError)
			return err
		}
		queue.Dispatch(queue.Job{
			Type: queue.SendMail,
			Func: func() error {
				return sendmail.SendNewDeviceMail(&dev.User, i)
			},
		})
		return fmt.Errorf("device exist")
	}
	return nil
}

func LoginV1(ctx iris.Context) {
	data := UserDeviceLoginForm{}
	if err := ctx.ReadJSON(&data); !logger.HandleError(err) {
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
			if ok := database.DB().Table("user_providers as up").
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
			if err := deviceExist(ctx, user.Id, data.DeviceId); err != nil {
				ctx.StatusCode(http.StatusForbidden)
				ctx.JSON(deviceExistError)
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
		} else if len(strings.TrimSpace(user.Device.Identifier)) > 0 && (user.Device.Identifier != data.DeviceId ||
			!strings.EqualFold(strings.TrimSpace(strings.ToLower(user.Device.Name)), strings.TrimSpace(strings.ToLower(data.DeviceName)))) {

			var i string
			var err error
			if i, err = creeteDeviceMail(user.Id.String(),
				user.UserName,
				user.Device.Identifier,
				user.Device.Name,
				time.Now().Add(time.Minute*10)); !logger.HandleError(err) {
				ctx.StatusCode(http.StatusInternalServerError)
				ctx.JSON(internalServerError)
				return
			}
			queue.Dispatch(queue.Job{
				Type: queue.SendMail,
				Func: func() error {
					return sendmail.SendNewDeviceMail(user, i)
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

	userExams := []UserExamWithName{}
	database.DB().Table("user_exams as ue").
		Select("ue.session, ue.payment_status, ue.created_at, ue.id, e.name, ue.expires_at").Joins("LEFT JOIN exams as e ON e.id = ue.exam_id").
		Where("user_id = ?", user.Id).
		Scan(&userExams)

	userWithExam := &UserWithExam{*user, userExams, len(userExams) != 0}
	token, err := ijwt.GenerateToken(userWithExam, user.UserName)
	if !logger.HandleError(err) {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(internalServerError)
		return
	}

	response := ijwt.LoginResponse{
		JwtToken: token, User: userWithExam,
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
	if err := ctx.ReadJSON(&data); !logger.HandleError(err) {
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
		// User does not use third-party login
		if !user.IsProvider {
			ctx.StatusCode(http.StatusNotAcceptable)
			ctx.JSON(apiResponse{
				"status":  "failed",
				"message": "Please use your username and password to login",
			})
			return
		}
		{
			_provider := &struct {
				Name string
			}{}
			if userHasProvider := database.DB().Table("providers as p").
				Select("up.*, p.name").Joins("LEFT JOIN user_providers as up ON up.provider_id = p.id").
				First(_provider, "up.user_id = ?", user.Id).Error == nil; userHasProvider && _provider.Name != provider.Name {
				ctx.StatusCode(http.StatusNotAcceptable)
				ctx.JSON(apiResponse{
					"status":  "failed",
					"message": fmt.Sprintf("Please use %s login", _provider.Name),
				})
				return
			}
		}

		{
			if user.Device.Id == uuid.Nil {
				if err := deviceExist(ctx, user.Id, data.DeviceId); err != nil {
					ctx.StatusCode(http.StatusForbidden)
					ctx.JSON(deviceExistError)
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
			} else if len(strings.TrimSpace(user.Device.Identifier)) > 0 && (user.Device.Identifier != data.DeviceId ||
				!strings.EqualFold(strings.TrimSpace(strings.ToLower(user.Device.Name)), strings.TrimSpace(strings.ToLower(data.DeviceName)))) {
				var i string
				var err error
				if i, err = creeteDeviceMail(user.Id.String(),
					user.UserName,
					user.Device.Identifier,
					user.Device.Name,
					time.Now().Add(time.Minute*10)); err != nil {
					ctx.StatusCode(http.StatusInternalServerError)
					ctx.JSON(internalServerError)
					return
				}
				queue.Dispatch(queue.Job{
					Type: queue.SendMail,
					Func: func() error {
						return sendmail.SendNewDeviceMail(user, i)
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
		if err := deviceExist(ctx, user.Id, data.DeviceId); err != nil {
			ctx.StatusCode(http.StatusForbidden)
			ctx.JSON(deviceExistError)
			return
		}

		if data.UserName == "" {
			data.UserName = strings.Split(data.Email, "@")[0]
		}
		data.UserName = strings.ToLower(data.UserName)
		if ok := repository.NewRepository(user).FindOne("username = ?", data.UserName); ok {
			username, err := list.UniqueByField(user, data.UserName, "username")
			if !logger.HandleError(err) {
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

		if err := database.DB().Create(&models.UserProvider{
			UserId:     user.Id,
			ProviderId: provider.Id,
			IsLoggedIn: true,
		}).Error; !logger.HandleError(err) {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(internalServerError)
			return
		}
	}

	userExams := []UserExamWithName{}
	database.DB().Table("user_exams as ue").
		Select("ue.session, ue.payment_status, ue.created_at, ue.id, e.name, ue.expires_at").Joins("LEFT JOIN exams as e ON e.id = ue.exam_id").
		Where("user_id = ?", user.Id).
		Scan(&userExams)
	userWithExam := &UserWithExam{*user, userExams, len(userExams) != 0}

	token, err := ijwt.GenerateToken(userWithExam, user.UserName)
	if !logger.HandleError(err) {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(internalServerError)
		return
	}

	response := ijwt.LoginResponse{
		JwtToken: token, User: userWithExam,
	}
	ctx.JSON(apiResponse{
		"status":  "success",
		"message": "Logged in successfully",
		"data":    response,
	})
}

func Logout(ctx iris.Context) {
	user, _ := getUser(ctx)
	cache.Forget(user.UserName + ".access")
	cache.Forget(user.UserName + ".refresh")

	device := &models.Device{}
	if ok := repository.NewRepository(device).FindOne("user_id = ?", user.Id); ok {
		device.Name = ""
		device.Identifier = ""
		database.DB().Save(device)
	}
	ctx.JSON(apiResponse{
		"status":  "success",
		"message": "Logged out successfully",
	})
}
