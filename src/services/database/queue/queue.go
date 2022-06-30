package queue

import (
	"fmt"
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
	_queue   queue   = make(queue)
	SendMail JobType = "mail"
	NilJob   JobType = "nil"

	defaultRetries = 3
)

func Dispatch(j Job) {
	select {
	case _queue <- j:
	default:
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
					fmt.Println(err)
					if !logger.HandleError(err) {
						fmt.Println(err)
						if j.Retries == 1 {
							fmt.Println("retries exceeded save to db")
							return
						}
						fmt.Println("retrying failed job")
						j.Retries--
						j.Schedule = time.Now().Add(time.Second * 5)
						Dispatch(j)
					}
					fmt.Println("success")
				}
				fmt.Println("start")
				if time.Since(j.Schedule).Milliseconds() < 0 {
					fmt.Println("is scheduled")
					go func() {
						time.Sleep(time.Duration(-time.Since(j.Schedule).Milliseconds() * int64(time.Millisecond)))
						runFunc()
					}()
					continue
				}
				fmt.Println("is not scheduled")
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
