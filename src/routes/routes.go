package routes

import (
	"github.com/Prep50mobileApp/prep50-api/src/controllers"
	"github.com/Prep50mobileApp/prep50-api/src/middlewares"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/ijwt"
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
	resources.Get("/subjects", ijwt.JwtGuardMiddleware, middlewares.Protected, controllers.GetSubjects)

	// User Protected Routes
	user := app.Party("/user", ijwt.JwtGuardMiddleware, middlewares.Protected)
	user.Get("/exams", controllers.GetExamTypes)
	user.Post("/exams", controllers.RegisterUserExams)
	user.Get("/subjects", controllers.IndexUserSubjects)
	user.Post("/subjects", controllers.CreateUserSubjects)
	user.Put("/subjects")
	user.Delete("/subjects")

	study := app.Party("/study", ijwt.JwtGuardMiddleware, middlewares.Protected, middlewares.MustRegisterSubject)
	study.Get("/subjects", controllers.StudySubjects)
	study.Get("/topics", middlewares.MustSubscribe, controllers.StudyTopics)

	quiz := app.Party("/weekly-quiz", ijwt.JwtGuardMiddleware, middlewares.Protected, middlewares.MustSubscribe)
	quiz.Get("/", controllers.GetWeekQuiz)

	// Admin Protected Routes
	app.Post("/admin", controllers.AdminLogin)
	admin := app.Party("/admin", ijwt.JwtGuardMiddleware, middlewares.Protected, middlewares.AdminUser)
	admin.Get("/weekly/index", middlewares.ResourcePermission, controllers.UpdateWeeklyQuiz)
	admin.Post("/weekly/create", middlewares.ResourcePermission, controllers.CreateWeeklyQuiz)
	admin.Post("/weekly/update", middlewares.ResourcePermission, controllers.UpdateWeeklyQuizQuestion)
	admin.Put("/weekly/update", middlewares.ResourcePermission, controllers.UpdateWeeklyQuiz)
}
