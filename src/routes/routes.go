package routes

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/Prep50mobileApp/prep50-api/src/controllers"
	admin_controllers "github.com/Prep50mobileApp/prep50-api/src/controllers/admin"
	"github.com/Prep50mobileApp/prep50-api/src/middlewares"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/ijwt"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
)

func RegisterApiRoutes(app *iris.Application) {
	app.Use(iris.Compression)
	app.AllowMethods(iris.MethodOptions)

	app.HandleDir("storage", iris.Dir("./storage"), iris.DirOptions{Compress: true})

	mvc.New(app.Party("password-reset")).Handle(new(controllers.PasswordResetController))
	app.Get("/deregister-device", controllers.DeregisterDevice)
	app.Post("/pay-verify", ijwt.JwtGuardMiddleware, middlewares.Protected, controllers.VerifyPayment)
	app.Post("/pay-init", ijwt.JwtGuardMiddleware, middlewares.Protected, controllers.InitializePayment)

	newsfeedApi := app.Party("/newsfeed", ijwt.JwtGuardMiddleware, middlewares.Protected)
	newsfeedApi.Any("/{action:string}", func(ctx iris.Context) {
		if a := ctx.Params().Get("action"); strings.EqualFold(a, "report-comment") {
			controllers.NewsFeedReportComment(ctx)
			return
		}
		switch ctx.Method() {
		case "POST":
			controllers.NewsFeedInteract(ctx)
			return
		case "GET":
			controllers.NewsFeedView(ctx)
			return
		case "PUT":
			controllers.NewsFeedInteractUpdateComment(ctx)
			return
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
	support.Get("/contacts", controllers.Contacts)
	app.Get("/terms", controllers.Terms)
	app.Get("/privacy", controllers.Privacy)

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
	study.Post("/topics", middlewares.MustSubscribe, controllers.StudyLessons)
	study.Get("/podcast", middlewares.MustSubscribe, controllers.StudyPodcasts)
	study.Get("/quiz", controllers.QuickQuiz)
	study.Post("/quiz", controllers.QuickQuizScore)

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
	app.Post("/admin", admin_controllers.AdminLogin)
	app.Post("/admin/refresh-token", ijwt.JwtGuardMiddleware, controllers.Refresh)
	admin := app.Party("/admin", ijwt.JwtGuardMiddleware, middlewares.Protected, middlewares.AdminUser)

	admin.Get("/weekly", middlewares.ResourcePermission, admin_controllers.GetCurrentWeekQuiz)
	admin.Get("/weekly/index", middlewares.ResourcePermission, admin_controllers.IndexWeeklyQuiz)
	admin.Get("/weekly/view", middlewares.ResourcePermission, admin_controllers.ViewWeeklyQuizQuestions)
	admin.Post("/weekly/create", middlewares.ResourcePermission, admin_controllers.CreateWeeklyQuiz)
	admin.Post("/weekly/update", middlewares.ResourcePermission, admin_controllers.UpdateWeeklyQuizQuestion)
	admin.Put("/weekly/update", middlewares.ResourcePermission, admin_controllers.UpdateWeeklyQuiz)
	admin.Delete("/weekly/delete", middlewares.ResourcePermission, admin_controllers.DeleteWeeklyQuizz)

	admin.Get("/mock/index", middlewares.ResourcePermission, admin_controllers.IndexMock)
	admin.Get("/mock/view", middlewares.ResourcePermission, admin_controllers.ViewMockQuestions)
	admin.Post("/mock/create", middlewares.ResourcePermission, admin_controllers.CreateMock)
	admin.Post("/mock/update", middlewares.ResourcePermission, admin_controllers.UpdateMockQuestion)
	admin.Put("/mock/update", middlewares.ResourcePermission, admin_controllers.UpdateMock)
	admin.Delete("/mock/delete", middlewares.ResourcePermission, admin_controllers.DeleteMock)

	admin.Get("/podcast/index", middlewares.ResourcePermission, admin_controllers.IndexPodcast)
	admin.Get("/podcast/view", middlewares.ResourcePermission, admin_controllers.ViewPodcast)
	admin.Post("/podcast/create", middlewares.ResourcePermission, admin_controllers.CreatePodcast)
	admin.Put("/podcast/update", middlewares.ResourcePermission, admin_controllers.UpdatePodcast)
	admin.Delete("/podcast/delete", middlewares.ResourcePermission, admin_controllers.DeletePodcast)

	admin.Get("/newsfeed/index", middlewares.ResourcePermission, admin_controllers.IndexNewsfeed)
	admin.Get("/newsfeed/view", middlewares.ResourcePermission, admin_controllers.ViewNewsfeed)
	admin.Post("/newsfeed/create", middlewares.ResourcePermission, admin_controllers.CreateNewsfeed)
	admin.Put("/newsfeed/update", middlewares.ResourcePermission, admin_controllers.UpdateNewsfeed)
	admin.Delete("/newsfeed/delete", middlewares.ResourcePermission, admin_controllers.DeleteNewsfeed)

	admin.Get("/settings/{setting:string}", admin_controllers.Settings)
	admin.Post("/settings/{setting:string}", admin_controllers.SetSettings)

	adminFaqsApi := admin.Party("/faqs")
	mvc.New(adminFaqsApi).Handle(new(admin_controllers.AdminFaqController))

	mvc.New(admin.Party("/notifications")).Handle(new(admin_controllers.AdminNotificationController))

	mvc.New(admin.Party("/users")).Handle(new(admin_controllers.UserController))
}

func RegisterWebRoutes(app *iris.Application) {
	app.Get("/", func(ctx iris.Context) {
		if err := ctx.ServeFile("./public/index.html"); err != nil {
			ctx.StatusCode(404)
			return
		}
	})
	app.Get("/{path:path}", func(ctx iris.Context) {
		path := ctx.Params().Get("path")
		path = filepath.Clean(fmt.Sprintf("./public/%s", path))
	SERVE_FILE:
		if err := ctx.ServeFile(path); err != nil {
			if !strings.Contains(path, ".html") {
				path = path + ".html"
				goto SERVE_FILE
			}
			ctx.StatusCode(404)
			return
		}
	})
}
