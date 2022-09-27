package controllers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Prep50mobileApp/prep50-api/src/models"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/logger"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/repository"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/settings"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/validation"
	"github.com/Prep50mobileApp/prep50-api/src/services/database"
	"github.com/google/uuid"
	"github.com/kataras/iris/v12"
)

//+++++++++++++++++++++++++++++++++++++++++++++++++++
func UserExams(ctx iris.Context) {
	type UserExamWithName struct {
		Id            uuid.UUID            `json:"id"`
		Name          string               `json:"exam"`
		Session       uint                 `gorm:"notnull" json:"session"`
		PaymentStatus models.PaymentStatus `json:"payment_status"`
		CreatedAt     time.Time            `json:"created_at"`
	}
	user, _ := getUser(ctx)
	session := settings.Get("examSession", time.Now().Year())
	userExams := []UserExamWithName{}
	if err := database.UseDB("app").Table("user_exams as ue").
		Select("ue.session, ue.payment_status, ue.created_at, ue.id, e.name").Joins("LEFT JOIN exams as e ON e.id = ue.exam_id").
		Where("user_id = ? AND session = ?", user.Id, session).
		Scan(&userExams).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(internalServerError)
		return
	}

	ctx.JSON(apiResponse{
		"status": "success",
		"data":   userExams,
	})
}

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

//+++++++++++++++++++++++++++++++++++++++++++++++++++

//+++++++++++++++++++++++++++++++++++++++++++++++++++
func PayNow(ctx iris.Context) {

}

//+++++++++++++++++++++++++++++++++++++++++++++++++++
