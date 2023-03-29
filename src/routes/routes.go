package routes

import (
	"strings"

	"github.com/Prep50mobileApp/prep50-api/src/controllers"
	admin_controllers "github.com/Prep50mobileApp/prep50-api/src/controllers/admin"
	"github.com/Prep50mobileApp/prep50-api/src/middlewares"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/ijwt"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
)

func RegisterApiRoutes(app *iris.Application) {
	app.AllowMethods(iris.MethodOptions)

	app.HandleDir("storage", iris.Dir("./storage"), iris.DirOptions{Compress: true})

	mvc.New(app.Party("password-reset")).Handle(new(controllers.PasswordResetController))
	app.Post("/pay-verify", ijwt.JwtGuardMiddleware, middlewares.Protected, controllers.VerifyPayment)

	newsfeedApi := app.Party("/newsfeed", ijwt.JwtGuardMiddleware, middlewares.Protected)
	newsfeedApi.Any("/{action:string}", func(ctx iris.Context) {
		if a := ctx.Params().Get("action"); strings.EqualFold(a, "report-comment") {
			controllers.NewsFeedReportComment(ctx)
			return
		}
		switch ctx.Method() {
		case "POST":
			controllers.NewsFeedInteract(ctx)
		case "GET":
			controllers.NewsFeedView(ctx)
		case "PUT":
			controllers.NewsFeedInteractUpdateComment(ctx)
		default:
			ctx.StatusCode(405)
		}
	})
	mvc.New(newsfeedApi).Handle(new(controllers.NewsFeedController))

	auth := app.Party("/auth")
	auth.Get("/query", controllers.QueryUsernameV1)
	auth.Post("/register", controllers.RegisterV1)
	auth.Post("/login", controllers.LoginV1)
	auth.Post("/provider/{provider:string}", controllers.SocialV1)
	auth.Post("/refresh-token", ijwt.JwtGuardMiddleware, controllers.Refresh)
	auth.Get("/logout", ijwt.JwtGuardMiddleware, middlewares.Protected, controllers.Logout)

	resources := app.Party("/resources")
	resources.Get("/subjects", controllers.GetSubjects)
	resources.Get("/exams", controllers.GetExamTypes)
	resources.Get("/question-types", controllers.GetQuestionTypes)
	resources.Get("/static/{page:string}", controllers.GetStatic)

	support := app.Party("/support")
	faqApi := support.Party("/faq")
	mvc.New(faqApi).Handle(new(controllers.FaqController))

	// User Protected Routes
	user := app.Party("/user", ijwt.JwtGuardMiddleware, middlewares.Protected)
	user.Post("/change-password", controllers.ChangePassword)
	mvc.New(user.Party("/profile")).Handle(new(controllers.AccountController))
	mvc.New(user.Party("/notifications")).Handle(new(controllers.NotificationController))

	userExamApi := user.Party("/exams")
	mvc.New(userExamApi).Handle(new(controllers.UserExamController))

	userSubjectApi := user.Party("/subjects")
	mvc.New(userSubjectApi).Handle(new(controllers.UserSubjectController))

	study := app.Party("/study", ijwt.JwtGuardMiddleware, middlewares.Protected, middlewares.MustRegisterSubject)
	study.Post("/subjects", controllers.StudySubjects)
	study.Post("/topics", middlewares.MustSubscribe, controllers.StudyTopics)
	study.Post("/podcast", middlewares.MustSubscribe, controllers.StudyPodcasts)
	study.Post("/quiz", controllers.QuickQuiz)

	weeklyQuizApiv1 := app.Party("/weekly-quiz",
		ijwt.JwtGuardMiddleware,
		middlewares.Protected,
		middlewares.MustRegisterSubject,
		middlewares.MustSubscribe)
	mvc.New(weeklyQuizApiv1).Handle(new(controllers.WeeklyQuizController))
	weeklyQuizApiv1.Get("/result", controllers.LeaderBoard)

	mock := app.Party("/mock", ijwt.JwtGuardMiddleware, middlewares.Protected, middlewares.MustRegisterSubject, middlewares.MustSubscribe)
	mvc.New(mock).Handle(new(controllers.MockController))

	// Admin Protected Routes
	app.Post("/admin", controllers.AdminLogin)
	app.Post("/admin/refresh-token", ijwt.JwtGuardMiddleware, controllers.Refresh)
	admin := app.Party("/admin", ijwt.JwtGuardMiddleware, middlewares.Protected, middlewares.AdminUser)
	admin.Get("/weekly", middlewares.ResourcePermission, controllers.GetCurrentWeekQuiz)
	admin.Get("/weekly/index", middlewares.ResourcePermission, controllers.IndexWeeklyQuiz)
	admin.Get("/weekly/view", middlewares.ResourcePermission, controllers.ViewWeeklyQuizQuestions)
	admin.Post("/weekly/create", middlewares.ResourcePermission, controllers.CreateWeeklyQuiz)
	admin.Post("/weekly/update", middlewares.ResourcePermission, controllers.UpdateWeeklyQuizQuestion)
	admin.Put("/weekly/update", middlewares.ResourcePermission, controllers.UpdateWeeklyQuiz)
	admin.Delete("/weekly/delete", middlewares.ResourcePermission, controllers.DeleteWeeklyQuizz)

	admin.Get("/mock/index", middlewares.ResourcePermission, controllers.IndexMock)
	admin.Get("/mock/view", middlewares.ResourcePermission, controllers.ViewMockQuestions)
	admin.Post("/mock/create", middlewares.ResourcePermission, controllers.CreateMock)
	admin.Post("/mock/update", middlewares.ResourcePermission, controllers.UpdateMockQuestion)
	admin.Put("/mock/update", middlewares.ResourcePermission, controllers.UpdateMock)
	admin.Delete("/mock/delete", middlewares.ResourcePermission, controllers.DeleteMock)

	admin.Get("/podcast/index", middlewares.ResourcePermission, controllers.IndexPodcast)
	admin.Get("/podcast/view", middlewares.ResourcePermission, controllers.ViewPodcast)
	admin.Post("/podcast/create", middlewares.ResourcePermission, controllers.CreatePodcast)
	admin.Put("/podcast/update", middlewares.ResourcePermission, controllers.UpdatePodcast)
	admin.Delete("/podcast/delete", middlewares.ResourcePermission, controllers.DeletePodcast)

	admin.Get("/newsfeed/index", middlewares.ResourcePermission, controllers.IndexNewsfeed)
	admin.Get("/newsfeed/view", middlewares.ResourcePermission, controllers.ViewNewsfeed)
	admin.Post("/newsfeed/create", middlewares.ResourcePermission, controllers.CreateNewsfeed)
	admin.Put("/newsfeed/update", middlewares.ResourcePermission, controllers.UpdateNewsfeed)
	admin.Delete("/newsfeed/delete", middlewares.ResourcePermission, controllers.DeleteNewsfeed)

	admin.Get("/settings/{setting:string}", controllers.Settings)
	admin.Post("/settings/{setting:string}", controllers.SetSettings)

	adminFaqsApi := admin.Party("/faqs")
	mvc.New(adminFaqsApi).Handle(new(admin_controllers.AdminFaqController))

	mvc.New(admin.Party("/notifications")).Handle(new(admin_controllers.AdminNotificationController))

	mvc.New(admin.Party("/users")).Handle(new(admin_controllers.UserController))
}

func RegisterWebRoutes(app *iris.Application) {
	app.Get("/", func(ctx iris.Context) {
		if err := ctx.ServeFile("./public/index.html"); err != nil {
			ctx.StatusCode(404)
			ctx.JSON(err.Error())
			return
		}
	})
}
