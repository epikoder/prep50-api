package routes

import (
	"github.com/Prep50mobileApp/prep50-api/src/controllers"
	"github.com/kataras/iris/v12"
)

func RegisterApiRoutes(app *iris.Application) {
	auth := app.Party("/auth")
	auth.Post("/register", controllers.RegisterV1)
	auth.Post("/login", controllers.LoginV1)
	auth.Post("/social/{id}", controllers.SocialV1)
}
