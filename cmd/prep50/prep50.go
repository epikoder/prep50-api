package prep50

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/Prep50mobileApp/prep50-api/src/middlewares"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/config"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/crypto"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/ijwt"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/logger"
	"github.com/Prep50mobileApp/prep50-api/src/routes"
	"github.com/go-playground/validator/v10"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/middleware/accesslog"
)

type Prep50 struct {
	App *iris.Application
}

func NewApp() *Prep50 {
	return &Prep50{iris.New()}
}

func (prep50 *Prep50) RegisterStructValidation() {
	v := validator.New()
	prep50.App.Validator = v
}

func (prep50 *Prep50) StartServer() {
	serverConfigPath := "server.yml"
	{
		if env := os.Getenv("APP_ENV"); env != "production" || env == "" {
			port, err := strconv.Atoi(os.Getenv("PORT"))
			if err != nil && config.Conf.App.Port != 0 {
				port = config.Conf.App.Port
			}
			addr := func() string {
				return fmt.Sprintf(":%d", port)
			}
			prep50.App.Run(iris.Addr(addr()), iris.WithConfiguration(iris.YAML(serverConfigPath)))
			return
		}
	}

	prep50.App.Run(iris.TLS(":443",
		"server.crt",
		"server.key"),
		iris.WithConfiguration(iris.YAML(serverConfigPath)))
}

func (prep50 *Prep50) RegisterMiddlewares() {
	prep50.App.UseGlobal(middlewares.CORS)
	prep50.App.UseGlobal(middlewares.Security)
}

func (prep50 *Prep50) RegisterAppRoutes() {
	routes.RegisterApiRoutes(prep50.App)
	routes.RegisterWebRoutes(prep50.App)
}

func (prep50 *Prep50) AuthConfig() {
	if d := config.Conf.Jwt.Access; d != 0 {
		ijwt.SetAccessLife(d)
	}
	if d := config.Conf.Jwt.Refresh; d != 0 {
		ijwt.SetRefreshLife(d)
	}
}

func (prep50 *Prep50) UseEncryption() {
	if config.Conf.Aes.Key == "" || config.Conf.Aes.Iv == "" {
		config.Conf.Aes.Key = crypto.Base64(32)
		config.Conf.Aes.Iv = crypto.Base64(16)
		config.Update()
	}
}

func (prep50 *Prep50) RegisterViews() {
	prep50.App.RegisterView(iris.Amber("templates/views", ".amber"))
	prep50.App.HandleDir("/static", "public/assets", iris.DirOptions{Compress: true})
}

func (prep50 *Prep50) StartLogging() {
	var ac *accesslog.AccessLog
	var output io.Writer
	if os.Getenv("APP_ENV") == "local" || os.Getenv("LOG_STACK") == "file" {
		os.Mkdir("logs", os.ModeDir|os.ModePerm)
		f, err := os.OpenFile(fmt.Sprintf("logs/access-%d-%d-%d.log", ((func() []interface{} {
			y, m, d := time.Now().Date()
			return []interface{}{d, int(m), y}
		})())...), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0655)
		if !logger.HandleError(err) {
			return
		}
		output = io.MultiWriter(f)
	} else {
		output = io.MultiWriter()
	}

	ac = accesslog.New(output)
	ac.Delim = '|'
	ac.TimeFormat = "2006-01-02 15:04:05"
	ac.Async = false
	ac.IP = true
	ac.BytesReceivedBody = true
	ac.BytesSentBody = true
	ac.BytesReceived = false
	ac.BytesSent = false
	ac.BodyMinify = true
	ac.RequestBody = true
	ac.ResponseBody = false
	ac.KeepMultiLineError = true
	ac.PanicLog = accesslog.LogHandler
	ac.SetFormatter(&accesslog.JSON{
		Indent:    "  ",
		HumanTime: true,
	})
	prep50.App.UseGlobal(ac.Handler)
}
