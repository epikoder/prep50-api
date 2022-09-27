package helper

import (
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"os"
	"strings"

	"github.com/Prep50mobileApp/prep50-api/config"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/list"
	"github.com/h2non/bimg"
	"github.com/kataras/iris/v12"
)

func SaveTempImage(f multipart.File) (s string, err error) {
	defer f.Close()
	buf, err := ioutil.ReadAll(f)
	if err != nil {
		return
	}
	file, err := ConvertImage(buf, "")
	if err != nil {
		return
	}
	defer file.Close()
	arr := strings.Split(file.Name(), "/")
	s = arr[len(arr)-1]
	return
}

func SaveTempVideo(f multipart.File) (s string, err error) {
	defer f.Close()
	buf, err := ioutil.ReadAll(f)
	if err != nil {
		return
	}
	file, err := os.CreateTemp(os.TempDir(), "*.mp4")
	if err != nil {
		return
	}
	err = ioutil.WriteFile(file.Name(), buf, os.ModeAppend)
	if err != nil {
		return
	}
	defer file.Close()
	arr := strings.Split(file.Name(), "/")
	s = arr[len(arr)-1]
	return
}

func CopyTempFile(path, name string) error {
	f, err := os.OpenFile(fmt.Sprintf("%s/%s", os.TempDir(), name), os.O_APPEND|os.O_RDWR, os.ModeAppend)
	if err != nil {
		return err
	}
	defer func() {
		f.Close()
		os.Remove(f.Name())
	}()
	var buf []byte
	buf, err = ioutil.ReadAll(f)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, buf, os.ModePerm)
}

func CopyTempFilePreserve(path, name string) (f *os.File, err error) {
	f, err = os.OpenFile(fmt.Sprintf("%s/%s", os.TempDir(), name), os.O_APPEND|os.O_RDWR, os.ModeAppend)
	if err != nil {
		return
	}
	var buf []byte
	buf, err = ioutil.ReadAll(f)
	if err != nil {
		return
	}
	return f, ioutil.WriteFile(path, buf, os.ModePerm)
}

func ConvertImage(buf []byte, prefix string) (file *os.File, err error) {
	file, err = os.CreateTemp(os.TempDir(), prefix+"*.webp")
	if err != nil {
		return
	}
	buf, err = bimg.NewImage(buf).Convert(bimg.WEBP)
	if err != nil {
		return
	}
	err = ioutil.WriteFile(file.Name(), buf, os.ModeAppend)
	if err != nil {
		return
	}
	return
}

func GetOrigin(ctx iris.Context) string {
	origin := func() string {
		i := config.Conf.App.Host
		isLocal := func() bool {
			en := os.Getenv("APP_ENV")
			return en != "" && en != "production"
		}()
		protocol := func() string {
			if isLocal {
				return "http://"
			}
			return "https://"
		}()

		ref := func() string {
			arr := strings.Split(ctx.Request().Referer(), "//")
			if len(arr) < 2 {
				return config.Conf.App.Host
			}
			arr2 := strings.Split(arr[1], "/")
			if len(arr2) < 2 {
				return strings.Trim(arr[1], "/")
			}
			return strings.Trim(arr2[0], "/")
		}()
		if list.Contains(config.Conf.Cors.Origins, ref) {
			return fmt.Sprintf("%s%s", protocol, ref)
		}
		return func() string {
			if isLocal {
				return fmt.Sprintf("%s%s:%d", protocol, i, config.Conf.App.Port)
			}
			return fmt.Sprintf("%s%s", protocol, i)
		}()
	}
	return origin()
}
