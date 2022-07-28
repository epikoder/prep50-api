package controllers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Prep50mobileApp/prep50-api/src/models"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/hash"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/logger"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/repository"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/sendmail"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/validation"
	"github.com/Prep50mobileApp/prep50-api/src/services/queue"
	"github.com/google/uuid"
	"github.com/kataras/iris/v12"
)

type (
	query struct {
		Id      uuid.UUID
		ExamId  uuid.UUID
		Name    string
		Session int
		Status  bool
	}
)

var (
	internalServerError = apiResponse{
		"status":  "failed",
		"message": "error occcured",
	}

	getUser = func(ctx iris.Context) (u *models.User, err error) {
		i, err := ctx.User().GetRaw()
		if err != nil {
			return nil, err
		}
		var ok bool
		if u, ok = i.(*models.User); !ok {
			return nil, fmt.Errorf("user is nil")
		}
		return u, nil
	}
)

func PasswordReset(ctx iris.Context) {
	type PasswordResetStruct struct {
		Email string `validate:"required,email"`
	}
	prs := PasswordResetStruct{}
	if err := ctx.ReadJSON(&prs); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(validation.Errors(err))
		return
	}
	user := &models.User{}
	{
		if ok := repository.NewRepository(user).FindOne("email = ?", prs.Email); !ok {
			ctx.StatusCode(http.StatusUnauthorized)
			ctx.JSON(apiResponse{
				"status":  "failed",
				"message": "user not found",
			})
			return
		}

		if user.IsProvider {
			ctx.StatusCode(http.StatusForbidden)
			ctx.JSON(apiResponse{
				"status":  "failed",
				"code":    400,
				"message": "password reset for available for this account",
			})
			return
		}
	}

	queue.Dispatch(queue.Job{
		Type: queue.SendMail,
		Func: func() error {
			return sendmail.SendPasswordResetMail(user)
		},
		Schedule: time.Now().Add(time.Second * 5),
		Retries:  3,
	})

	ctx.JSON(apiResponse{
		"status":  "success",
		"message": "Use the code sent to your email address to continue",
	})
}

func CompletePasswordReset(ctx iris.Context) {
	type CompletePasswordResetStruct struct {
		Email    string `validate:"required,email"`
		Code     string `validate:"required,numeric"`
		Password string `validate:"required,alphanum"`
	}

	cprs := CompletePasswordResetStruct{}
	if err := ctx.ReadJSON(&cprs); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(validation.Errors(err))
		return
	}

	{
		pr := &models.PasswordReset{}
		if ok := repository.NewRepository(pr).FindOne("code = ? AND email = ?", cprs.Code, cprs.Email); !ok || time.Since(pr.CreatedAt).Minutes() > 30 {
			ctx.StatusCode(http.StatusBadRequest)
			ctx.JSON(apiResponse{
				"status":  "failed",
				"message": "expired or invalid code",
			})
		}
		return
	}

	user := &models.User{}
	{
		if ok := repository.NewRepository(user).FindOne("email = ?", cprs.Email); !ok {
			ctx.StatusCode(http.StatusUnauthorized)
			ctx.JSON(apiResponse{
				"status":  "failed",
				"message": "user does not exist",
			})
			return
		}

		if user.IsProvider {
			ctx.StatusCode(http.StatusForbidden)
			ctx.JSON(apiResponse{
				"status":  "failed",
				"code":    400,
				"message": "password reset for available for this account",
			})
			return
		}
		if ok := hash.CheckHash(user.Password, cprs.Password); ok {
			ctx.StatusCode(http.StatusForbidden)
			ctx.JSON(apiResponse{
				"status":  "failed",
				"code":    401,
				"message": "New password is same as old password",
			})
			return
		}
	}
	p, err := hash.MakeHash(cprs.Password)
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(apiResponse{
			"status":  "failed",
			"message": "error occcured",
		})
		return
	}

	user.Password = p
	if err := repository.NewRepository(user).Save(); !logger.HandleError(err) {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(apiResponse{
			"status":  "failed",
			"message": "error occcured",
		})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(apiResponse{
		"status":  "success",
		"message": "Password reset successful",
	})
}
