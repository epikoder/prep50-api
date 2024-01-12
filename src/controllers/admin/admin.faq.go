package admin

import (
	"github.com/Prep50mobileApp/prep50-api/src/models"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/list"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/validation"
	"github.com/Prep50mobileApp/prep50-api/src/services/database"
	"github.com/google/uuid"
	"github.com/kataras/iris/v12"
)

type AdminFaqController struct {
	Ctx iris.Context
}

func (c *AdminFaqController) Get() {
	faqs := []models.Faq{}
	database.DB().Find(&faqs)
	c.Ctx.JSON(apiResponse{
		"status": "success",
		"data":   faqs,
	})
}

func (c *AdminFaqController) Post() {
	data := models.Faq{}
	if err := c.Ctx.ReadJSON(&data); err != nil {
		c.Ctx.StatusCode(400)
		c.Ctx.JSON(validation.Errors(err))
		return
	}

	slug, err := list.UniqueSlug(&models.Faq{}, list.Slug(data.Title))
	if err != nil {
		c.Ctx.JSON(apiResponse{
			"status":  "failed",
			"message": "Could not generate slug",
		})
		return
	}
	database.DB().Create(&models.Faq{
		Id:      uuid.New(),
		Title:   data.Title,
		Slug:    slug,
		Content: data.Content,
	})
	c.Ctx.JSON(apiResponse{
		"status":  "success",
		"message": "Faq created successfully",
	})
}

func (c *AdminFaqController) Put() {}
