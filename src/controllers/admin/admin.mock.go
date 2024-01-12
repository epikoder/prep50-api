package admin

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Prep50mobileApp/prep50-api/src/models"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/config"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/list"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/logger"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/repository"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/settings"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/validation"
	"github.com/google/uuid"
	"github.com/kataras/iris/v12"
	"gorm.io/gorm"
)

func IndexMock(ctx iris.Context) {
	mock := []models.Mock{}
	if err := repository.NewRepository(&models.Mock{}).FindMany(&mock); !logger.HandleError(err) {
		ctx.StatusCode(500)
		ctx.JSON(internalServerError)
		return
	}
	ctx.JSON(apiResponse{
		"status": "success",
		"data":   mock,
	})
}

func ViewMockQuestions(ctx iris.Context) {
	mock := &models.Mock{}
	session := settings.Get("exam.session", time.Now().Year())
	repository.NewRepository(mock).Preload("Questions", func(db *gorm.DB) *gorm.DB {
		return db.Table(fmt.Sprintf("%s.questions", config.Conf.Database.Name))
	}).FindOne("session = ? AND end_time > ?", session, time.Now())
	ctx.JSON(apiResponse{
		"status": "success",
		"data":   mock,
	})
}

func CreateMock(ctx iris.Context) {
	data := &models.MockForm{}
	if err := ctx.ReadJSON(data); !logger.HandleError(err) {
		res := validation.Errors(err).(map[string]interface{})
		ctx.StatusCode(http.StatusBadRequest)
		if strings.Contains(data.Start_Time.String(), "0001-01-01") {
			e, ok := res["error"]
			if !ok {
				e = map[string]interface{}{}
			}
			i := e.(map[string]interface{})
			i["start_time"] = "time is invalid"
			res["error"] = i
		}
		ctx.JSON(res)
		return
	}

	if time.Now().After(data.Start_Time) || time.Now().After(data.End_Time) {
		ctx.JSON(apiResponse{
			"status":  "failed",
			"message": "Mock start/end time can not be in past",
		})
		return
	}

	user, _ := getUser(ctx)
	mock := &models.Mock{
		Id:        uuid.New(),
		StartTime: data.Start_Time,
		EndTime:   data.End_Time,
		Amount:    data.Amount,
		Duration:  data.Duration,
		Session:   uint(settings.Get("exam.session", time.Now().Year()).(int)),
		CreatedBy: user.Id.String(),
	}
	if err := repository.NewRepository(mock).Create(); !logger.HandleError(err) {
		ctx.JSON(internalServerError)
		ctx.StatusCode(http.StatusInternalServerError)
		return
	}
	ctx.JSON(apiResponse{
		"status":  "success",
		"message": "Mock created successfully",
	})
}

func UpdateMockQuestion(ctx iris.Context) {
	type updateQuizForm struct {
		Id     uuid.UUID
		Add    []uint
		Remove []uint
	}
	data := &updateQuizForm{}
	if err := ctx.ReadJSON(data); !logger.HandleError(err) {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(validation.Errors(err))
		return
	}
	if len(data.Add) > 0 {
		data.Add = list.Unique(data.Add).([]uint)
	}
	if len(data.Remove) > 0 {
		data.Remove = list.Unique(data.Remove).([]uint)
	}

	mock := &models.Mock{}
	if ok := repository.NewRepository(mock).
		FindOne("id = ?", data.Id); !ok {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(apiResponse{
			"status":  "failed",
			"message": "mock not found",
		})
		return
	}
	mockQues, err := mock.MockQuestions()
	if !logger.HandleError(err) {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(internalServerError)
		return
	}

	tmpMockQues := mockQues
	for i, q := range tmpMockQues {
		if list.Contains(data.Remove, q.QuestionId) && !list.Contains(data.Add, q.QuestionId) {
			if err := mock.Database().Delete(q, "mock_id = ? AND question_id = ?", mock.Id, q.QuestionId).Error; !logger.HandleError(err) {
				ctx.StatusCode(500)
				ctx.JSON(internalServerError)
			}
			mockQues = append(append(make([]models.MockQuestion, 0), mockQues[:i]...), mockQues[i+1:]...)
		}

		tmp := data.Add
		for index, id := range tmp {
			if id == q.QuestionId {
				data.Add = append(append(make([]uint, 0), data.Add[:index]...), data.Add[index+1:]...)
			}
		}
	}

	user, _ := getUser(ctx)
	mockQues = []models.MockQuestion{}
	for _, id := range data.Add {
		mockQues = append(mockQues, models.MockQuestion{
			Id:         uuid.New(),
			MockId:     mock.Id,
			QuestionId: id,
			CreatedBy:  user.Id.String(),
		})
	}

	if len(data.Add) > 0 {
		if err := mock.Database().Save(mockQues).Error; !logger.HandleError(err) {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(internalServerError)
			return
		}
	}
	ctx.JSON(apiResponse{
		"status":  "success",
		"message": "Mock questions updated successfully",
	})
}

func UpdateMock(ctx iris.Context) {
	data := &models.MockUpdateForm{}
	ctx.ReadJSON(data)
	mock := &models.Mock{}
	if ok := repository.NewRepository(mock).FindOne("id = ?", data.Id); !ok {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(apiResponse{
			"status":  "failed",
			"message": "Mock not found",
		})
		return
	}
	if mock.StartTime.Before(time.Now()) {
		ctx.JSON(apiResponse{
			"status":  "failed",
			"message": "Mock can not longer be changed",
		})
		return
	}

	if data.Duration != 0 {
		mock.Duration = data.Duration
	}
	if data.Amount != "" {
		if amount, err := strconv.Atoi(data.Amount); err == nil {
			mock.Amount = uint(amount)
		}
	}

	if !strings.Contains(data.Start_Time.String(), "0001-01-01") {
		mock.StartTime = data.Start_Time
	}
	if !strings.Contains(data.End_Time.String(), "0001-01-01") {
		mock.EndTime = data.End_Time
	}
	if err := repository.NewRepository(mock).Save(); !logger.HandleError(err) {
		ctx.StatusCode(500)
		ctx.JSON(internalServerError)
		return
	}
	ctx.JSON(apiResponse{
		"status":  "success",
		"message": "Updated successfully",
	})
}

func DeleteMock(ctx iris.Context) {
	mock := &models.Mock{}
	if ok := repository.NewRepository(mock).FindOne("id = ?", ctx.URLParam("id")); !ok {
		ctx.JSON(apiResponse{
			"status":  "failed",
			"message": "not found",
		})
		return
	}

	if err := repository.NewRepository(&models.Mock{}).
		Delete(mock); !logger.HandleError(err) {
		ctx.JSON(apiResponse{
			"status":  "failed",
			"message": "delete failed",
		})
		return
	}
	ctx.JSON(apiResponse{
		"status":  "success",
		"message": "Mock deleted successfully",
	})
}
