package controllers

import (
	"github.com/Prep50mobileApp/prep50-api/src/models"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/logger"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/repository"
	"github.com/kataras/iris/v12"
)

type FaqController struct {
	Ctx iris.Context
}

func (c *FaqController) Get() {
	faqs := []models.Faq{}
	if err := repository.NewRepository(&models.Faq{}).FindMany(&faqs); !logger.HandleError(err) {
		c.Ctx.StatusCode(500)
		c.Ctx.JSON(internalServerError)
		return
	}
	c.Ctx.JSON(apiResponse{
		"status": "success",
		"data":   faqs,
	})
}
