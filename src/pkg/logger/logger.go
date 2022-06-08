package logger

import (
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"runtime"
	"strings"
)

func HandleError(err error) (ok bool) {
	if os.Getenv("APP_ENV") == "local" || os.Getenv("LOG_STACK") == "file" {
		__DIR__, _ := os.Getwd()
		if err := os.Mkdir(__DIR__+"/logs", fs.ModeDir); !strings.Contains(err.Error(), "file exists") {
			panic(err)
		}
		f, err := os.OpenFile(__DIR__+"/logs/log.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("error opening file: %v", err)
		}
		defer f.Close()
		wrt := io.MultiWriter(os.Stdout, f)
		log.SetOutput(wrt)
	}

	if err != nil {
		pc, fn, line, _ := runtime.Caller(1)
		fmt.Printf("ERROR in %s \n[ %s:%d ] \n%v", runtime.FuncForPC(pc).Name(), fn, line, err)
		return false
	}
	return true
}
