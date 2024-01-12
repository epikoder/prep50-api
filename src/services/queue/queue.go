package queue

import (
	"fmt"
	"reflect"
	"runtime"
	"time"

	"github.com/Prep50mobileApp/prep50-api/src/pkg/logger"
)

type (
	JobType string
	Job     struct {
		Type     JobType
		Func     func() error
		Schedule time.Time
		Retries  int
	}
	queue chan Job
)

var (
	_queue   queue   = make(queue, 1024)
	SendMail JobType = "mail"
	Action   JobType = "action"
	NilJob   JobType = "nil"
)

func Dispatch(j Job) {
	fmt.Println("Scheduling Job: ", j.Type)
	select {
	case _queue <- j:
		fmt.Println("Scheduled Job: ", j.Type, runtime.FuncForPC(reflect.ValueOf(j.Func).Pointer()).Name())
	default:
		fmt.Println("Failed to Schedule Job: ", j.Type)
	}
}

func Run() {
	for j := range _queue {
		go func(j Job) {
			if err := j.Func(); err != nil {
				logger.HandleError(err)
			}
		}(j)
	}

}

func Len() int {
	return len(_queue)
}
