package controllers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Prep50mobileApp/prep50-api/src/models"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/list"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/logger"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/repository"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/settings"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/validation"
	"github.com/Prep50mobileApp/prep50-api/src/services/database"
	"github.com/google/uuid"
	"github.com/kataras/iris/v12"
)

func RegisterUserExams(ctx iris.Context) {
	type RegisterExamForm struct {
		Exams []string `validate:"required"`
	}
	session := settings.Get("examSession", time.Now().Year())
	eff := &RegisterExamForm{}
	if err := ctx.ReadJSON(eff); !logger.HandleError(err) {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(validation.Errors(err))
		return
	}

	user, _ := getUser(ctx)
	userExams := []models.UserExam{}
	for _, v := range eff.Exams {
		e := &models.Exam{}
		if !repository.NewRepository(e).FindOne("name = ?", v) {
			ctx.StatusCode(http.StatusNotFound)
			ctx.JSON(apiResponse{
				"status":  "failed",
				"message": fmt.Sprintf("Exam :%s not found", v),
			})
			return
		}
		uRECS := &models.UserExam{}
		if err := repository.NewRepository(&models.Exam{}).
			FindOneDst(uRECS, "exam_id = ? AND user_id = ? AND session = ?", e.Id, user.Id, session); err != nil &&
			!strings.Contains(err.Error(), "not found") {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(internalServerError)
			return
		}

		if uRECS.Id != uuid.Nil {
			ctx.StatusCode(http.StatusForbidden)
			ctx.JSON(apiResponse{
				"status":  "failed",
				"code":    400,
				"message": fmt.Sprintf("cannot register %s again for current session", v),
			})
			return
		}
		userExams = append(userExams, models.UserExam{
			Id:            uuid.New(),
			UserId:        user.Id,
			ExamId:        e.Id,
			Session:       (uint)(session.(int)),
			PaymentStatus: models.Pending,
			CreatedAt:     time.Now(),
		})
	}
	if len(userExams) > 0 {
		database.UseDB("app").Save(userExams)
	}

	ctx.JSON(apiResponse{
		"status":  "success",
		"message": "exams registered successfully",
	})
}

func PayNow(ctx iris.Context) {

}

// SUBJECTS
func IndexUserSubjects(ctx iris.Context) {
	type (
		Response struct {
			Name string `json:"name"`
			Exam string `json:"exam"`
			models.UserSubject
		}
	)
	user, _ := getUser(ctx)
	session := settings.Get("examSession", time.Now().Year())
	q := []query{}
	if err := database.UseDB("app").Table("user_exams as ue").
		Select("ue.id, ue.exam_id, ue.session, e.name, e.status").
		Joins("LEFT JOIN exams as e on ue.exam_id = e.id").
		Where("ue.user_id = ? AND ue.session = ?", user.Id, session).
		Scan(&q).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(internalServerError)
		return
	}

	qid := []uuid.UUID{}
	for _, i := range q {
		qid = append(qid, i.Id)
	}

	userSubjects := []models.UserSubject{}
	if err := repository.NewRepository(&models.UserSubject{}).
		FindMany(&userSubjects, "user_id = ? AND user_exam_id IN ?", user.Id, qid); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(internalServerError)
		return
	}

	ids := []uint{}
	for _, us := range userSubjects {
		ids = append(ids, us.SubjectId)
	}
	subjects := []models.Subject{}
	if err := repository.NewRepository(&models.Subject{}).FindMany(&subjects, "id IN ?", ids); !logger.HandleError(err) {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(internalServerError)
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

	ctx.JSON(apiResponse{
		"status": "success",
		"data":   response,
	})
}

func CreateUserSubjects(ctx iris.Context) {
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
	data := &UserSubjectForm{}
	ctx.ReadJSON(data)

	user, _ := getUser(ctx)
	session := settings.Get("examSession", time.Now().Year())
	userSubjects := []models.UserSubject{}
	for e, v := range *data {
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
			ctx.StatusCode(http.StatusNotFound)
			ctx.JSON(apiResponse{
				"status":  "failed",
				"message": fmt.Sprintf("exam :%s not found", e),
			})
			return
		}

		if q.SubjectCount < len(v) {
			ctx.StatusCode(http.StatusForbidden)
			ctx.JSON(apiResponse{
				"status":  "failed",
				"code":    401,
				"message": fmt.Sprintf("too many subjects for %s, allowed is %d", q.Name, q.SubjectCount),
			})
			return
		}

		userExamSubject := []models.UserSubject{}
		for _, id := range v {
			if ok := repository.NewRepository(&models.Subject{}).FindOne("id = ?", id); !ok {
				ctx.StatusCode(http.StatusNotFound)
				ctx.JSON(apiResponse{
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
		if err := repository.NewRepository(user).FindMany(&currentUserSubjects, "user_id = ? AND user_exam_id = ?", user.Id, q.ExamId); err != nil {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(internalServerError)
			return
		}
		if q.SubjectCount < len(currentUserSubjects)+len(userExamSubject) {
			ctx.StatusCode(http.StatusForbidden)
			ctx.JSON(apiResponse{
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
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(internalServerError)
			return
		}
	}

	ctx.JSON(apiResponse{
		"status":  "success",
		"message": "subjects added successfully",
	})
}
