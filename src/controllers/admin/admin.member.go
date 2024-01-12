package admin

import (
	"github.com/Prep50mobileApp/prep50-api/src/models"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/logger"
	"github.com/Prep50mobileApp/prep50-api/src/services/database"
	"github.com/kataras/iris/v12"
)

type MemberController struct{ Ctx iris.Context }

func (c *MemberController) Get() {
	users := []models.User{}
	if err := database.DB().Find(&users, "is_admin = 1").Error; !logger.HandleError(err) {
		c.Ctx.StatusCode(500)
		c.Ctx.JSON(internalServerError)
		return
	}
	c.Ctx.JSON(apiResponse{
		"status": "success",
		"data":   users,
	})
}

func (c *MemberController) Post() {
	users := []models.User{}
	c.Ctx.JSON(apiResponse{
		"status": "success",
		"data":   users,
	})
}
