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
		Database database
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
	database struct {
		Core dbConnection
		App  dbConnection
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

func (db database) UseDB(s string) dbConnection {
	if s == "core" {
		return db.Core
	}
	return db.App
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

	if len(Conf.Database.App.Url) == 0 {
		if len(Conf.Database.App.Password) == 0 {
			Conf.Database.App.Password = os.Getenv("DB_APP_PASSWORD")
		}
	}

	if len(Conf.Database.Core.Url) == 0 {
		if len(Conf.Database.Core.Password) == 0 {
			Conf.Database.Core.Password = os.Getenv("DB_CORE_PASSWORD")
		}
	}

	if len(Conf.Redis.Password) == 0 {
		Conf.Redis.Password = os.Getenv("REDIS_PASSWORD")
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
