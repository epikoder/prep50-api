package controllers

import (
	"fmt"
	"os"
	"time"

	"github.com/Prep50mobileApp/prep50-api/src/models"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/validation"
	"github.com/Prep50mobileApp/prep50-api/src/services/database"
	"github.com/kataras/iris/v12"
)

type MockController struct {
	Ctx iris.Context
}

func (c *MockController) Get() {
	user, _ := getUser(c.Ctx)
	mock := struct {
		models.Mock
		Available  bool `json:"available"`
		Registered bool `json:"registered"`
	}{}
	database.UseDB("app").Table("mocks as m").
		Select(`m.*, 
		CASE 
			WHEN um.user_id = ? THEN 1 
			ELSE 0 
		END as registered,
		CASE 
			WHEN m.end_time > ? THEN 1 
			ELSE 0 
		END as available`, user.Id, time.Now()).
		Joins("LEFT JOIN user_mocks as um ON m.id = um.mock_id").
		Order("start_time DESC").
		First(&mock)
	c.Ctx.JSON(apiResponse{
		"status": "success",
		"data":   mock,
	})
}

func (c *MockController) Post() {
	data := struct {
		Id      string `validate:"required"`
		Subject []int  `validate:"required,max=4"`
	}{}

	if err := c.Ctx.ReadJSON(&data); err != nil {
		c.Ctx.StatusCode(400)
		c.Ctx.JSON(validation.Errors(err))
		return
	}

	user, _ := getUser(c.Ctx)
	mock := struct {
		models.Mock
		Registered bool `json:"registered"`
	}{}
	tx := database.UseDB("app").Table("mocks as m").
		Select(`m.*, 
	CASE 
		WHEN um.user_id = ? THEN 1 
		ELSE 0 
	END as registered`, user.Id).
		Joins("LEFT JOIN user_mocks as um ON m.id = um.mock_id")

	production := func() bool {
		env := os.Getenv("APP_ENV")
		return env == "" || env == "production"
	}()

	if production {
		tx.Where("m.end_time > ?", time.Now())
	}
	if err := tx.First(&mock, "m.id = ? AND um.user_id = ?", data.Id, user.Id).Error; err != nil {
		c.Ctx.JSON(apiResponse{
			"status":  "failed",
			"message": "Mock not found, not registered or ended",
		})
		return
	}

	if mock.StartTime.After(time.Now()) && production {
		c.Ctx.JSON(apiResponse{
			"status": "success",
			"data": apiResponse{
				"mock":      mock,
				"message":   fmt.Sprintf("Mock starts in %v", mock.StartTime),
				"questions": []interface{}{},
			},
		})
		return
	}

	mq, err := mock.MockQuestions()
	if err != nil {
		c.Ctx.StatusCode(500)
		c.Ctx.JSON(internalServerError)
		return
	}
	ids := []uint{}
	for _, q := range mq {
		ids = append(ids, q.QuestionId)
	}
	ques := []models.Question{}
	err = database.UseDB("core").Find(&ques, "id IN ? AND subject_id IN ? ORDER BY RAND()", ids, data.Subject).Error
	if err != nil {
		c.Ctx.StatusCode(500)
		c.Ctx.JSON(internalServerError)
		return
	}

	c.Ctx.JSON(apiResponse{
		"status": "success",
		"data": apiResponse{
			"mock":      mock,
			"questions": ques,
		},
	})
}

func GetMockVideo(ctx iris.Context) {
	ctx.JSON(apiResponse{
		"status": "success",
		"data":   "http://prep50.com/video/mock.mp4",
	})
}
