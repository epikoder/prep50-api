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
		Url         string
		Collections collections
	}
	throttle struct {
		Limit int
		Burst int
	}
	collections struct {
		Users string
	}
	aes struct {
		Key string
		Iv  string
	}
)

var (
	__DIR__, path string
	Conf          *conf
)

func init() {
	godotenv.Load()
	__DIR__, err := os.Getwd()
	if !logger.HandleError(err) {
		os.Exit(1)
	}

	path = "config.yml"
	if os.Getenv("APP_ENV") == "local" {
		path = "config.local.yml"
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
