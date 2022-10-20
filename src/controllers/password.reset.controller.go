package controllers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/Prep50mobileApp/prep50-api/src/models"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/crypto"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/helper"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/logger"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/repository"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/sendmail"
	"github.com/Prep50mobileApp/prep50-api/src/services/queue"
	"github.com/kataras/iris/v12"
)

type PasswordResetController struct {
	Ctx iris.Context
}

func (c *PasswordResetController) Get() {
	_token, err := crypto.Aes256Decode(c.Ctx.URLParam("token"))
	if err == nil {
		c.Ctx.View("password_reset", iris.Map{
			"message": "Invalid/Expired Link",
		})
		return
	}
	token := sendmail.VerificationToken{}
	if err := json.Unmarshal([]byte(_token), &token); !logger.HandleError(err) {
		c.Ctx.View("password_reset", iris.Map{
			"message": "Invalid/Expired Link",
		})
		return
	}

	if token.Expires.Before(time.Now()) {
		c.Ctx.View("password_reset", iris.Map{
			"message": "Link has Expired",
		})
		return
	}

	user := &models.User{}
	if ok := repository.NewRepository(user).FindOne("email = ?", token.Email); !ok {
		return
	}
	c.Ctx.View("password_reset", iris.Map{
		"message": "Create new password",
		"user":    user,
	})
}

type PasswordResetForm struct {
	User string `validate:"required"`
}

func (c *PasswordResetController) Post(prs PasswordResetForm) {
	user := &models.User{}
	{
		if ok := repository.NewRepository(user).FindOne("email = ? OR username = ?", prs.User, prs.User); !ok {
			c.Ctx.StatusCode(http.StatusNotFound)
			c.Ctx.JSON(apiResponse{
				"status":  "failed",
				"message": "user not found",
			})
			return
		}

		if user.IsProvider {
			c.Ctx.StatusCode(http.StatusForbidden)
			c.Ctx.JSON(apiResponse{
				"status":  "failed",
				"code":    400,
				"message": "password reset not available for this account",
			})
			return
		}
	}

	queue.Dispatch(queue.Job{
		Type: queue.SendMail,
		Func: func() error {
			return sendmail.SendPasswordResetMail(user, helper.GetOrigin(c.Ctx))
		},
		Schedule: time.Now().Add(time.Second * 5),
		Retries:  3,
	})

	c.Ctx.JSON(apiResponse{
		"status":  "success",
		"message": "We just sent you an email so you can create a new password for your account and get back to your fun studies.",
	})
}
