package controllers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Prep50mobileApp/prep50-api/config"
	"github.com/Prep50mobileApp/prep50-api/src/models"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/list"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/logger"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/settings"
	"github.com/Prep50mobileApp/prep50-api/src/services/database"
	"github.com/kataras/iris/v12"
	"gorm.io/gorm"
)

//++++++++++++++++++++++++++++++++++++++++++++++++
func StudySubjects(ctx iris.Context) {
	type (
		subjectForm struct {
			Exam []string
		}
	)
	var response map[string][]models.UserSubjectProgress = make(map[string][]models.UserSubjectProgress)
	data := &subjectForm{}
	ctx.ReadJSON(data)
	user, _ := getUser(ctx)
	session := settings.Get("exam.session", time.Now().Year())
	progress := []models.UserProgress{}
	database.UseDB("app").Find(&progress, "user_id = ?", user.Id)
	subjects := []struct {
		Id          uint   `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Exam        string
	}{}
	if err := database.UseDB("core").Table("subjects as s").
		Select("s.*, e.name as exam").
		Joins(fmt.Sprintf("LEFT JOIN %s.user_subjects as us ON s.id = us.subject_id", config.Conf.Database.App.Name)).
		Joins(fmt.Sprintf("LEFT JOIN %s.user_exams as ue ON us.user_exam_id = ue.id", config.Conf.Database.App.Name)).
		Joins(fmt.Sprintf("LEFT JOIN %s.exams as e ON ue.exam_id = e.id", config.Conf.Database.App.Name)).
		Find(&subjects, "us.user_id = ? AND ue.session = ? order by subject_id asc",
			user.Id,
			session).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(internalServerError)
		return
	}

	for _, sub := range subjects {
		response[sub.Exam] = append(response[sub.Exam], models.UserSubjectProgress{
			Id:          sub.Id,
			Name:        sub.Name,
			Description: sub.Description,
			Progress:    models.FindSubjectProgressFromList(progress, sub.Id),
		})
	}
	ctx.JSON(apiResponse{
		"status": "success",
		"data":   response,
	})
}

func StudyLessons(ctx iris.Context) {
	type topicForm struct {
		Subject              []int
		Objective            []int
		WithObjective        bool
		WithLesson           bool
		FilterEmptyTopic     bool
		FilterEmptyObjective bool
	}
	data := &topicForm{}
	ctx.ReadJSON(data)
	if len(data.Subject) > 0 {
		data.Subject = list.Unique(data.Subject).([]int)
	}
	if len(data.Objective) > 0 {
		data.Objective = list.Unique(data.Objective).([]int)
	}

	user, _ := getUser(ctx)
	session := settings.Get("exam.session", time.Now().Year())
	topics := []models.Topic{}
	{
		db := database.UseDB("core")
		if data.WithObjective {
			if len(data.Objective) > 0 {
				db = db.Preload("Objectives",
					func(__db *gorm.DB) *gorm.DB {
						__db = __db.Table("objectives as o").
							Select("o.*, up.score").
							Joins(fmt.Sprintf("LEFT JOIN %s.user_progresses as up ON up.objective_id = o.id", config.Conf.Database.App.Name)).
							Where("o.id IN ?", data.Objective)
						return __db
					},
				)
			} else {
				db = db.Preload("Objectives", func(__db *gorm.DB) *gorm.DB {
					__db = __db.Table("objectives as o").
						Select("o.*, up.score").
						Joins(fmt.Sprintf("LEFT JOIN %s.user_progresses as up ON up.objective_id = o.id", config.Conf.Database.App.Name))
					return __db
				})
			}
		}
		if data.WithLesson {
			db = db.Preload("Objectives.Lessons")
		}

		var err error
		db = db.Table("topics as t").
			Select("t.*").
			Joins(fmt.Sprintf("LEFT JOIN %s.user_subjects as us ON t.subject_id = us.subject_id", config.Conf.Database.App.Name)).
			Joins(fmt.Sprintf("LEFT JOIN %s.user_exams as ue ON us.user_exam_id = ue.id", config.Conf.Database.App.Name)).
			Joins("LEFT JOIN subjects as s ON s.id = us.subject_id")

		if len(data.Subject) > 0 {
			err = db.
				Find(&topics, "us.user_id = ? AND ue.session = ? AND t.subject_id IN ? GROUP BY t.id order by subject_id asc",
					user.Id,
					session,
					data.Subject,
				).Error
		} else {
			err = db.
				Find(&topics, "us.user_id = ? AND ue.session = ? GROUP BY t.id order by subject_id asc",
					user.Id,
					session,
				).Error

		}
		if !logger.HandleError(err) {
			ctx.StatusCode(500)
			ctx.JSON(internalServerError)
			return
		}
	}

	useFilterEmptyObjective := func(arr *[]models.Objective, o models.Objective) {
		if data.FilterEmptyObjective {
			if len(o.Lessons) > 0 {
				*arr = append(*arr, o)
			}
		} else {
			*arr = append(*arr, o)
		}
	}

	useFilterEmptyTopic := func(arr *[]models.Topic, t models.Topic, objectives []models.Objective) {
		if data.FilterEmptyObjective {

			if len(objectives) > 0 {
				*arr = append(*arr, t)
			}
		} else {
			*arr = append(*arr, t)
		}
	}

	topicLessons := []models.Topic{}
	for _, t := range topics {
		objectives := []models.Objective{}
		for _, o := range t.Objectives {
			useFilterEmptyObjective(&objectives, o)
		}
		useFilterEmptyTopic(&topicLessons, t, objectives)
	}

	ctx.JSON(apiResponse{
		"status": "success",
		"data":   topicLessons,
	})
}

//++++++++++++++++++++++++++++++++++++++++++++++++

//++++++++++++++++++++++++++++++++++++++++++++++++
func StudyPodcasts(ctx iris.Context) {
	type topicForm struct {
		Subject int
	}
	data := &topicForm{}
	ctx.ReadQuery(data)
	topics := []models.PodcastTopic{}
	database.UseDB("core").Table("topics").Preload("Podcast", func(db *gorm.DB) *gorm.DB {
		return db.Table(fmt.Sprintf("%s.podcasts", config.Conf.Database.App.Name))
	}).Find(&topics, "subject_id = ?", data.Subject)
	ctx.JSON(apiResponse{
		"status": "success",
		"data":   topics,
	})
}

//++++++++++++++++++++++++++++++++++++++++++++++++
func QuickQuiz(ctx iris.Context) {
	type quizForm struct {
		Id string
	}

	form := &quizForm{}
	ctx.ReadQuery(form)
	questions := []models.Question{}
	database.UseDB("core").Table("objective_questions as oq").
		Select("q.*").Joins("LEFT JOIN questions as q ON oq.id = q.id").
		Find(&questions, "objective_id = ?", form.Id)
	ctx.JSON(apiResponse{
		"status": "success",
		"data":   questions,
	})
}

func QuickQuizScore(ctx iris.Context) {
	type quizForm struct {
		Id    uint
		Score uint
	}

	form := &quizForm{}
	ctx.ReadJSON(form)
	user, _ := getUser(ctx)

	objective := &models.Objective{}
	if err := database.UseDB("core").First(objective, "id = ?", form.Id).Error; err != nil {
		ctx.JSON(apiResponse{
			"status":  "failed",
			"message": "objective not found",
		})
		return
	}
	progress := &models.UserProgress{
		UserId:      user.Id,
		ObjectiveId: form.Id,
		SubjectId:   uint(objective.SubjectId),
		Score:       form.Score,
	}
	if err := database.UseDB("app").Save(progress).Error; err != nil {
		ctx.StatusCode(500)
		ctx.JSON(internalServerError)
		return
	}

	ctx.JSON(apiResponse{
		"status": "success",
	})
}

//++++++++++++++++++++++++++++++++++++++++++++++++
