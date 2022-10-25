package admin

import (
	"firebase.google.com/go/messaging"
	"github.com/Prep50mobileApp/prep50-api/src/models"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/notifications"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/validation"
	"github.com/Prep50mobileApp/prep50-api/src/services/database"
	"github.com/Prep50mobileApp/prep50-api/src/services/queue"
	"github.com/google/uuid"
	"github.com/kataras/iris/v12"
)

type AdminNotificationController struct {
	Ctx iris.Context
}

func (c *AdminNotificationController) Post() {
	data := models.Notification{}
	if err := c.Ctx.ReadJSON(&data); err != nil {
		c.Ctx.StatusCode(400)
		c.Ctx.JSON(validation.Errors(err))
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

func RegisterClient(s string) {

}
