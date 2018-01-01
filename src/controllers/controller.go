package controllers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Prep50mobileApp/prep50-api/src/models"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/hash"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/repository"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/sendmail"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/validation"
	"github.com/Prep50mobileApp/prep50-api/src/services/database/queue"
	"github.com/kataras/iris/v12"
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
		if ok := repository.NewRepository(user).FindByField("email = ?", prs.Email); !ok {
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
				"message": "",
			})
			return
		}
	}

	queue.Dispatch(queue.Job{
		Type: queue.SendMail,
		Func: func() error {
			return sendmail.SendPasswordResetMail(user)
			// fmt.Println(5)
			// return nil
		},
		Schedule: time.Now().Add(time.Second * 5),
		Retries:  3,
	})

	ctx.JSON(apiResponse{
		"status":  "success",
		"message": "Enter the code sent to your email address to continue",
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
		if ok := repository.NewRepository(pr).FindByField("code = ? AND email = ?", cprs.Code, cprs.Email); !ok || time.Since(pr.CreatedAt).Minutes() > 30 {
			ctx.StatusCode(http.StatusBadRequest)
			ctx.JSON(apiResponse{
				"status":  "failed",
				"message": "expired or invalid code",
			})
		}
		fmt.Println(*pr)
	}

	user := &models.User{}
	{
		if ok := repository.NewRepository(user).FindByField("email = ?", cprs.Email); !ok {
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
				"message": "",
			})
			return
		}
		if ok := hash.CheckHash(user.Password, cprs.Password); ok {
			ctx.StatusCode(http.StatusForbidden)
			ctx.JSON(apiResponse{
				"status":  "failed",
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
	if err := repository.NewRepository(user).Save(); err != nil {
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
