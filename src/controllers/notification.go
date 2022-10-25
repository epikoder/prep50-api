package controllers

import (
	"net/http"
	"os"
	"time"

	"firebase.google.com/go/messaging"
	"github.com/Prep50mobileApp/prep50-api/src/models"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/notifications"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/validation"
	"github.com/Prep50mobileApp/prep50-api/src/services/database"
	"github.com/Prep50mobileApp/prep50-api/src/services/queue"
	"github.com/google/uuid"
	"github.com/kataras/iris/v12"
)

type (
	NotificationController struct {
		Ctx iris.Context
	}
)

func (c *NotificationController) Post() {
	data := struct {
		Token string `validate:"required"`
	}{}
	if err := c.Ctx.ReadJSON(&data); err != nil {
		c.Ctx.StatusCode(400)
		c.Ctx.JSON(validation.Errors(err))
		return
	}
	user, _ := getUser(c.Ctx)
	database.UseDB("app").Preload("Fcm").First(user)
	if user.Fcm.Id == uuid.Nil {
		user.Fcm.Id = uuid.New()
		user.Fcm.UserId = user.Id
	}

	user.Fcm.Token = data.Token
	user.Fcm.Timestamp = time.Now()

	if err := user.Database().Save(user.Fcm).Error; err != nil {
		c.Ctx.StatusCode(500)
		c.Ctx.JSON(internalServerError)
		return
	}
	c.Ctx.JSON(apiResponse{
		"status": "success",
	})
}

func (c *NotificationController) Get() {
	notifications := []struct {
		models.Notification
		User bool `json:"user"`
	}{}
	user, _ := getUser(c.Ctx)
	database.UseDB("app").Table("notifications as n").Select(`n.*,
	CASE 
		WHEN user_id = ? 
		THEN 1
		ELSE 0
	END as user`, user.Id).Find(&notifications)

	c.Ctx.JSON(apiResponse{
		"status": "success",
		"data":   notifications,
	})
}

/*****************************/
/*********TEST ONLY***********/
func (c *NotificationController) Put(data models.Notification) {
	if env := os.Getenv("APP_ENV"); env == "" || env == "production" {
		c.Ctx.StatusCode(http.StatusForbidden)
		return
	}

	database.UseDB("app").Create(&models.Notification{
		Id:       uuid.New(),
		Title:    data.Title,
		Body:     data.Body,
		ImageUrl: data.ImageUrl,
	})

	queue.Dispatch(queue.Job{
		Type: queue.Action,
		Func: func() error {
			return notifications.SendNotifications(&messaging.Notification{
				Title:    data.Title,
				Body:     data.Body,
				ImageURL: data.ImageUrl,
			})
		},
	})
	c.Ctx.JSON(apiResponse{
		"status":  "success",
		"message": "Notification is been sent",
	})
}

/*****************************/
