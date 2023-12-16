package config

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/Prep50mobileApp/prep50-api/src/pkg/logger"
	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

type (
	conf struct {
		App      app
		Database dbConnection
		Throttle throttle
		Aes      aes
		Cors     cors
		Jwt      jwt
		Mail     mail
		Redis    redis
	}
)

// types
type (
	app struct {
		Name              string
		Host              string
		Port              int
		Key               string
		Version           string
		Defaultapiversion string `yaml:"defaultApiVersion"`
	}
	cors struct {
		Origins          []string
		Headers          string
		AllowCredentials string `yaml:"allowCredentials"`
	}
	dbConnection struct {
		Url      string
		Host     string
		Port     string
		User     string
		Password string
		Name     string
	}
	throttle struct {
		Limit int
		Burst int
	}

	aes struct {
		Key string
		Iv  string
	}

	jwt struct {
		Access  int
		Refresh int
	}

	mail struct {
		UserName MailUserName `yaml:"username"`
		From     string
		SmtpHost string `yaml:"smtpHost"`
		SmtpPort int    `yaml:"smtpPort"`
		Password string
	}

	redis struct {
		Host     string
		Port     int
		Password string
		Database int
	}
	MailUserName string
)

var (
	path string
	Conf *conf
)

func (m MailUserName) ToUserName() string {
	v := string(m)
	if !strings.Contains(v, "@") {
		v += "@prep50.ng"
	}
	return v
}

func init() {
	godotenv.Load()
	__DIR__, err := os.Getwd()
	if !logger.HandleError(err) {
		os.Exit(1)
	}

	path = "config.yml"
	if env := os.Getenv("APP_ENV"); env != "" && env != "production" {
		path = "config." + env + ".yml"
	}

	var file *os.File
	if file, err = os.OpenFile(fmt.Sprintf("%s/%s", __DIR__, path),
		os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644); !logger.HandleError(err) {
		panic(err)
	}
	defer file.Close()
	var buf []byte
	if buf, err = io.ReadAll(file); !logger.HandleError(err) {
		panic(err)
	}
	if err = yaml.Unmarshal(buf, &Conf); !logger.HandleError(err) {
		panic(err)
	}

	{

		if len(Conf.Database.Url) == 0 {
			if uri := os.Getenv("DB_URL"); len(uri) > 0 {
				Conf.Database.Url = uri
			} else {
				if len(Conf.Database.Host) == 0 {
					Conf.Database.Host = os.Getenv("DB_HOST")
				}
				if len(Conf.Database.Port) == 0 {
					Conf.Database.Port = os.Getenv("DB_PORT")
				}
				if len(Conf.Database.Name) == 0 {
					Conf.Database.Name = os.Getenv("DB_NAME")
				}
				if len(Conf.Database.User) == 0 {
					Conf.Database.User = os.Getenv("DB_USER")
				}
				if len(Conf.Database.Password) == 0 {
					Conf.Database.Password = os.Getenv("DB_PASSWORD")
				}
			}
		}
	}

	{
		if len(Conf.Redis.Password) == 0 {
			Conf.Redis.Password = os.Getenv("REDIS_PASSWORD")
		}
	}
}

func Update() {
	var file *os.File
	var err error

	__DIR__, err := os.Getwd()
	if !logger.HandleError(err) {
		panic(err)
	}
	filePath := fmt.Sprintf("%s/%s", __DIR__, path)
	fileBakPath := filePath + ".bak"
	if file, err = os.OpenFile(filePath,
		os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644); !logger.HandleError(err) {
		panic(err)
	}
	defer file.Close()
	var buf []byte
	if buf, err = io.ReadAll(file); !logger.HandleError(err) {
		panic(err)
	}
	err = os.WriteFile(fileBakPath, buf, os.ModePerm)
	logger.HandleError(err)
	if buf, err = yaml.Marshal(&Conf); !logger.HandleError(err) {
		panic(err)
	}
	os.WriteFile(filePath, buf, os.ModePerm)
}