package admin

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Prep50mobileApp/prep50-api/config"
	"github.com/Prep50mobileApp/prep50-api/src/models"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/list"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/logger"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/repository"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/settings"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/validation"
	"github.com/google/uuid"
	"github.com/kataras/iris/v12"
	"gorm.io/gorm"
)

func GetCurrentWeekQuiz(ctx iris.Context) {
	quiz := &models.WeeklyQuiz{}
	year, week := time.Now().ISOWeek()
	if ok := repository.NewRepository(quiz).Preload("Questions", func(db *gorm.DB) *gorm.DB {
		return db.Table(fmt.Sprintf("%s.questions", config.Conf.Database.Core.Name))
	}).
		FindOne("week = ? AND session = ?", week, settings.Get("exam.session", year)); !ok {
		ctx.StatusCode(404)
		ctx.JSON(internalServerError)
		return
	}
	questions := quiz.Questions
	ctx.JSON(apiResponse{
		"status": "success",
		"data": apiResponse{
			"quiz":      quiz,
			"questions": questions,
		},
	})
}

func IndexWeeklyQuiz(ctx iris.Context) {
	weeklyQuizzes := []models.WeeklyQuiz{}
	if err := repository.NewRepository(&models.WeeklyQuiz{}).FindMany(&weeklyQuizzes); !logger.HandleError(err) {
		ctx.StatusCode(500)
		ctx.JSON(internalServerError)
		return
	}
	ctx.JSON(apiResponse{
		"status": "success",
		"data":   weeklyQuizzes,
	})
}

func ViewWeeklyQuizQuestions(ctx iris.Context) {
	quizz := &models.WeeklyQuiz{}
	if ok := repository.NewRepository(quizz).Preload("Questions", func(db *gorm.DB) *gorm.DB {
		return db.Table(fmt.Sprintf("%s.questions", config.Conf.Database.Core.Name))
	}).
		FindOne("id = ?", ctx.URLParam("id")); !ok {
		ctx.JSON(apiResponse{
			"status":  "failed",
			"message": "quizz not found",
		})
		return
	}
	questions := quizz.Questions
	ctx.JSON(apiResponse{
		"status": "success",
		"data": apiResponse{
			"quizz":     quizz,
			"questions": questions,
		},
	})
}

func CreateWeeklyQuiz(ctx iris.Context) {
	data := &models.WeeklyQuizFormStruct{}
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
	user, _ := getUser(ctx)
	_, week := data.Start_Time.ISOWeek()
	session := settings.Get("exam.session", time.Now().Year()).(int)

	if repository.NewRepository(&models.WeeklyQuiz{}).FindOne("week = ? AND session = ?", week, session) {
		ctx.JSON(apiResponse{
			"status":  "failed",
			"message": "Quizz already exist for current week",
		})
		return
	}
	w := &models.WeeklyQuiz{
		Id:        uuid.New(),
		Week:      uint(week),
		Prize:     data.Prize,
		Message:   data.Message,
		Session:   uint(session),
		Duration:  data.Duration,
		StartTime: data.Start_Time,
		CreatedBy: user.Id.String(),
	}
	if err := repository.NewRepository(w).Create(); !logger.HandleError(err) {
		ctx.JSON(internalServerError)
		ctx.StatusCode(http.StatusInternalServerError)
		return
	}
	ctx.JSON(apiResponse{
		"status":  "success",
		"message": "weekly quizz created successfully",
	})
}

func UpdateWeeklyQuizQuestion(ctx iris.Context) {
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

	quiz := &models.WeeklyQuiz{}
	if ok := repository.NewRepository(quiz).
		FindOne("id = ?", data.Id); !ok {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(apiResponse{
			"status":  "failed",
			"message": "quiz not found",
		})
		return
	}
	quizQues, err := quiz.WeeklyQuestions()
	if !logger.HandleError(err) {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(internalServerError)
		return
	}
	tmpQuizQues := quizQues
	for i, q := range tmpQuizQues {
		if list.Contains(data.Remove, q.QuestionId) && !list.Contains(data.Add, q.QuestionId) {
			if err := quiz.Database().Delete(q, "quiz_id = ? AND question_id = ?", quiz.Id, q.QuestionId).Error; !logger.HandleError(err) {
				ctx.StatusCode(500)
				ctx.JSON(internalServerError)
			}
			quizQues = append(append(make([]models.WeeklyQuestion, 0), quizQues[:i]...), quizQues[i+1:]...)
		}

		tmp := data.Add
		for index, id := range tmp {
			if id == q.QuestionId {
				data.Add = append(append(make([]uint, 0), data.Add[:index]...), data.Add[index+1:]...)
			}
		}
	}

	user, _ := getUser(ctx)
	quizQues = []models.WeeklyQuestion{}
	for _, id := range data.Add {
		quizQues = append(quizQues, models.WeeklyQuestion{
			QuizId:     quiz.Id,
			QuestionId: id,
			CreatedBy:  user.Id.String(),
		})
	}

	if len(data.Add) > 0 {
		if err := quiz.Database().Save(quizQues).Error; !logger.HandleError(err) {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(internalServerError)
			return
		}
	}
	ctx.JSON(apiResponse{
		"status":  "success",
		"message": "Quizz questions updated successfully",
	})
}

func UpdateWeeklyQuiz(ctx iris.Context) {
	data := &models.WeeklyQuizUpdateForm{}
	ctx.ReadJSON(data)
	quiz := &models.WeeklyQuiz{}
	if ok := repository.NewRepository(quiz).FindOne("id = ?", data.Id); !ok {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(apiResponse{
			"status":  "failed",
			"message": "quizz not found",
		})
		return
	}
	if quiz.StartTime.Before(time.Now()) {
		ctx.JSON(apiResponse{
			"status":  "failed",
			"message": "quizz can not longer be changed",
		})
		return
	}

	if data.Duration != 0 {
		quiz.Duration = data.Duration
	}
	if data.Message != "" {
		quiz.Message = data.Message
	}
	if data.Prize != 0 {
		quiz.Prize = data.Prize
	}
	if !strings.Contains(data.Start_Time.String(), "0001-01-01") {
		quiz.StartTime = data.Start_Time
		_, week := data.Start_Time.ISOWeek()
		quiz.Week = uint(week)
	}
	if err := repository.NewRepository(quiz).Save(); !logger.HandleError(err) {
		ctx.StatusCode(500)
		ctx.JSON(internalServerError)
		return
	}
	ctx.JSON(apiResponse{
		"status":  "success",
		"message": "Updated successfully",
	})
}

func DeleteWeeklyQuizz(ctx iris.Context) {
	quiz := &models.WeeklyQuiz{}
	if ok := repository.NewRepository(quiz).FindOne("id = ?", ctx.URLParam("id")); !ok {
		ctx.JSON(apiResponse{
			"status":  "failed",
			"message": "not found",
		})
		return
	}

	if err := repository.NewRepository(&models.WeeklyQuiz{}).
		Delete(quiz); !logger.HandleError(err) {
		ctx.JSON(apiResponse{
			"status":  "failed",
			"message": "delete failed",
		})
		return
	}
	ctx.JSON(apiResponse{
		"status":  "success",
		"message": "quizz deleted successfully",
	})
}
