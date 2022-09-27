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
		hasRun   bool
	}
	queue chan Job
)

var (
	_queue   queue   = make(queue, 1024)
	SendMail JobType = "mail"
	Action   JobType = "action"
	NilJob   JobType = "nil"

	defaultRetries = 3
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
	for {
		select {
		case j := <-_queue:
			{
				if j.Retries == 0 && !j.hasRun {
					j.Retries = defaultRetries
				}
				runFunc := func() {
					j.hasRun = true
					err := j.Func()
					if !logger.HandleError(err) {
						if j.Retries == 1 {
							// TODO: save to db
							fmt.Println("retries exceeded save to db")
							return
						}
						j.Retries--
						j.Schedule = time.Now().Add(time.Second * 5)
						Dispatch(j)
					}
					fmt.Println("Completed Job: ", j.Type, runtime.FuncForPC(reflect.ValueOf(j.Func).Pointer()).Name())
				}
				if time.Since(j.Schedule).Milliseconds() < 0 {
					go func() {
						time.Sleep(time.Duration(-time.Since(j.Schedule).Milliseconds() * int64(time.Millisecond)))
						runFunc()
					}()
					continue
				}
				runFunc()
			}
		default:
			time.Sleep(time.Millisecond * 500)
		}
	}

}

func Len() int {
	return len(_queue)
}

func SetDefaultRetries(i int) {
	defaultRetries = i
}
