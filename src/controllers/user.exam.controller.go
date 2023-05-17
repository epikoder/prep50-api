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

type UserExamController struct {
	Ctx iris.Context
}

func (c *UserExamController) Get() {
	type UserExamWithName struct {
		Id            uuid.UUID            `json:"id"`
		Name          string               `json:"exam"`
		Session       uint                 `gorm:"notnull" json:"session"`
		PaymentStatus models.PaymentStatus `json:"payment_status"`
		CreatedAt     time.Time            `json:"created_at"`
		ExpiresAt     *time.Time           `json:"expires_at"`
	}
	user, _ := getUser(c.Ctx)
	session := settings.Get("exam.session", time.Now().Year())
	userExams := []UserExamWithName{}
	if err := database.UseDB("app").Table("user_exams as ue").
		Select("ue.session, ue.payment_status, ue.created_at, ue.id, e.name, ue.expires_at").Joins("LEFT JOIN exams as e ON e.id = ue.exam_id").
		Where("user_id = ? AND session = ?", user.Id, session).
		Scan(&userExams).Error; !logger.HandleError(err) {
		c.Ctx.StatusCode(http.StatusInternalServerError)
		c.Ctx.JSON(internalServerError)
		return
	}

	c.Ctx.JSON(apiResponse{
		"status": "success",
		"data":   userExams,
	})
}

func (c *UserExamController) Post() {
	type RegisterExamForm struct {
		Exams []string `validate:"required"`
	}
	session := settings.Get("exam.session", time.Now().Year())
	data := &RegisterExamForm{}
	if err := c.Ctx.ReadJSON(data); !logger.HandleError(err) {
		c.Ctx.StatusCode(http.StatusBadRequest)
		c.Ctx.JSON(validation.Errors(err))
		return
	}

	user, _ := getUser(c.Ctx)
	userExams := []models.UserExam{}
	for _, v := range data.Exams {
		e := &models.Exam{}
		if !repository.NewRepository(e).FindOne("name = ?", v) {
			c.Ctx.StatusCode(http.StatusNotFound)
			c.Ctx.JSON(apiResponse{
				"status":  "failed",
				"message": fmt.Sprintf("Exam :%s not found", v),
			})
			return
		}
		uRECS := &models.UserExam{}
		if err := repository.NewRepository(&models.Exam{}).
			FindOneDst(uRECS, "exam_id = ? AND user_id = ? AND session = ?", e.Id, user.Id, session); !logger.HandleError(err) &&
			!strings.Contains(err.Error(), "not found") {
			c.Ctx.StatusCode(http.StatusInternalServerError)
			c.Ctx.JSON(internalServerError)
			return
		}

		if uRECS.Id != uuid.Nil {
			c.Ctx.StatusCode(http.StatusForbidden)
			c.Ctx.JSON(apiResponse{
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

	c.Ctx.JSON(apiResponse{
		"status":  "success",
		"message": "exams registered successfully",
	})
}
