package controllers

import (
	"github.com/Prep50mobileApp/prep50-api/src/models"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/repository"
	"github.com/kataras/iris/v12"
)

func GetSubjects(ctx iris.Context) {
	subjects := []models.Subject{}
	if err := repository.NewRepository(&models.Subject{}).
		FindMany(&subjects); err != nil {
	}
}
