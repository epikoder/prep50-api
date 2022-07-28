package middlewares

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Prep50mobileApp/prep50-api/src/models"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/logger"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/repository"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/settings"
	"github.com/Prep50mobileApp/prep50-api/src/services/database"
	"github.com/google/uuid"
	"github.com/kataras/iris/v12"
)

type (
	query struct {
		Id           uuid.UUID
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
	session := settings.Get("examSession", time.Now().Year())
	examSubject := map[query][]models.UserSubject{}
	{
		q := []query{}
		if err := database.UseDB("app").Table("user_exams as ue").
			Select("ue.id, ue.exam_id, ue.session, e.name, e.status, e.subject_count").
			Joins("LEFT JOIN exams as e on ue.exam_id = e.id").
			Where("ue.session = ? AND e.status = ?", session, true).
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
		if l := len(s); e.SubjectCount > l {
			ctx.StatusCode(http.StatusForbidden)
			ctx.JSON(apiResponse{
				"status":  "failed",
				"code":    401,
				"message": fmt.Sprintf("you need to register at least %d subjects on %s, registered %d", e.SubjectCount, e.Name, l),
			})
			return
		}
	}

	ctx.Next()
}

func MustSubscribe(ctx iris.Context) {
	type form struct {
		WithLesson bool
	}
	data := &form{}
	ctx.ReadJSON(data)
	if data.WithLesson {
		user, _ := getUser(ctx)
		session := settings.Get("examSession", time.Now().Year())
		userExams := []models.UserExam{}
		if err := database.UseDB("app").
			Find(&userExams, "user_id = ? AND session = ? AND payment_status = ?",
				user.Id, session, models.Completed).Error; err != nil {
			ctx.StatusCode(500)
			ctx.JSON(internalServerError)
			return
		}
		if len(userExams) == 0 {
			ctx.StatusCode(http.StatusForbidden)
			ctx.JSON(apiResponse{
				"status":  "failed",
				"code":    402,
				"message": "No exam registered",
			})
			return
		}
		paid := false
		for _, ue := range userExams {
			paid = ue.PaymentStatus == models.Completed || paid
		}
		if !paid {
			ctx.StatusCode(http.StatusForbidden)
			ctx.JSON(apiResponse{
				"status":  "failed",
				"code":    403,
				"message": "Please complete payment for one of selected exam",
			})
			return
		}
	}
	ctx.Next()
}
