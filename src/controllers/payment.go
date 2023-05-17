package controllers

import (
	"encoding/json"
	"time"

	"github.com/Prep50mobileApp/prep50-api/src/models"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/logger"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/settings"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/validation"
	"github.com/Prep50mobileApp/prep50-api/src/services/database"
	"github.com/Prep50mobileApp/prep50-api/src/services/payment"
	"github.com/epikoder/paystack-go"
	"github.com/google/uuid"
	"github.com/kataras/iris/v12"
)

func VerifyPayment(ctx iris.Context) {
	type paymentData struct {
		Type      string `validate:"oneof=mock jamb waec"`
		Provider  string
		Reference string `validate:"required"`
		Id        string `validate:"required_if=Type mock"`
	}
	data := &paymentData{}
	if err := ctx.ReadJSON(data); !logger.HandleError(err) {
		ctx.JSON(validation.Errors(err))
		ctx.StatusCode(400)
		return
	}

	provider := payment.New(data.Provider)
	res, err := provider.IVerify(data.Reference)
	if !logger.HandleError(err) {
		ctx.JSON(apiResponse{
			"status":  "failed",
			"message": "Invalid transaction",
		})
		return
	}

	session := uint(settings.Get("exam.session", time.Now().Year()).(int))
	user, _ := getUser(ctx)
	tx := models.Transaction{}
	switch data.Provider {
	default:
		{
			_tx, ok := res.(*paystack.Transaction)
			if !ok || _tx.Status != "success" {
				ctx.JSON(apiResponse{
					"status":  "failed",
					"message": "Failed transaction",
				})
				return
			}
			if err := database.UseDB("app").First(&tx, "reference = ?", data.Reference).Error; err == nil {
				ctx.JSON(apiResponse{
					"status":  "failed",
					"message": "Duplicate transaction",
				})
				return
			}
			b, err := json.Marshal(res)
			if err != nil {
				ctx.JSON(apiResponse{
					"status":  "failed",
					"message": "Unable to validate transaction source",
				})
				return
			}

			tx.Id = uuid.New()
			tx.UserId = user.Id
			tx.Amount = uint(_tx.Amount / 100)
			tx.Reference = _tx.Reference
			tx.Response = string(b)
			tx.Provider = provider.Name()
			tx.Status = string(models.Completed)
			tx.Item = data.Type
		}
	}

	switch item := data.Type; item {
	case "jamb", "waec", "both":
		{
			us := models.UserExam{}
			exam := models.Exam{}
			if notFound := database.UseDB("app").First(&exam, "name = ? AND status = 1", item).Error != nil; notFound {
				ctx.JSON(apiResponse{
					"status":  "failed",
					"message": "Selected exam not found",
				})
				return
			}
			expiresAt := time.Now().AddDate(0, 1, 0)
			if err := database.UseDB("app").Table("user_exams as ue").
				Joins("LEFT JOIN exams as e ON e.id = ue.exam_id").
				First(&us, "e.name = ? AND ue.user_id = ?", item, user.Id).Error; err != nil {
				if exam.Amount != tx.Amount {
					ctx.JSON(apiResponse{
						"status":  "failed",
						"message": "Paid amount is incorrect",
					})
					return
				}

				us := &models.UserExam{
					Id:            uuid.New(),
					UserId:        user.Id,
					ExamId:        exam.Id,
					PaymentStatus: models.Completed,
					TransactionId: tx.Id,
					ExpiresAt:     expiresAt,
				}
				if err := database.UseDB("app").Create(us).Error; err != nil {
					ctx.StatusCode(500)
					ctx.JSON(internalServerError)
					return
				}
			} else {
				if exam.Amount != tx.Amount {
					ctx.JSON(apiResponse{
						"status":  "failed",
						"message": "Paid amount is incorrect",
					})
					return
				}
				us.PaymentStatus = models.Completed
				us.TransactionId = tx.Id
				us.ExpiresAt = expiresAt
				if err := database.UseDB("app").Save(us).Error; err != nil {
					ctx.StatusCode(500)
					ctx.JSON(internalServerError)
					return
				}
			}
		}
	case "mock":
		{
			m := models.Mock{}
			if err := database.UseDB("app").First(&m, `
			id = ? AND session = ? AND start_time > ?
			`, data.Id, session, time.Now()).Error; err != nil {
				ctx.JSON(apiResponse{
					"status":  "failed",
					"message": "Selected mock not found",
				})
				return
			}
			if m.Amount != tx.Amount {
				ctx.JSON(apiResponse{
					"status":  "failed",
					"message": "Paid amount is incorrect",
				})
				return
			}
			user.Mock = append(user.Mock, m)
			if err := user.Database().Save(user.Mock).Error; err != nil {
				ctx.StatusCode(500)
				ctx.JSON(internalServerError)
				return
			}
		}
	}
	if err := database.UseDB("app").Create(&tx).Error; err != nil {
		ctx.StatusCode(500)
		ctx.JSON(internalServerError)
		return
	}

	ctx.JSON(apiResponse{
		"status":  "success",
		"message": "Transaction successful",
	})
}

func HandlePayment(ctx iris.Context) {
	type paymentData struct {
		Action   string
		Type     string `validate:"required"`
		Provider string
		Id       string
	}
	data := &paymentData{}
	if err := ctx.ReadJSON(data); !logger.HandleError(err) {
		ctx.JSON(validation.Errors(err))
		ctx.StatusCode(400)
		return
	}

	provider := payment.New(data.Provider)
	switch data.Action {
	case "verify":
		provider.IVerify(data.Id)
	default:
		provider.ICharge()
	}
}

func PaymentHook(ctx iris.Context) {

}
