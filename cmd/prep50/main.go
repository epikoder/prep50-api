package main

import (
	"fmt"
	"os"

	"github.com/Prep50mobileApp/prep50-api/config"
	"github.com/Prep50mobileApp/prep50-api/src/middlewares"
	"github.com/Prep50mobileApp/prep50-api/src/routes"
	"github.com/Prep50mobileApp/prep50-api/src/services/database"
	"github.com/go-playground/validator/v10"
	"github.com/kataras/iris/v12"
	"gorm.io/gorm"
)

type Prep50 struct {
	App *iris.Application
	DB  *gorm.DB
}

func main() {
	prep50 := &Prep50{iris.New(), database.DBConnection}
	prep50.registerAppRoutes()
	prep50.registerMiddlewares()
	prep50.RegisterStructValidation()
	prep50.StartServer()
}

func (prep50 *Prep50) RegisterStructValidation() {
	v := validator.New()
	prep50.App.Validator = v
}

func (prep50 *Prep50) StartServer() {
	{
		if os.Getenv("APP_ENV") != "local" {
			prep50.App.Run(iris.TLS(fmt.Sprintf("%s:443",
				config.Conf.App.Host),
				"server.crt",
				"server.key"),
				iris.WithConfiguration(iris.YAML("server.yml")))
			return
		}
	}

	addr := func() string {
		if h := config.Conf.App.Host; h != "" {
			return fmt.Sprintf("%s:%d", h, config.Conf.App.Port)
		}
		return fmt.Sprintf(":%d", config.Conf.App.Port)
	}
	prep50.App.Run(iris.Addr(addr()), iris.WithConfiguration(iris.YAML("server.yml")))
}

func (prep50 *Prep50) registerMiddlewares() {
	prep50.App.UseGlobal(middlewares.CORS)
}

func (prep50 *Prep50) registerAppRoutes() {
	routes.RegisterApiRoutes(prep50.App)
}
