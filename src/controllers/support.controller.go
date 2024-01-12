package controllers

import (
	"github.com/Prep50mobileApp/prep50-api/src/models"
	"github.com/Prep50mobileApp/prep50-api/src/services/database"
	"github.com/kataras/iris/v12"
)

func Contacts(ctx iris.Context) {
	st := &models.GeneralSetting{}
	database.DB().First(st)
	ctx.JSON(apiResponse{
		"status": "success",
		"data": apiResponse{
			"email":    st.Email,
			"phone":    st.Phone,
			"website":  st.Website,
			"location": st.Location,
		},
	})
}
