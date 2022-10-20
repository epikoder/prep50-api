package controllers

import (
	"reflect"

	"github.com/Prep50mobileApp/prep50-api/src/models"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/logger"
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
	if err := repository.NewRepository(&models.Subject{}).FindMany(&subjects); !logger.HandleError(err) {
		ctx.StatusCode(500)
		ctx.JSON(internalServerError)
		return
	}
	ctx.JSON(apiResponse{
		"status": "success",
		"data":   subjects,
	})
}

func GetQuestionTypes(ctx iris.Context) {
	type QT struct {
		Id   uint   `json:"id"`
		Name string `json:"name"`
	}
	ctx.JSON(apiResponse{
		"status": "success",
		"data": []QT{
			{
				Name: "Objective",
				Id:   models.OBJECTIVE,
			},
			{
				Name: "Theory",
				Id:   models.THEORY,
			},
			{
				Name: "Practical",
				Id:   models.PRACTICAL,
			},
		},
	})
}

func GetStatic(ctx iris.Context) {
	gs := &models.GeneralSetting{}
	if err := repository.NewRepository(gs).FindOneDst(gs); !logger.HandleError(err) {
		ctx.StatusCode(500)
		ctx.JSON(internalServerError)
		return
	}
	v := reflect.ValueOf(*gs).FieldByName(gs.Field(ctx.Params().GetString("page")))
	if (v == reflect.Value{}) {
		ctx.JSON(apiResponse{
			"status": "failed",
			"data":   "not found",
		})
		return
	}
	ctx.JSON(apiResponse{
		"status": "success",
		"data":   v.Interface(),
	})
}
