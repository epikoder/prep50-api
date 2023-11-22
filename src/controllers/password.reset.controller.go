package controllers

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/Prep50mobileApp/prep50-api/src/models"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/crypto"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/logger"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/repository"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/sendmail"
	"github.com/Prep50mobileApp/prep50-api/src/services/database"
	"github.com/Prep50mobileApp/prep50-api/src/services/queue"
	"github.com/kataras/iris/v12"
)

type PasswordResetController struct {
	Ctx iris.Context
}

func (c *PasswordResetController) Get() {
	_token, err := crypto.Aes256Decode(c.Ctx.URLParam("token"))
	if err != nil {
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
			message := "Please use the social login"
			{
				provider := &struct{ Name string }{}
				if err := database.UseDB("app").
					Table("providers as p").
					Select("p.name").
					Joins("LEFT JOIN user_providers as up ON up.provider_id = p.id").
					First(provider, "user_id = ?", user.Id).Error; err == nil {
					message = "Please use " + provider.Name + " login"
				}
			}
			c.Ctx.StatusCode(http.StatusForbidden)
			c.Ctx.JSON(apiResponse{
				"status":  "failed",
				"code":    400,
				"message": message,
			})
			return
		}
	}

	queue.Dispatch(queue.Job{
		Type: queue.SendMail,
		Func: func() error {
			return sendmail.SendPasswordResetMail(user, os.Getenv("HOST_NAME"))
		},
		Retries: 3,
	})

	c.Ctx.JSON(apiResponse{
		"status":  "success",
		"message": "We just sent you an email so you can create a new password for your account and get back to your fun studies.",
	})
}
