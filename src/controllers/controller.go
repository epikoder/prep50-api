package controllers

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/Prep50mobileApp/prep50-api/src/models"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/cache"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/logger"
	"github.com/Prep50mobileApp/prep50-api/src/services/database"
	"github.com/google/uuid"
	"github.com/kataras/iris/v12"
)

type (
	query struct {
		Id      uuid.UUID
		UserId  uuid.UUID
		ExamId  uuid.UUID
		Name    string
		Session int
		Status  bool
	}
)

var (
	internalServerError = apiResponse{
		"status":  "failed",
		"message": "error occcured",
	}

	getUser = func(ctx iris.Context) (u *models.User, err error) {
		i, err := ctx.User().GetRaw()
		if !logger.HandleError(err) {
			return nil, err
		}
		var ok bool
		if u, ok = i.(*models.User); !ok {
			return nil, fmt.Errorf("user is nil")
		}
		return u, nil
	}
)

func DeregisterDevice(ctx iris.Context) {
	mp := struct {
		Token string `validate:"required"`
	}{}
	if err := ctx.ReadQuery(&mp); !logger.HandleError(err) {
		ctx.View("deregister_device", iris.Map{
			"message": "Malformed token",
			"status":  false,
		})
		return
	}

	if err := validateMailToken(mp.Token); !logger.HandleError(err) {
		ctx.View("deregister_device", iris.Map{
			"message": err.Error(),
			"status":  false,
		})
		return
	}
	ctx.View("deregister_device", iris.Map{
		"message": "Device deregistered successfully",
		"status":  true,
	})
}

func validateMailToken(k string) error {
	token, ok := cache.Get(k)
	if !ok {
		return fmt.Errorf("malformed token")
	}
	m := &NewDeviceMail{}
	if err := json.Unmarshal([]byte(token), m); err != nil {
		return fmt.Errorf("malformed token")
	}

	if time.Now().After(m.Expires) {
		return fmt.Errorf("token expired")
	}
	if err := database.UseDB("app").
		Raw("DELETE from devices WHERE identifier = ? AND user_id = ?", m.DeviceId, m.UserId).
		Error; !logger.HandleError(err) {
		return fmt.Errorf("something went wrong")
	}
	cache.Forget(m.Username + ".access")
	cache.Forget(m.Username + ".refresh")
	return nil
}

//+++++++++++++++++++++++++++++++++++++++++++++++++

func Terms(ctx iris.Context) {
	st := &models.GeneralSetting{}
	database.UseDB("app").First(st)
	ctx.JSON(apiResponse{
		"status": "success",
		"data":   st.Terms,
	})
}
func Privacy(ctx iris.Context) {
	st := &models.GeneralSetting{}
	database.UseDB("app").First(st)
	ctx.JSON(apiResponse{
		"status": "success",
		"data":   st.Privacy,
	})
}
