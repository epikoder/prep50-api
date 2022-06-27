package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/Prep50mobileApp/prep50-api/config"
	"github.com/Prep50mobileApp/prep50-api/src/middlewares"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/ijwt"
	"github.com/Prep50mobileApp/prep50-api/src/routes"
	"github.com/Prep50mobileApp/prep50-api/src/services/database/queue"
	"github.com/go-playground/validator/v10"
	"github.com/kataras/iris/v12"
)

type Prep50 struct {
	App *iris.Application
}

func main() {
	prep50 := &Prep50{iris.New()}
	prep50.registerAppRoutes()
	prep50.registerMiddlewares()
	prep50.RegisterStructValidation()
	prep50.AuthConfig()
	go queue.Run()
	prep50.StartServer()
}

func (prep50 *Prep50) RegisterStructValidation() {
	v := validator.New()
	prep50.App.Validator = v
}

func (prep50 *Prep50) StartServer() {
	serverConfigPath := "server.yml"
	{
		if env := os.Getenv("APP_ENV"); env == "local" || env == "devel" {
			port, err := strconv.Atoi(os.Getenv("PORT"))
			fmt.Println(err != nil, config.Conf.App.Port != 0)
			if err != nil && config.Conf.App.Port != 0 {
				port = config.Conf.App.Port
			}
			addr := func() string {
				if h := config.Conf.App.Host; h != "" {

					return fmt.Sprintf("%s:%d", h, port)
				}
				return fmt.Sprintf(":%d", port)
			}
			prep50.App.Run(iris.Addr(addr()), iris.WithConfiguration(iris.YAML(serverConfigPath)))
			return
		}
	}

	prep50.App.Run(iris.TLS(fmt.Sprintf("%s:443",
		config.Conf.App.Host),
		"server.crt",
		"server.key"),
		iris.WithConfiguration(iris.YAML(serverConfigPath)))
}

func (prep50 *Prep50) registerMiddlewares() {
	prep50.App.UseGlobal(middlewares.CORS)
}

func (prep50 *Prep50) registerAppRoutes() {
	routes.RegisterApiRoutes(prep50.App)
}

func (prep50 *Prep50) AuthConfig() {
	if d := config.Conf.Jwt.Access; d != 0 {
		ijwt.SetAccessLife(d)
	}
	if d := config.Conf.Jwt.Refresh; d != 0 {
		ijwt.SetRefreshLife(d)
	}
}
