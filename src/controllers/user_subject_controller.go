package controllers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Prep50mobileApp/prep50-api/src/models"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/color"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/list"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/logger"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/repository"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/settings"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/validation"
	"github.com/Prep50mobileApp/prep50-api/src/services/database"
	"github.com/google/uuid"
	"github.com/kataras/iris/v12"
)

type UserSubjectController struct {
	Ctx iris.Context
}

func (c *UserSubjectController) Get() {
	type (
		Response struct {
			Name string `json:"name"`
			Exam string `json:"exam"`
			models.UserSubject
		}
	)
	user, _ := getUser(c.Ctx)
	session := settings.Get("examSession", time.Now().Year())
	q := []query{}
	if err := database.UseDB("app").Table("user_exams as ue").
		Select("ue.id, ue.exam_id, ue.user_id, ue.session, e.name, e.status").
		Joins("LEFT JOIN exams as e on ue.exam_id = e.id").
		Where("ue.user_id = ? AND ue.session = ? AND e.status = 1 AND ue.user_id = ?", user.Id, session, user.Id).
		Scan(&q).Error; err != nil {
		c.Ctx.StatusCode(http.StatusInternalServerError)
		c.Ctx.JSON(internalServerError)
		return
	}

	qid := []uuid.UUID{}
	for _, i := range q {
		qid = append(qid, i.Id)
	}

	userSubjects := []models.UserSubject{}
	if err := repository.NewRepository(&models.UserSubject{}).
		FindMany(&userSubjects, "user_id = ? AND user_exam_id IN ?", user.Id, qid); err != nil {
		c.Ctx.StatusCode(http.StatusInternalServerError)
		c.Ctx.JSON(internalServerError)
		return
	}

	ids := []uint{}
	for _, us := range userSubjects {
		ids = append(ids, us.SubjectId)
	}
	subjects := []models.Subject{}
	if err := repository.NewRepository(&models.Subject{}).FindMany(&subjects, "id IN ?", ids); !logger.HandleError(err) {
		c.Ctx.StatusCode(http.StatusInternalServerError)
		c.Ctx.JSON(internalServerError)
		return
	}

	response := []Response{}
	for _, exam := range q {
		for _, s := range subjects {
			for _, us := range userSubjects {
				if s.Id == us.SubjectId && exam.Id == us.UserExamId {
					response = append(response, Response{
						Name:        s.Name,
						Exam:        exam.Name,
						UserSubject: us,
					})
				}
			}
		}
	}

	c.Ctx.JSON(apiResponse{
		"status": "success",
		"data":   response,
	})
}

type (
	UserSubjectForm map[string][]int
	createQuery     struct {
		Id           uuid.UUID
		ExamId       uuid.UUID
		Name         string
		Session      int
		Status       bool
		SubjectCount int
	}
)

func (c *UserSubjectController) Post() {
	var form UserSubjectForm
	if err := c.Ctx.ReadJSON(&form); err != nil {
		c.Ctx.StatusCode(400)
		c.Ctx.JSON(validation.Errors(err))
		return
	}

	user, _ := getUser(c.Ctx)
	session := settings.Get("examSession", time.Now().Year())
	userSubjects := []models.UserSubject{}
	for e, v := range form {
		if len(v) == 0 {
			continue
		}
		v = list.Unique(v).([]int)
		q := createQuery{}
		if err := database.UseDB("app").Table("user_exams as ue").
			Select("ue.id, ue.exam_id, ue.session, e.name, e.subject_count, e.status").
			Joins("LEFT JOIN exams as e on ue.exam_id = e.id").
			Where("e.name = ? AND ue.session = ?", e, session).
			Scan(&q).Error; err != nil || q.Id == uuid.Nil {
			c.Ctx.StatusCode(http.StatusNotFound)
			c.Ctx.JSON(apiResponse{
				"status":  "failed",
				"message": fmt.Sprintf("exam :%s not found", e),
			})
			return
		}

		if q.SubjectCount < len(v) {
			c.Ctx.StatusCode(http.StatusForbidden)
			c.Ctx.JSON(apiResponse{
				"status":  "failed",
				"code":    401,
				"message": fmt.Sprintf("too many subjects for %s, allowed is %d", q.Name, q.SubjectCount),
			})
			return
		}

		userExamSubject := []models.UserSubject{}
		for _, id := range v {
			if ok := repository.NewRepository(&models.Subject{}).FindOne("id = ?", id); !ok {
				c.Ctx.StatusCode(http.StatusNotFound)
				c.Ctx.JSON(apiResponse{
					"status":  "failed",
					"message": fmt.Sprintf("subject :%d not found", id),
				})
				return
			}
			// Avoid adding already existing subject
			if ok := repository.NewRepository(&models.UserSubject{}).FindOne("subject_id = ? AND user_id = ? AND user_exam_id = ?", id, user.Id, q.Id); ok {
				continue
			}
			userExamSubject = append(userExamSubject, models.UserSubject{
				Id:         uuid.New(),
				SubjectId:  uint(id),
				UserId:     user.Id,
				UserExamId: q.Id,
			})
		}
		currentUserSubjects := []models.UserSubject{}
		if err := repository.NewRepository(user).FindMany(&currentUserSubjects, "user_id = ? AND user_exam_id = ?", user.Id, q.Id); err != nil {
			c.Ctx.StatusCode(http.StatusInternalServerError)
			c.Ctx.JSON(internalServerError)
			return
		}
		fmt.Println(color.Red, q, q.SubjectCount, len(currentUserSubjects), len(userExamSubject), color.Reset)
		if q.SubjectCount < len(currentUserSubjects)+len(userExamSubject) {
			c.Ctx.StatusCode(http.StatusForbidden)
			c.Ctx.JSON(apiResponse{
				"status":  "failed",
				"code":    401,
				"message": fmt.Sprintf("too many subjects for %s, allowed is %d", q.Name, q.SubjectCount),
			})
			return
		}
		userSubjects = append(userSubjects, userExamSubject...)
	}

	if len(userSubjects) > 0 {
		if err := repository.NewRepository(&models.UserSubject{}).Save(userSubjects); err != nil {
			c.Ctx.StatusCode(http.StatusInternalServerError)
			c.Ctx.JSON(internalServerError)
			return
		}
	}

	c.Ctx.JSON(apiResponse{
		"status":  "success",
		"message": "subjects added successfully",
	})
}

func (c *UserSubjectController) Put() {

}
