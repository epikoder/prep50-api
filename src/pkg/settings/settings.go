package settings

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/Prep50mobileApp/prep50-api/src/pkg/logger"
	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

var (
	path            string
	generalSettings map[string]interface{} = map[string]interface{}{}
)

func SeedSettings() {
	godotenv.Load()
	__DIR__, err := os.Getwd()
	if !logger.HandleError(err) {
		os.Exit(1)
	}

	path = "settings.yml"
	var file *os.File
	defer func() {
		if file != nil {
			file.Close()
		}
	}()
	if file, err = os.OpenFile(fmt.Sprintf("%s/%s", __DIR__, path),
		os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644); !logger.HandleError(err) {
		panic(err)
	}
	var buf []byte
	if buf, err = ioutil.ReadAll(file); !logger.HandleError(err) {
		panic(err)
	}
	if err = yaml.Unmarshal(buf, &generalSettings); !logger.HandleError(err) {
		panic(err)
	}
}

func Update() {
	var file *os.File
	var err error

	defer func() {
		if file != nil {
			file.Close()
		}
	}()
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
	if buf, err = yaml.Marshal(&generalSettings); !logger.HandleError(err) {
		panic(err)
	}
	ioutil.WriteFile(filePath, buf, os.ModePerm)
}

func Get(k string, d interface{}) (v interface{}) {
	SeedSettings()
	v, ok := generalSettings[k]
	if !ok {
		return d
	}
	return
}

func GetString(k string, d string) (s string) {
	SeedSettings()
	v, ok := generalSettings[k]
	if !ok {
		return d
	}
	return v.(string)
}

func Set(k string, v interface{}) {
	generalSettings[k] = v
	Update()
}
