package controllers

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/Prep50mobileApp/prep50-api/src/models"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/hash"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/helper"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/ijwt"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/logger"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/repository"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/settings"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/validation"
	"github.com/Prep50mobileApp/prep50-api/src/services/database"
	"github.com/kataras/iris/v12"
)

type AccountController struct {
	Ctx iris.Context
}

func (c *AccountController) Get() {
	user, _ := getUser(c.Ctx)

	userExams := []UserExamWithName{}
	database.DB().Table("user_exams as ue").
		Select("ue.session, ue.payment_status, ue.created_at, ue.id, e.name, ue.expires_at").Joins("LEFT JOIN exams as e ON e.id = ue.exam_id").
		Where("user_id = ?", user.Id).
		Scan(&userExams)
	userWithExam := &UserWithExam{*user, userExams, len(userExams) != 0}
	c.Ctx.JSON(apiResponse{
		"status": "success",
		"data":   userWithExam,
	})
}

func (c *AccountController) Put() {
	data := models.User{}
	if err := c.Ctx.ReadJSON(&data); !logger.HandleError(err) {
		c.Ctx.StatusCode(400)
		c.Ctx.JSON(validation.Errors(err))
		return
	}

	user, _ := getUser(c.Ctx)
	user.Phone = data.Phone
	user.Address = data.Address
	user.Gender = models.GetGender(string(data.Gender))
	if data.Photo != "" {
		os.Mkdir("storage/avatar", os.ModePerm)
		fileName := fmt.Sprintf("storage/avatar/%s", data.Photo)
		if err := helper.CopyTempFile(fileName, data.Photo); !logger.HandleError(err) {
			c.Ctx.JSON(apiResponse{
				"status":  "failed",
				"message": "Could not save selected photo",
			})
			return
		}
		user.Photo = fileName
	}

	if err := repository.NewRepository(user).Save(); !logger.HandleError(err) {
		c.Ctx.StatusCode(500)
		c.Ctx.JSON(internalServerError)
		return
	}

	userExams := []UserExamWithName{}
	database.DB().Table("user_exams as ue").
		Select("ue.session, ue.payment_status, ue.created_at, ue.id, e.name, ue.expires_at").Joins("LEFT JOIN exams as e ON e.id = ue.exam_id").
		Where("user_id = ?", user.Id).
		Scan(&userExams)

	userWithExam := &UserWithExam{*user, userExams, len(userExams) != 0}
	token, err := ijwt.GenerateToken(userWithExam, user.UserName)
	if !logger.HandleError(err) {
		c.Ctx.StatusCode(http.StatusInternalServerError)
		c.Ctx.JSON(internalServerError)
		return
	}
	response := ijwt.LoginResponse{
		JwtToken: token, User: userWithExam,
	}

	c.Ctx.JSON(apiResponse{
		"status":  "success",
		"message": "Profile updated successfully",
		"data":    response,
	})
}

func (c *AccountController) Post() {
	f, h, err := c.Ctx.FormFile("photo")
	if !logger.HandleError(err) || !strings.Contains(h.Header.Get("Content-Type"), "image/") {
		c.Ctx.JSON(apiResponse{
			"status":  "failed",
			"message": "No image found",
		})
		return
	}

	if size := settings.Get("max.image.size", 2097152).(int); int(h.Size) > size {
		c.Ctx.JSON(apiResponse{
			"status":  "failed",
			"message": fmt.Sprintf("File exceeds %dM", (size / 1024 / 1024)),
		})
		return
	}

	s, err := helper.SaveTempImage(f)
	if !logger.HandleError(err) {
		c.Ctx.JSON(apiResponse{
			"status":  "failed",
			"message": "Error Occoured",
		})
		return
	}
	c.Ctx.JSON(apiResponse{
		"status": "success",
		"photo":  s,
	})
}

func ChangePassword(ctx iris.Context) {
	data := struct {
		Password     string `validate:"required,min=6"`
		Old_Password string `validate:"required"`
	}{}

	if err := ctx.ReadJSON(&data); !logger.HandleError(err) {
		ctx.StatusCode(400)
		ctx.JSON(validation.Errors(err))
		return
	}

	user, _ := getUser(ctx)
	if ok := hash.CheckHash(user.Password, data.Old_Password); !ok {
		ctx.JSON(apiResponse{
			"status":  "failed",
			"message": "Password is incorrect",
		})
		return
	}

	p, err := hash.MakeHash(data.Password)
	if !logger.HandleError(err) {
		ctx.JSON(apiResponse{
			"status":  "failed",
			"message": "unable to save password",
		})
		return
	}
	user.Password = p
	if err := repository.NewRepository(user).Save(); !logger.HandleError(err) {
		ctx.StatusCode(500)
		ctx.JSON(internalServerError)
		return
	}

	ctx.JSON(apiResponse{
		"status":  "success",
		"message": "Password updated successfully",
	})
}
