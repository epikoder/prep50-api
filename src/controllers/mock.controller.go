package controllers

import (
	"fmt"
	"os"
	"reflect"
	"time"

	"github.com/Prep50mobileApp/prep50-api/src/models"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/cache"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/config"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/logger"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/validation"
	"github.com/Prep50mobileApp/prep50-api/src/services/database"
	"github.com/google/uuid"
	"github.com/kataras/iris/v12"
	"gorm.io/gorm"
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
	if notFound := database.DB().Table("mocks as m").
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
		First(&mock).Error != nil; notFound {
		c.Ctx.JSON(apiResponse{
			"status": "success",
			"data":   nil,
		})
		return
	}
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
	tx := database.DB().Table("mocks as m").
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
	err = database.DB().Find(&ques, "id IN ? AND subject_id IN ? ORDER BY RAND()", ids, data.Subject).Error
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

func (c *MockController) Put() {
	var userAnswer = struct {
		Id     uuid.UUID
		Answer Answer
	}{}
	if err := c.Ctx.ReadJSON(&userAnswer); !logger.HandleError(err) {
		c.Ctx.StatusCode(400)
		c.Ctx.JSON(validation.Errors(err))
		return
	}

	answers := Answer{}
	mock := &models.Mock{}

	if err := database.DB().Table(mock.Tag()).Preload("Questions", func(db *gorm.DB) *gorm.DB {
		return db.Table(fmt.Sprintf("%s.questions", config.Conf.Database.Name))
	}).First(&mock, "id = ?", userAnswer.Id).Error; err != nil {
		c.Ctx.JSON(apiResponse{
			"status":  "failed",
			"message": "Mock not found",
		})
		return
	}

	if env := os.Getenv("APP_ENV"); env != "" && env != "production" {
		goto SKIP
	}
	if mock.StartTime.Before(time.Now()) || mock.EndTime.After(time.Now()) {
		c.Ctx.JSON(apiResponse{
			"status":  "failed",
			"message": "Mock is not active",
		})
		return
	}

SKIP:
	user, _ := getUser(c.Ctx)
	_time := time.Now()
	key := fmt.Sprintf("mock.%s.answer", mock.Id)
	ans, ok := cache.Get(key)
	if err := answers.UnmarshalBinary([]byte(ans)); !ok || !logger.HandleError(err) {

		for _, question := range mock.Questions {
			if question.QuestionTypeId == uint(models.OBJECTIVE) {
				answers[question.Id] = question.ShortAnswer
			} else {
				answers[question.Id] = question.FullAnswer
			}
		}
		logger.HandleError(cache.Set(key, answers, cache.Duration(mock.EndTime.Unix())))
	}

	v := reflect.ValueOf(*mock)
	{
		t := v.Type()
		sf := []reflect.StructField{}
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			if f.Name == "Questions" {
				f.Tag = `json:"-"`
			}
			sf = append(sf, f)
		}
		s := reflect.StructOf(sf)
		if t.ConvertibleTo(s) {
			v = v.Convert(s)
		}
	}
	{
		result := &models.MockResult{}
		if err := database.DB().First(result, "user_id = ? AND mock_id = ?", user.Id, mock.Id).Error; err == nil {
			c.Ctx.JSON(apiResponse{
				"Status": "success",
				"data": apiResponse{
					"score":  result.Score,
					"mock":   v.Interface(),
					"answer": answers,
				},
			})
			return
		}
	}

	var score uint = 0
	for id, ans := range answers {
		ua, ok := userAnswer.Answer[id]
		if !ok {
			continue
		}
		if ua == ans {
			score++
		}
	}

	if err := database.DB().Create(models.MockResult{
		MockId:   mock.Id,
		UserId:   user.Id,
		Score:    score,
		Duration: uint(mock.StartTime.Sub(_time).Minutes()),
	}).Error; !logger.HandleError(err) {
		c.Ctx.StatusCode(500)
		c.Ctx.JSON(internalServerError)
		return
	}

	c.Ctx.JSON(apiResponse{
		"Status": "success",
		"data": apiResponse{
			"score":   score,
			"mock":    v.Interface(),
			"answers": answers,
		},
	})
}

func GetMockVideo(ctx iris.Context) {
	ctx.JSON(apiResponse{
		"status": "success",
		"data":   "http://prep50.com/video/mock.mp4",
	})
}
