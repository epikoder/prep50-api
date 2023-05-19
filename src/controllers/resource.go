package controllers

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/Prep50mobileApp/prep50-api/src/models"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/logger"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/repository"
	"github.com/kataras/iris/v12"
)

func GetExamTypes(ctx iris.Context) {
	exams := []models.Exam{}
	repository.NewRepository(&models.Exam{}).
		FindMany(&exams)

	bothExam := models.Exam{
		Name:         "BOTH",
		Amount:       0,
		SubjectCount: 0,
		Description:  "Pay for Both Exam",
		Status:       true,
	}
	for _, e := range exams {
		fmt.Println(e.Name)
		if strings.EqualFold(e.Name, "waec") || strings.EqualFold(e.Name, "jamb") {
			bothExam.Amount += e.Amount
			bothExam.SubjectCount += e.SubjectCount
		}
	}
	exams = append(exams, bothExam)
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
