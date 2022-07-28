package controllers

import (
	"net/http"
	"time"

	"github.com/Prep50mobileApp/prep50-api/src/models"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/list"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/logger"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/repository"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/settings"
	"github.com/Prep50mobileApp/prep50-api/src/services/database"
	"github.com/kataras/iris/v12"
)

func StudySubjects(ctx iris.Context) {
	type (
		subjectForm struct {
			Exam          []string
			WithObjective bool
		}
	)
	var response map[string][]models.Subject = make(map[string][]models.Subject)

	data := &subjectForm{}
	ctx.ReadJSON(data)
	user, _ := getUser(ctx)
	session := settings.Get("examSession", time.Now().Year())
	examSubject := map[query][]models.UserSubject{}
	{
		q := []query{}
		if err := database.UseDB("app").Table("user_exams as ue").
			Select("ue.id, ue.exam_id, ue.session, e.name, e.status").
			Joins("LEFT JOIN exams as e on ue.exam_id = e.id").
			Where("ue.session = ?", session).
			Scan(&q).Error; err != nil {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(internalServerError)
			return
		}

		for _, e := range q {
			if len(data.Exam) > 0 && !list.Contains(data.Exam, e.Name) || !e.Status {
				continue
			}

			userSubjects := []models.UserSubject{}
			if err := repository.NewRepository(&models.UserSubject{}).FindMany(&userSubjects, "user_id = ? AND user_exam_id = ?", user.Id, e.Id); !logger.HandleError(err) {
				ctx.StatusCode(http.StatusInternalServerError)
				ctx.JSON(internalServerError)
				return
			}
			examSubject[e] = append(examSubject[e], userSubjects...)
		}
	}
	ids := map[string][]uint{}
	for e, us := range examSubject {
		for _, s := range us {
			ids[e.Name] = append(ids[e.Name], s.SubjectId)
		}
	}
	repo := repository.NewRepository(&models.Subject{})
	if data.WithObjective {
		repo.Preload("Objectives")
	}
	for e := range examSubject {
		subjects := []models.Subject{}
		if err := repo.FindMany(&subjects, "id IN ?", ids[e.Name]); err != nil {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(internalServerError)
			return
		}
		response[e.Name] = append(response[e.Name], subjects...)
	}
	ctx.JSON(apiResponse{
		"status": "success",
		"data":   response,
	})
}

func StudyTopics(ctx iris.Context) {
	// TODO:  paid
	type topicForm struct {
		Subject       []int
		Objective     []int
		WithObjective bool
		WithLesson    bool
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
	session := settings.Get("examSession", time.Now().Year())
	type ID struct {
		SubjectId     int
		PaymentStatus models.PaymentStatus
	}
	ids := []ID{}
	if err := database.UseDB("app").
		Table((&models.UserSubject{}).Tag()+" as us").
		Select("us.subject_id").
		Joins("LEFT JOIN user_exams as ue on ue.id = us.user_exam_id").
		Where("us.user_id = ? AND ue.session = ? AND ue.payment_status = ?", user.Id, session, models.Completed).
		Find(&ids).Error; err != nil {
		ctx.StatusCode(500)
		ctx.JSON(internalServerError)
		return
	}

	allowedIds := []int{}
	for _, i := range ids {
		allowedIds = append(allowedIds, i.SubjectId)
	}

	if len(data.Subject) > 0 {
		tmp := allowedIds
		allowedIds = []int{}
		for _, id := range data.Subject {
			if list.Contains(tmp, id) {
				allowedIds = append(allowedIds, id)
			}
		}
	}

	topics := []models.Topic{}
	repo := repository.NewRepository(&models.Topic{})
	if data.WithObjective {
		if len(data.Objective) > 0 {
			repo.Preload("Objectives", "id IN ?", data.Objective)
		} else {
			repo.Preload("Objectives")
		}
	}
	if data.WithLesson {
		repo.Preload("Objectives.Lessons")
	}

	if err := repo.FindMany(&topics, "subject_id in ? order by subject_id asc", allowedIds); err != nil {
		ctx.StatusCode(500)
		ctx.JSON(internalServerError)
		return
	}

	if len(data.Objective) > 0 {
		tmp := topics
		topics = []models.Topic{}
		for _, t := range tmp {
			tmpObj := t.Objectives
			t.Objectives = []models.TopicObjective{}
			for _, o := range tmpObj {
				if list.Contains(data.Objective, int(o.Id)) {
					t.Objectives = append(t.Objectives, o)
				}
			}
			if len(t.Objectives) > 0 {
				topics = append(topics, t)
			}
		}
	}

	ctx.JSON(apiResponse{
		"status": "success",
		"data":   topics,
	})
}
