package middlewares

import (
	"fmt"
	"net/http"

	"github.com/Prep50mobileApp/prep50-api/src/models"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/logger"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/repository"
	"github.com/Prep50mobileApp/prep50-api/src/services/database"
	"github.com/google/uuid"
	"github.com/kataras/iris/v12"
)

type (
	query struct {
		Id           uuid.UUID
		UserId       uuid.UUID
		ExamId       uuid.UUID
		Name         string
		Session      int
		Status       bool
		SubjectCount int
	}
)

func MustRegisterSubject(ctx iris.Context) {
	i, err := ctx.User().GetRaw()
	if err != nil {
		ctx.StatusCode(http.StatusUnauthorized)
		return
	}
	user, _ := i.(*models.User)
	examSubject := map[query][]models.UserSubject{}
	{
		q := []query{}
		if err := database.UseDB("app").Table("user_exams as ue").
			Select("ue.id, ue.exam_id, ue.user_id, e.name, e.status, e.subject_count").
			Joins("LEFT JOIN exams as e on ue.exam_id = e.id").
			Where("e.status = ? AND e.status = 1 AND ue.user_id = ?", true, user.Id).
			Scan(&q).Error; err != nil {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(internalServerError)
			return
		}
		if len(q) == 0 {
			ctx.StatusCode(http.StatusForbidden)
			ctx.JSON(apiResponse{
				"status":  "failed",
				"code":    402,
				"message": "No exam registered",
			})
			return
		}

		for _, e := range q {
			userSubjects := []models.UserSubject{}
			if err := repository.NewRepository(&models.UserSubject{}).FindMany(&userSubjects, "user_id = ? AND user_exam_id = ?", user.Id, e.Id); !logger.HandleError(err) {
				ctx.StatusCode(http.StatusInternalServerError)
				ctx.JSON(internalServerError)
				return
			}
			examSubject[e] = append(examSubject[e], userSubjects...)
		}
		if len(examSubject) == 0 {
			ctx.StatusCode(http.StatusForbidden)
			ctx.JSON(apiResponse{
				"status":  "failed",
				"code":    400,
				"message": "you have no registered subject",
			})
			return
		}
	}

	for e, s := range examSubject {
		if l := len(s); l < 1 {
			ctx.StatusCode(http.StatusForbidden)
			ctx.JSON(apiResponse{
				"status":  "failed",
				"code":    401,
				"message": fmt.Sprintf("you need to register at least 1 subjects on %s, registered %d", e.Name, l),
			})
			return
		}
	}

	ctx.Next()
}

func MustSubscribe(ctx iris.Context) {
	// type form struct {
	// 	WithLesson bool
	// }
	// data := &form{}
	// ctx.ReadJSON(data)
	// if data.WithLesson {
	// 	user, _ := getUser(ctx)
	// 	session := settings.Get("exam.session", time.Now().Year())
	// 	userExam := &models.UserExam{}
	// 	if err := database.UseDB("app").
	// 		Find(userExam, "user_id = ? AND session = ? AND payment_status = ?",
	// 			user.Id, session, models.Completed).Error; err != nil {
	// 		ctx.StatusCode(500)
	// 		ctx.JSON(internalServerError)
	// 		return
	// 	}
	// 	if userExam.Id == uuid.Nil {
	// 		ctx.StatusCode(http.StatusForbidden)
	// 		ctx.JSON(apiResponse{
	// 			"status":  "failed",
	// 			"code":    403,
	// 			"message": "Please complete payment for registered exam",
	// 		})
	// 		return
	// 	}
	// }
	ctx.Next()
}
