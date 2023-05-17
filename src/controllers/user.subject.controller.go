package controllers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Prep50mobileApp/prep50-api/src/models"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/color"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/list"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/logger"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/repository"
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
	q := []query{}
	if err := database.UseDB("app").Table("user_exams as ue").
		Select("ue.id, ue.exam_id, ue.user_id, e.name, e.status").
		Joins("LEFT JOIN exams as e on ue.exam_id = e.id").
		Where("ue.user_id = ? AND e.status = 1", user.Id).
		Scan(&q).Error; !logger.HandleError(err) {
		c.Ctx.StatusCode(http.StatusInternalServerError)
		c.Ctx.JSON(internalServerError)
		return
	}
	fmt.Println(color.Red, q, color.Reset)

	qid := []uuid.UUID{}
	for _, i := range q {
		qid = append(qid, i.Id)
	}

	userSubjects := []models.UserSubject{}
	if err := repository.NewRepository(&models.UserSubject{}).
		FindMany(&userSubjects, "user_id = ? AND user_exam_id IN ?", user.Id, qid); !logger.HandleError(err) {
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
	UserExamQuery   struct {
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
	if err := c.Ctx.ReadJSON(&form); !logger.HandleError(err) {
		c.Ctx.StatusCode(400)
		c.Ctx.JSON(validation.Errors(err))
		return
	}

	user, _ := getUser(c.Ctx)
	userSubjects := []models.UserSubject{}
	for e, v := range form {
		if len(v) == 0 {
			continue
		}
		v = list.Unique(v).([]int)

	QUERY_USER_EXAM:
		q := UserExamQuery{}
		if err := database.UseDB("app").Table("user_exams as ue").
			Select("ue.id, ue.exam_id, ue.user_id, e.name, e.subject_count, e.status").
			Joins("LEFT JOIN exams as e on ue.exam_id = e.id").
			Where("e.name = ? AND ue.user_id = ?", e, user.Id).
			First(&q).Error; !logger.HandleError(err) {
			exam := &models.Exam{}
			if !repository.NewRepository(exam).FindOne("name = ?", e) {
				c.Ctx.StatusCode(http.StatusNotFound)
				c.Ctx.JSON(apiResponse{
					"status":  "failed",
					"message": fmt.Sprintf("Exam :%s not found", e),
				})
				return
			}
			userExams := &models.UserExam{
				Id:            uuid.New(),
				UserId:        user.Id,
				ExamId:        exam.Id,
				PaymentStatus: models.Pending,
				CreatedAt:     time.Now(),
			}
			if err := database.UseDB("app").Create(userExams).Error; !logger.HandleError(err) {
				c.Ctx.StatusCode(http.StatusInternalServerError)
				c.Ctx.JSON(internalServerError)
				return
			}
			goto QUERY_USER_EXAM
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
		if err := repository.NewRepository(user).FindMany(&currentUserSubjects, "user_id = ? AND user_exam_id = ?", user.Id, q.Id); !logger.HandleError(err) {
			c.Ctx.StatusCode(http.StatusInternalServerError)
			c.Ctx.JSON(internalServerError)
			return
		}

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
		if err := repository.NewRepository(&models.UserSubject{}).Save(userSubjects); !logger.HandleError(err) {
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
	type SubjectUpdateForm map[string]struct {
		Action string
		Id     []uint
	}
	data := SubjectUpdateForm{}
	if err := c.Ctx.ReadJSON(&data); !logger.HandleError(err) {
		c.Ctx.StatusCode(400)
		c.Ctx.JSON(validation.Errors(err))
		return
	}

	user, _ := getUser(c.Ctx)
	for e, v := range data {
		q := UserExamQuery{}
		if err := database.UseDB("app").Table("user_exams as ue").
			Select("ue.id, ue.exam_id, e.name, e.subject_count, e.status").
			Joins("LEFT JOIN exams as e on ue.exam_id = e.id").
			Where("e.name = ? AND ue.user_id = ?", e, user.Id).
			First(&q).Error; !logger.HandleError(err) {
			c.Ctx.StatusCode(http.StatusNotFound)
			c.Ctx.JSON(apiResponse{
				"status":  "failed",
				"message": fmt.Sprintf("exam :%s not found", e),
			})
			return
		}

		currentUserSubjects := []models.UserSubject{}
		if err := repository.NewRepository(user).FindMany(&currentUserSubjects, "user_id = ? AND user_exam_id = ?", user.Id, q.Id); !logger.HandleError(err) {
			c.Ctx.StatusCode(http.StatusInternalServerError)
			c.Ctx.JSON(internalServerError)
			return
		}

		switch action := v.Action; strings.ToLower(action) {
		case "remove":
			{
				for i, us := range currentUserSubjects {
					if list.Contains(v.Id, us.SubjectId) && len(currentUserSubjects) > 0 {
						if err := us.Database().Delete(us).Error; logger.HandleError(err) {
							currentUserSubjects = append(append(make([]models.UserSubject, 0), currentUserSubjects[:i]...), currentUserSubjects[i+1:]...)
						}
					}
				}
			}
		default:
			{
				for _, us := range currentUserSubjects {
					tmp := v.Id
					for index, id := range tmp {
						if id == us.SubjectId {
							v.Id = append(append(make([]uint, 0), v.Id[:index]...), v.Id[index+1:]...)
						}
					}
				}

				if len(currentUserSubjects)+len(v.Id) > q.SubjectCount {
					c.Ctx.JSON(apiResponse{
						"statsu":  "failed",
						"message": fmt.Sprintf("Maximum number of subject for %s exceeded", q.Name),
					})
					return
				}

				if len(v.Id) > 0 {
					subs := []models.UserSubject{}
					for _, id := range v.Id {
						subs = append(subs, models.UserSubject{
							Id:         uuid.New(),
							UserId:     user.Id,
							UserExamId: q.Id,
							SubjectId:  id,
						})
					}
					if err := database.UseDB("app").Create(subs).Error; !logger.HandleError(err) {
						c.Ctx.StatusCode(500)
						c.Ctx.JSON(internalServerError)
						return
					}
				}
			}
		}

		c.Ctx.JSON(apiResponse{
			"status":  "success",
			"message": "Subjects updated successfully",
		})
	}
}
