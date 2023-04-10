package admin

import (
	"fmt"
	"reflect"
	"time"

	"github.com/Prep50mobileApp/prep50-api/src/models"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/logger"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/repository"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/settings"
	"github.com/Prep50mobileApp/prep50-api/src/services/database"
	"github.com/kataras/iris/v12"
)

func Settings(ctx iris.Context) {
	setting := ctx.Params().GetString("setting")
	switch setting {
	case "session":
		ctx.JSON(apiResponse{
			"status": "success",
			"data":   settings.Get("exam.session", time.Now().Year()),
		})
	case "privacy", "terms":
		gs := &models.GeneralSetting{}
		if err := repository.NewRepository(gs).FindOneDst(gs); !logger.HandleError(err) {
			ctx.StatusCode(500)
			ctx.JSON(internalServerError)
			return
		}
		v := reflect.ValueOf(*gs).FieldByName(gs.Field(setting))
		if (v == reflect.Value{}) {
			ctx.JSON(apiResponse{
				"status":  "failed",
				"message": "setting not found",
			})
			return
		}
		ctx.JSON(apiResponse{
			"status": "success",
			"data":   v.Interface(),
		})
	}
}

func SetSettings(ctx iris.Context) {
	var setSetting map[string]interface{} = map[string]interface{}{}
	ctx.ReadJSON(&setSetting)

	fmt.Println(setSetting)
	setting := ctx.Params().GetString("setting")
	v, ok := setSetting[setting]
	if !ok {
		ctx.JSON(apiResponse{
			"status":  "failed",
			"message": "setting not found",
		})
		return
	}
	switch setting {
	case "session":
		settings.Set("exam.session", v)
		ctx.JSON(apiResponse{
			"status":  "success",
			"message": "updated successfully",
		})
	case "privacy", "terms":
		gs := &models.GeneralSetting{}
		if err := repository.NewRepository(gs).FindOneDst(gs); !logger.HandleError(err) {
			ctx.StatusCode(500)
			ctx.JSON(internalServerError)
			return
		}
		if err := database.UseDB("app").Model(gs).Update(setting, v).Error; !logger.HandleError(err) {
			ctx.StatusCode(500)
			ctx.JSON(internalServerError)
			return
		}
		ctx.JSON(apiResponse{
			"status":  "success",
			"message": "updated successfully",
		})
	}
}
