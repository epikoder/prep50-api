package config

import (
	"fmt"
	"io/ioutil"
	"os"

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
		UserName string `yaml:"username"`
		From     string
		SmtpHost string `yaml:"smtpHost"`
		SmtpPort int    `yaml:"smtpPort"`
		Password string
	}
)

var (
	path string
	Conf *conf
)

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
	if env := os.Getenv("APP_ENV"); env == "" || env != "production" {
		path = "config." + env + ".yml"
	}

	var file *os.File
	if file, err = os.OpenFile(fmt.Sprintf("%s/%s", __DIR__, path),
		os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644); !logger.HandleError(err) {
		panic(err)
	}
	var buf []byte
	if buf, err = ioutil.ReadAll(file); !logger.HandleError(err) {
		panic(err)
	}
	if err = yaml.Unmarshal(buf, &Conf); !logger.HandleError(err) {
		panic(err)
	}
}

func Save() {
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
	var buf []byte
	if buf, err = ioutil.ReadAll(file); !logger.HandleError(err) {
		panic(err)
	}
	err = ioutil.WriteFile(fileBakPath, buf, os.ModePerm)
	logger.HandleError(err)
	if buf, err = yaml.Marshal(&Conf); !logger.HandleError(err) {
		panic(err)
	}
	ioutil.WriteFile(filePath, buf, os.ModePerm)
}
