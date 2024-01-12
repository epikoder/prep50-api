package admin

import (
	"time"

	"github.com/Prep50mobileApp/prep50-api/src/models"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/settings"
	"github.com/Prep50mobileApp/prep50-api/src/services/database"
	"github.com/kataras/iris/v12"
)

type UserController struct{ Ctx iris.Context }

func (c *UserController) Get() {
	session := settings.Get("exam.session", time.Now().Year())
	users := []struct {
		IsSubscribed bool `json:"is_subscribed"`
		models.User
	}{}
	if err := database.DB().Table("users as u").Select(`*,
	CASE WHEN ue.payment_status = 'completed' 
	THEN 1 
	ELSE 0 
	END as is_subscribed`).Joins("LEFT JOIN user_exams as ue ON u.id = ue.user_id").
		Where("ue.session = ?", session).
		Limit(500).
		Find(&users).
		Error; err != nil {
		c.Ctx.StatusCode(500)
		c.Ctx.JSON(internalServerError)
		return
	}
	subscribed := 0
	unsubscribed := 0
	for _, u := range users {
		if u.IsSubscribed {
			subscribed++
		} else {
			unsubscribed++
		}
	}
	c.Ctx.JSON(apiResponse{
		"status": "success",
		"data": apiResponse{
			"subscribed":   subscribed,
			"unsubscribed": unsubscribed,
			"users":        users,
		},
	})
}
