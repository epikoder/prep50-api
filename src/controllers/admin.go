package controllers

import (
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/Prep50mobileApp/prep50-api/src/models"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/hash"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/ijwt"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/list"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/logger"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/repository"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/settings"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/validation"
	"github.com/Prep50mobileApp/prep50-api/src/services/database"
	"github.com/google/uuid"
	"github.com/kataras/iris/v12"
)

//+++++++++++++++++++++++++++++++++++++++++++++++++++++++++//
func AdminLogin(ctx iris.Context) {
	type (
		adminUser struct {
			models.User
			Permissions []string `json:"permisions"`
			Roles       []string `json:"roles"`
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
	roles := []string{}
	for _, r := range user.Roles {
		roles = append(roles, r.Name)
	}
	token, err := ijwt.GenerateToken(&adminUser{*user, permissions, roles}, user.UserName)
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

//+++++++++++++++++++++++++++++++++++++++++++++++++++++++++//

//+++++++++++++++++++++++++++++++++++++++++++++++++++++++++//
func GetCurrentWeekQuiz(ctx iris.Context) {
	quiz := &models.WeeklyQuiz{}
	year, week := time.Now().ISOWeek()
	if ok := repository.NewRepository(quiz).
		FindOne("week = ? AND session = ?", week, settings.Get("examSession", year)); !ok {
		ctx.StatusCode(404)
		ctx.JSON(internalServerError)
		return
	}
	questions, err := quiz.Questions()
	if err != nil {
		ctx.StatusCode(500)
		ctx.JSON(internalServerError)
		return
	}
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
	if err := repository.NewRepository(&models.WeeklyQuiz{}).FindMany(&weeklyQuizzes); err != nil {
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
	if ok := repository.NewRepository(quizz).
		FindOne("id = ?", ctx.URLParam("id")); !ok {
		ctx.JSON(apiResponse{
			"status":  "failed",
			"message": "quizz not found",
		})
		return
	}
	ques, err := quizz.Questions()
	if err != nil {
		ctx.StatusCode(500)
		ctx.JSON(internalServerError)
		return
	}
	ctx.JSON(apiResponse{
		"status": "success",
		"data": apiResponse{
			"quizz":     quizz,
			"questions": ques,
		},
	})
}

func CreateWeeklyQuiz(ctx iris.Context) {
	data := &models.WeeklyQuizFormStruct{}
	if err := ctx.ReadJSON(data); err != nil {
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
	session := settings.Get("examSession", time.Now().Year()).(int)

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
	if err := repository.NewRepository(w).Create(); err != nil {
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
	if err := repository.NewRepository(quiz).Save(); err != nil {
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
		Delete(quiz); err != nil {
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

//+++++++++++++++++++++++++++++++++++++++++++++++++++++++++//

//+++++++++++++++++++++++++++++++++++++++++++++++++++++++++//
func IndexMock(ctx iris.Context) {
	mock := []models.Mock{}
	if err := repository.NewRepository(&models.Mock{}).FindMany(&mock); err != nil {
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
	if ok := repository.NewRepository(mock).
		FindOne("id = ?", ctx.URLParam("id")); !ok {
		ctx.JSON(apiResponse{
			"status":  "failed",
			"message": "Mock not found",
		})
		return
	}
	ques, err := mock.Questions()
	if err != nil {
		ctx.StatusCode(500)
		ctx.JSON(internalServerError)
		return
	}
	ctx.JSON(apiResponse{
		"status": "success",
		"data": apiResponse{
			"mock":      mock,
			"questions": ques,
		},
	})
}

func CreateMock(ctx iris.Context) {
	data := &models.MockForm{}
	if err := ctx.ReadJSON(data); err != nil {
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
	mock := &models.Mock{
		Id:        uuid.New(),
		StartTime: data.Start_Time,
		EndTime:   data.End_Time,
		Amount:    data.Amount,
		Duration:  data.Duration,
		Session:   uint(settings.Get("examSession", time.Now().Year()).(int)),
		CreatedBy: user.Id.String(),
	}
	if err := repository.NewRepository(mock).Create(); err != nil {
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
	for i, q := range mockQues {
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
	if err := repository.NewRepository(mock).Save(); err != nil {
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
		Delete(mock); err != nil {
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

//+++++++++++++++++++++++++++++++++++++++++++++++++++++++++//

//+++++++++++++++++++++++++++++++++++++++++++++++++++++++++//
func IndexPodcast(ctx iris.Context) {
	podcasts := []models.Podcast{}
	if err := repository.NewRepository(&models.Podcast{}).FindMany(&podcasts); err != nil {
		ctx.StatusCode(500)
		ctx.JSON(internalServerError)
		return
	}
	ctx.JSON(apiResponse{
		"status": "success",
		"data":   podcasts,
	})
}

func ViewPodcast(ctx iris.Context) {
	podcast := &models.Podcast{}
	if ok := repository.NewRepository(podcast).
		FindOne("id = ?", ctx.URLParam("id")); !ok {
		ctx.JSON(apiResponse{
			"status":  "failed",
			"message": "Podcast not found",
		})
		return
	}

	ctx.JSON(apiResponse{
		"status": "success",
		"data":   podcast,
	})
}

func CreatePodcast(ctx iris.Context) {
	_, h, err := ctx.FormFile("file")
	if err != nil {
		return
	}
	fmt.Println(h.Filename, h.Size, h.Header)
	data := &models.PodcastForm{}
	if err := ctx.ReadForm(data); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(validation.Errors(err))
		return
	}
	fmt.Println(data)

	podcast := &models.Podcast{
		Id:          uuid.New(),
		SubjectId:   data.Subject,
		ObjectiveId: data.Objective,
		Title:       data.Title,
		Url:         "",
	}
	if err := repository.NewRepository(podcast).Create(); err != nil {
		ctx.JSON(internalServerError)
		ctx.StatusCode(http.StatusInternalServerError)
		return
	}
	ctx.JSON(apiResponse{
		"status":  "success",
		"message": "Podcast created successfully",
	})
}

func UpdatePodcast(ctx iris.Context) {
	data := &models.PodcastUpdateForm{}
	ctx.ReadJSON(data)
	podcast := &models.Podcast{}
	if ok := repository.NewRepository(podcast).FindOne("id = ?", data.Id); !ok {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(apiResponse{
			"status":  "failed",
			"message": "Podcast not found",
		})
		return
	}

	if err := repository.NewRepository(podcast).Save(); err != nil {
		ctx.StatusCode(500)
		ctx.JSON(internalServerError)
		return
	}
	ctx.JSON(apiResponse{
		"status":  "success",
		"message": "Updated successfully",
	})
}

func DeletePodcast(ctx iris.Context) {
	podcast := &models.Podcast{}
	if ok := repository.NewRepository(podcast).FindOne("id = ?", ctx.URLParam("id")); !ok {
		ctx.JSON(apiResponse{
			"status":  "failed",
			"message": "not found",
		})
		return
	}

	if err := repository.NewRepository(&models.Mock{}).
		Delete(podcast); err != nil {
		ctx.JSON(apiResponse{
			"status":  "failed",
			"message": "delete failed",
		})
		return
	}
	ctx.JSON(apiResponse{
		"status":  "success",
		"message": "Podcast deleted successfully",
	})
}

//+++++++++++++++++++++++++++++++++++++++++++++++++++++++++//

//+++++++++++++++++++++++++++++++++++++++++++++++++++++++++//
func IndexNewsfeed(ctx iris.Context) {
	newsfeed := []models.Newsfeed{}
	if err := repository.NewRepository(&models.Newsfeed{}).FindMany(&newsfeed); err != nil {
		ctx.StatusCode(500)
		ctx.JSON(internalServerError)
		return
	}
	ctx.JSON(apiResponse{
		"status": "success",
		"data":   newsfeed,
	})
}

func ViewNewsfeed(ctx iris.Context) {
	newsfeed := &models.Newsfeed{}
	if ok := repository.NewRepository(newsfeed).
		FindOne("id = ?", ctx.URLParam("id")); !ok {
		ctx.JSON(apiResponse{
			"status":  "failed",
			"message": "Not found",
		})
		return
	}
	ctx.JSON(apiResponse{
		"status": "success",
		"data":   newsfeed,
	})
}

func CreateNewsfeed(ctx iris.Context) {
	data := &models.NewsfeedForm{}
	if err := ctx.ReadJSON(data); err != nil {
		ctx.JSON(validation.Errors(err))
		return
	}

	slug, err := models.UniqueSlug(&models.Newsfeed{}, list.Slug(data.Title))
	if err != nil {
		ctx.JSON(internalServerError)
		ctx.StatusCode(http.StatusInternalServerError)
		return
	}
	user, _ := getUser(ctx)
	newsfeed := &models.Newsfeed{
		Id:      uuid.New(),
		UserId:  user.Id,
		Slug:    slug,
		Title:   data.Title,
		Content: data.Content,
	}

	if err := repository.NewRepository(newsfeed).Create(); err != nil {
		ctx.JSON(internalServerError)
		ctx.StatusCode(http.StatusInternalServerError)
		return
	}
	ctx.JSON(apiResponse{
		"status":  "success",
		"message": "Created successfully",
	})
}

func UpdateNewsfeed(ctx iris.Context) {
	data := &models.NewsfeedUpdateForm{}
	ctx.ReadJSON(data)
	newsfeed := &models.Newsfeed{}
	if ok := repository.NewRepository(newsfeed).FindOne("slug = ?", data.Slug); !ok {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(apiResponse{
			"status":  "failed",
			"message": "Not found",
		})
		return
	}

	if data.Title != "" {
		newsfeed.Title = data.Title
	}
	if data.Content != "" {
		newsfeed.Content = data.Content
	}

	if err := repository.NewRepository(newsfeed).Save(); err != nil {
		ctx.StatusCode(500)
		ctx.JSON(internalServerError)
		return
	}
	ctx.JSON(apiResponse{
		"status":  "success",
		"message": "Updated successfully",
	})
}

func DeleteNewsfeed(ctx iris.Context) {
	newsfeed := &models.Newsfeed{}
	if ok := repository.NewRepository(newsfeed).FindOne("slug = ?", ctx.URLParam("slug")); !ok {
		ctx.JSON(apiResponse{
			"status":  "failed",
			"message": "not found",
		})
		return
	}

	if err := repository.NewRepository(&models.Mock{}).
		Delete(newsfeed); err != nil {
		ctx.JSON(apiResponse{
			"status":  "failed",
			"message": "delete failed",
		})
		return
	}
	ctx.JSON(apiResponse{
		"status":  "success",
		"message": "Deleted successfully",
	})
}

//+++++++++++++++++++++++++++++++++++++++++++++++++++++++++//

//+++++++++++++++++++++++++++++++++++++++++++++++++++++++++//
// Settings
func Settings(ctx iris.Context) {
	setting := ctx.Params().GetString("setting")
	switch setting {
	case "session":
		ctx.JSON(apiResponse{
			"status": "success",
			"data":   settings.Get("examSession", time.Now().Year()),
		})
	case "privacy", "terms":
		gs := &models.GeneralSetting{}
		if err := repository.NewRepository(gs).FindOneDst(gs); err != nil {
			ctx.StatusCode(500)
			ctx.JSON(internalServerError)
			return
		}
		v := reflect.ValueOf(*gs).FieldByName(gs.Field(setting))
		if (v == reflect.Value{}) {
			ctx.JSON(apiResponse{
				"status":  "failed",
				"message": "setting not found",
			})
			return
		}
		ctx.JSON(apiResponse{
			"status": "success",
			"data":   v.Interface(),
		})
	}
}

func SetSettings(ctx iris.Context) {
	var setSetting map[string]interface{} = map[string]interface{}{}
	ctx.ReadJSON(&setSetting)

	fmt.Println(setSetting)
	setting := ctx.Params().GetString("setting")
	v, ok := setSetting[setting]
	if !ok {
		ctx.JSON(apiResponse{
			"status":  "failed",
			"message": "setting not found",
		})
		return
	}
	switch setting {
	case "session":
		settings.Set("examSession", v)
		ctx.JSON(apiResponse{
			"status":  "success",
			"message": "updated successfully",
		})
	case "privacy", "terms":
		gs := &models.GeneralSetting{}
		if err := repository.NewRepository(gs).FindOneDst(gs); err != nil {
			ctx.StatusCode(500)
			ctx.JSON(internalServerError)
			return
		}
		if err := database.UseDB("app").Model(gs).Update(setting, v).Error; err != nil {
			ctx.StatusCode(500)
			ctx.JSON(internalServerError)
			return
		}
		ctx.JSON(apiResponse{
			"status":  "success",
			"message": "updated successfully",
		})
	}
}

//+++++++++++++++++++++++++++++++++++++++++++++++++++++++++//
