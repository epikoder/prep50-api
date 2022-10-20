package controllers

import (
	"fmt"
	"time"

	"github.com/Prep50mobileApp/prep50-api/src/models"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/logger"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/repository"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/settings"
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

	queryWithPayment struct {
		Id            uuid.UUID
		ExamId        uuid.UUID
		Name          string
		Session       int
		Status        bool
		PaymentStatus models.PaymentStatus
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

//+++++++++++++++++++++++++++++++++++++++++++++++++
func GetMocks(ctx iris.Context) {
	mocks := []models.Mock{}
	if err := repository.NewRepository(&models.Mock{}).
		FindMany(&mocks, "session = ? AND start_time < ?", settings.Get("examSession", time.Now().Year()), time.Now()); !logger.HandleError(err) {
		ctx.StatusCode(500)
		ctx.JSON(internalServerError)
		return
	}
	ctx.JSON(mocks)
}

//+++++++++++++++++++++++++++++++++++++++++++++++++
