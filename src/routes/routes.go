package routes

import (
	"github.com/Prep50mobileApp/prep50-api/src/controllers"
	"github.com/Prep50mobileApp/prep50-api/src/middlewares"
	"github.com/kataras/iris/v12"
)

func RegisterApiRoutes(app *iris.Application) {
	auth := app.Party("/auth")
	auth.Post("/register", controllers.RegisterV1)
	auth.Post("/login", controllers.LoginV1)
	auth.Post("/provider/{provider:string}", controllers.SocialV1)

	app.Get("/password-reset", middlewares.RateLimiter(), controllers.PasswordReset)
	app.Post("/password-reset", middlewares.RateLimiter(), controllers.CompletePasswordReset)

	resources := app.Party("/resources")
	resources.Get("/subjects", controllers.GetSubjects)
}
