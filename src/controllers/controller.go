package controllers

import (
	"fmt"

	"github.com/Prep50mobileApp/prep50-api/src/models"
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
