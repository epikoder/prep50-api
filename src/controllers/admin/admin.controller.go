package admin

import (
	"fmt"

	"github.com/Prep50mobileApp/prep50-api/src/models"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/logger"
	"github.com/kataras/iris/v12"
)

type (
	apiResponse map[string]interface{}
)

var (
	internalServerError = apiResponse{
		"status":  "failed",
		"message": "error occcured",
	}
	getUser = func(ctx iris.Context) (u *models.User, err error) {
		i, err := ctx.User().GetRaw()
		if !logger.HandleError(err) {
			return nil, err
		}
		var ok bool
		if u, ok = i.(*models.User); !ok {
			return nil, fmt.Errorf("user is nil")
		}
		return u, nil
	}
)
