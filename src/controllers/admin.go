package controllers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/Prep50mobileApp/prep50-api/src/models"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/hash"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/ijwt"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/list"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/logger"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/repository"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/validation"
	"github.com/google/uuid"
	"github.com/kataras/iris/v12"
)

func AdminLogin(ctx iris.Context) {
	type (
		adminUser struct {
			models.User
			Permissions []string
		}
	)

	data := models.UserLoginFormStruct{}
	if err := ctx.ReadJSON(&data); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(validation.Errors(err))
		return
	}

	var user = &models.User{}
	{
		if ok := repository.NewRepository(user).
			Preload("Permissions").
			FindOne("username = ? OR email = ?", data.UserName, data.UserName); !ok {
			ctx.StatusCode(http.StatusUnauthorized)
			ctx.JSON(apiResponse{
				"status":  "failed",
				"message": "account not found",
			})
			return
		}
		if ok := hash.CheckHash(user.Password, data.Password); !ok {
			ctx.StatusCode(http.StatusUnauthorized)
			ctx.JSON(apiResponse{
				"status":  "failed",
				"message": "invalid username or password",
			})
			return
		}

		if !user.HasRole("admin") {
			ctx.StatusCode(http.StatusUnauthorized)
			ctx.JSON(apiResponse{
				"status":  "failed",
				"message": "unauthorized access",
			})
			return
		}
	}

	permissions := []string{}
	for _, p := range user.Permissions {
		permissions = append(permissions, p.Name)
	}
	token, err := ijwt.GenerateToken(&adminUser{*user, permissions}, user.UserName)
	if !logger.HandleError(err) {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(internalServerError)
		return
	}

	ctx.JSON(apiResponse{
		"status":  "success",
		"message": "logged in successfully",
		"data":    token,
	})
}

func CreateWeeklyQuiz(ctx iris.Context) {
	data := &models.WeeklyQuizFormStruct{}
	if err := ctx.ReadJSON(data); err != nil {
		res := validation.Errors(err).(map[string]interface{})
		ctx.StatusCode(http.StatusBadRequest)
		if strings.Contains(data.Start_Time.String(), "0001-01-01") {
			e, ok := res["error"]
			fmt.Println(e)
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
	w := &models.WeeklyQuiz{
		Id:            uuid.New(),
		Prize:         data.Prize,
		Message:       data.Message,
		QuestionCount: data.Question_Count,
		Duration:      data.Duration,
		StartTime:     data.Start_Time,
		CreatedBy:     user.Id.String(),
	}
	if err := repository.NewRepository(w).Create(); err != nil {
		ctx.JSON(internalServerError)
		ctx.StatusCode(http.StatusInternalServerError)
		return
	}
	ctx.JSON(w)
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
	if err := repository.NewRepository(&models.WeeklyQuestion{}).
		FindMany(quiz, "id = ?", data.Id); !logger.HandleError(err) {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(internalServerError)
		return
	}
	quizQues, err := quiz.WeeklyQuestions()
	if !logger.HandleError(err) {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(internalServerError)
		return
	}
	for i, q := range quizQues {
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
	if v := len(quizQues) + len(data.Add); v > int(quiz.QuestionCount) {
		ctx.StatusCode(http.StatusForbidden)
		ctx.JSON(apiResponse{
			"status":  "failed",
			"message": fmt.Sprintf("maximum question for quiz exceeded, %d given", v),
		})
		return
	}
	fmt.Println(data.Add)
	fmt.Println(quizQues)
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
}

func UpdateWeeklyQuiz(ctx iris.Context) {
	type updateQuizForm struct {
		Id     uuid.UUID
		Add    []uint
		Remove []int
	}
}
