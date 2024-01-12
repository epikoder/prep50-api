package controllers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
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
	"gorm.io/gorm"
)

var mu sync.Mutex

func InitializePayment(ctx iris.Context) {
	type paymentInfo struct {
		Email        string `validate:"required"`
		Amount       int    `validate:"required"`
		Callback_Url string `valudate:"required"`
	}
	data := &paymentInfo{}
	if err := ctx.ReadJSON(data); !logger.HandleError(err) {
		ctx.JSON(validation.Errors(err))
		ctx.StatusCode(400)
		return
	}
	provider := payment.New(nil)

	_res, err := provider.Initialize(payment.TransactionRequest{
		Email:       data.Email,
		Amount:      float32(data.Amount),
		Currency:    "NGN",
		CallbackURL: data.Callback_Url,
		Reference:   fmt.Sprintf("T%d", time.Now().Unix()),
		Metadata: map[string]interface{}{
			"action": "initialize",
		},
	})
	if err != nil {
		ctx.JSON(apiResponse{
			"status":  "failed",
			"message": "Unable to initialize transaction",
		})
		return
	}
	ctx.JSON(apiResponse{
		"status": "success",
		"data":   _res,
	})
}

func _savePayment(db *gorm.DB, exam models.Exam, us *models.UserExam) error {
	us.PaymentStatus = models.Completed
	us.CreatedAt = time.Now()
	if us.ExpiresAt.Valid && us.ExpiresAt.Time.Before(us.CreatedAt) {
		us.ExpiresAt.Time = us.CreatedAt.AddDate(0, 1, 0)
	} else {
		us.ExpiresAt = sql.NullTime{Time: us.CreatedAt, Valid: true}
	}
	return db.Save(us).Error
}

func VerifyPayment(ctx iris.Context) {
	type paymentData struct {
		Type      string `validate:"oneof=mock jamb waec both"`
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

	provider := payment.New(&data.Provider)
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
	transaction := models.Transaction{}
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
			if err := database.DB().First(&transaction, "reference = ?", data.Reference).Error; err == nil {
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

			transaction.Id = uuid.New()
			transaction.UserId = user.Id
			transaction.Amount = uint(_tx.Amount / 100)
			transaction.Reference = data.Reference
			transaction.Response = string(b)
			transaction.Provider = provider.Name()
			transaction.Status = string(models.Completed)
			transaction.Item = data.Type
			session := settings.Get("exam.session", time.Now().Year())
			transaction.Session = uint(session.(int))
		}
	}

	switch item := data.Type; item {
	case "jamb", "waec":
		{
			us := &models.UserExam{}
			exam := models.Exam{}
			if notFound := database.DB().First(&exam, "name = ? AND status = 1", item).Error != nil; notFound {
				ctx.JSON(apiResponse{
					"status":  "failed",
					"message": "Selected exam not found",
				})
				return
			}
			if exam.Amount != transaction.Amount {
				ctx.JSON(apiResponse{
					"status":  "failed",
					"message": "Paid amount is incorrect",
				})
				return
			}

			mu.Lock()
			defer mu.Unlock()
			if err = database.DB().Transaction(func(tx *gorm.DB) (err error) {
				if notFound := tx.Table("user_exams as ue").
					Joins("LEFT JOIN exams as e ON e.id = ue.exam_id").
					First(us, "e.name = ? AND ue.user_id = ?", item, user.Id).Error != nil; notFound {

					us = &models.UserExam{
						Id:            uuid.New(),
						UserId:        user.Id,
						ExamId:        exam.Id,
						TransactionId: &transaction.Id,
					}
					if err = _savePayment(tx, exam, us); err != nil {
						return
					}
				} else {
					us.TransactionId = &transaction.Id
					if err = _savePayment(tx, exam, us); err != nil {
						return
					}
				}

				return tx.Create(&transaction).Error
			}); !logger.HandleError(err) {
				ctx.StatusCode(500)
				ctx.JSON(internalServerError)
				return
			}
		}
	case "both":
		exams := []models.Exam{}
		database.DB().Find(&exams, "name = ? OR name = ?", "waec", "jamb")
		bothExam := models.Exam{
			Amount: 0,
		}
		for _, exam := range exams {
			bothExam.Amount += exam.Amount
		}
		if bothExam.Amount != transaction.Amount {
			ctx.JSON(apiResponse{
				"status":  "failed",
				"message": "Paid amount is incorrect",
			})
			return
		}

		mu.Lock()
		defer mu.Unlock()
		if err = database.DB().Transaction(func(tx *gorm.DB) (err error) {
			for _, exam := range exams {
				us := &models.UserExam{}
				if err = database.DB().Table("user_exams as ue").
					Joins("LEFT JOIN exams as e ON e.id = ue.exam_id").
					First(&us, "e.name = ? AND ue.user_id = ?", exam.Name, user.Id).Error; err != nil {

					us = &models.UserExam{
						Id:            uuid.New(),
						UserId:        user.Id,
						ExamId:        exam.Id,
						TransactionId: &transaction.Id,
					}
					if err = _savePayment(tx, exam, us); err != nil {
						return
					}
				} else {
					us.TransactionId = &transaction.Id
					if err = _savePayment(tx, exam, us); err != nil {
						ctx.StatusCode(500)
						ctx.JSON(internalServerError)
						return
					}
				}
			}
			return tx.Create(&transaction).Error
		}); err != nil {
			ctx.StatusCode(500)
			ctx.JSON(internalServerError)
			return
		}

	case "mock":
		{
			m := models.Mock{}
			if err = database.DB().First(&m, `
			id = ? AND session = ? AND start_time > ?
			`, data.Id, session, time.Now()).Error; err != nil {
				ctx.JSON(apiResponse{
					"status":  "failed",
					"message": "Selected mock not found",
				})
				return
			}
			if m.Amount != transaction.Amount {
				ctx.JSON(apiResponse{
					"status":  "failed",
					"message": "Paid amount is incorrect",
				})
				return
			}

			mu.Lock()
			defer mu.Unlock()

			user.Mock = append(user.Mock, m)
			if err = database.DB().Transaction(func(tx *gorm.DB) error {
				if err := tx.Save(user).Error; err != nil {
					return err
				}
				return tx.Create(&transaction).Error
			}); err != nil {
				ctx.StatusCode(500)
				ctx.JSON(internalServerError)
				return
			}
		}
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

	provider := payment.New(&data.Provider)
	switch data.Action {
	case "verify":
		provider.IVerify(data.Id)
	default:
		provider.ICharge(payment.ChargeRequest{})
	}
}

func PaymentHook(ctx iris.Context) {

}
