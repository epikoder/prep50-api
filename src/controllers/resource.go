package controllers

import (
	"github.com/Prep50mobileApp/prep50-api/src/models"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/repository"
	"github.com/kataras/iris/v12"
)

func GetExamTypes(ctx iris.Context) {
	exams := []models.Exam{}
	repository.NewRepository(&models.Exam{}).
		FindMany(&exams)
	ctx.JSON(apiResponse{
		"status": "success",
		"data":   exams,
	})
}

func GetSubjects(ctx iris.Context) {
	subjects := []models.Subject{}
	if err := repository.NewRepository(&models.Subject{}).FindMany(&subjects); err != nil {
		ctx.StatusCode(500)
		ctx.JSON(internalServerError)
		return
	}
	ctx.JSON(apiResponse{
		"status": "success",
		"data":   subjects,
	})
}
