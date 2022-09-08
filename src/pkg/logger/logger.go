package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"strings"
	"time"
)

func HandleError(err error) (ok bool) {
	if os.Getenv("APP_ENV") == "local" || os.Getenv("LOG_STACK") == "file" {
		__DIR__, _ := os.Getwd()
		if err := os.Mkdir(__DIR__+"/logs", 0744); err != nil && !strings.Contains(err.Error(), "file exists") {
			panic(err)
		}
		f, err := os.OpenFile(__DIR__+fmt.Sprintf("/logs/app.%d-%d-%d.log", ((func() []interface{} {
			y, m, d := time.Now().Date()
			return []interface{}{d, int(m), y}
		})())...), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0655)
		if err != nil {
			log.Fatalf("error opening file: %v", err)
		}
		defer f.Close()
		wrt := io.MultiWriter(os.Stdout, f)
		log.SetOutput(wrt)
	}

	if err != nil {
		pc, fn, line, _ := runtime.Caller(1)
		log.Default().Printf("ERROR: %s\n[\n	%s:%d \n	%v\n]", runtime.FuncForPC(pc).Name(), fn, line, err)
		return false
	}
	return true
}
