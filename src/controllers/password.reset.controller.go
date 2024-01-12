package controllers

import (
	"encoding/json"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/Prep50mobileApp/prep50-api/src/models"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/crypto"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/hash"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/logger"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/page"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/repository"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/sendmail"
	"github.com/Prep50mobileApp/prep50-api/src/services/database"
	"github.com/Prep50mobileApp/prep50-api/src/services/queue"
	"github.com/kataras/iris/v12"
)

type (
	UserToken struct {
		Time time.Time
		V    string
	}
	PasswordResetController struct {
		Ctx iris.Context
	}
)

func (c *PasswordResetController) Get() {
	_token, err := crypto.Aes256Decode(c.Ctx.URLParam("token"))
	if err != nil {
		page.Render(c.Ctx, "password_reset", iris.Map{
			"message": "Invalid/Expired Link",
		})
		return
	}
	token := sendmail.VerificationToken{}
	{
		if err := json.Unmarshal([]byte(_token), &token); !logger.HandleError(err) {
			page.Render(c.Ctx, "password_reset", iris.Map{
				"message": "Invalid/Expired Link",
			})
			return
		}
	}

	if token.Expires.Before(time.Now()) {
		page.Render(c.Ctx, "password_reset", iris.Map{
			"message": "Link has Expired",
		})
		return
	}

	user := &models.User{}
	{
		if ok := repository.NewRepository(user).FindOne("email = ?", token.Email); !ok {
			page.Render(c.Ctx, "password_reset", iris.Map{
				"message": "User not found",
			})
			return
		}
	}
	var t string
	{
		v := UserToken{
			Time: time.Now().Add(time.Minute * 20),
			V:    user.Id.String(),
		}
		b, err := json.Marshal(v)
		if err != nil {
			page.Render(c.Ctx, "password_reset", iris.Map{
				"message": "Something went wrong",
			})
		}
		if t, err = crypto.Aes256Encode(string(b)); err != nil {
			page.Render(c.Ctx, "password_reset", iris.Map{
				"message": "User Error",
			})
			return
		}
	}
	page.Render(c.Ctx, "password_reset", iris.Map{
		"message": "Create new password",
		"user":    user,
		"token":   t,
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
				if err := database.DB().
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

type Password struct {
	Password string `validate:"required"`
	Token    string `validate:"required"`
}

func (c *PasswordResetController) Put(psr Password) {
	user := &models.User{}
	{
		v, err := crypto.Aes256Decode(psr.Token)
		if err != nil {
			c.Ctx.JSON(apiResponse{
				"message": "Invalid Token",
				"status":  "failed",
			})
			return
		}
		token := &UserToken{}
		if err = json.Unmarshal([]byte(v), token); err != nil || token.Time.Before(time.Now()) {
			c.Ctx.JSON(apiResponse{
				"message": "Invalid Token",
				"status":  "failed",
			})
			return
		}

		if ok := repository.NewRepository(user).FindOne("id = ?", token.V); !ok {
			c.Ctx.JSON(apiResponse{
				"message": "User not found",
				"status":  "failed",
			})
			return
		}
	}

	reg := regexp.MustCompile("^[a-zA-Z0-9!@#$&'*+?^_-]+$")
	if len(psr.Password) < 6 || !reg.Match([]byte(psr.Password)) {
		c.Ctx.JSON(apiResponse{
			"status":  "failed",
			"message": "User not found",
			"user":    user,
			"token":   psr.Token,
		})
		return
	}

	var err error
	if user.Password, err = hash.MakeHash(psr.Password); err != nil {
		c.Ctx.JSON(apiResponse{
			"status":  "failed",
			"message": "Something went wrong",
		})
		return
	}
	if err = database.DB().Save(user).Error; err != nil {
		c.Ctx.JSON(apiResponse{
			"status":  "failed",
			"message": "Something went wrong",
		})
		return
	}

	c.Ctx.StatusCode(http.StatusAccepted)
	page.Render(c.Ctx, "password_reset_success", iris.Map{})
}
