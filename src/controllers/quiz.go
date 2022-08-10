package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/Prep50mobileApp/prep50-api/src/models"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/cache"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/logger"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/repository"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/validation"
	"github.com/kataras/iris/v12"
)

type (
	Answer map[uint]string
)

func (a Answer) MarshalBinary() (data []byte, err error) {
	return json.Marshal(a)
}

func (a Answer) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, &a)
}

func WeekQuiz(ctx iris.Context) {
	type WQ struct {
		models.WeeklyQuiz
		Started   bool              `json:"started"`
		Questions []models.Question `json:"questions"`
	}
	quiz := &models.WeeklyQuiz{}
	_, w := time.Now().ISOWeek()
	if ok := repository.NewRepository(quiz).FindOne("week", w); !ok {
		ctx.JSON(apiResponse{
			"status":  "failed",
			"message": "no quizz avaliable for current week",
		})
		return
	}

	var questions []models.Question = []models.Question{}
	started := quiz.StartTime.After(time.Now())
	if env := os.Getenv("APP_ENV"); env != "" && env != "production" {
		started = true
	}
	if started {
		var err error
		if questions, err = quiz.Questions(); err != nil {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(internalServerError)
			return
		}
	}

	ctx.JSON(apiResponse{
		"status": "success",
		"data":   WQ{*quiz, started, models.RandomizeQuestions(questions)},
	})
}

func WeekUserScore(ctx iris.Context) {
	var userAnswer Answer = Answer{}
	if err := ctx.ReadJSON(&userAnswer); err != nil {
		ctx.StatusCode(400)
		ctx.JSON(validation.Errors(err))
		return
	}

	answers := Answer{}
	_, w := time.Now().ISOWeek()
	quiz := &models.WeeklyQuiz{}
	if ok := repository.NewRepository(quiz).FindOne("week = ?", w); !ok {
		ctx.JSON(apiResponse{
			"status":  "failed",
			"message": "No quiz available for this week",
		})
		return
	}
	if env := os.Getenv("APP_ENV"); env != "" && env != "production" {
		goto SKIP
	}
	if quiz.StartTime.After(time.Now()) {
		ctx.JSON(apiResponse{
			"status":  "failed",
			"message": "Quiz is not running",
		})
		return
	}

SKIP:
	key := fmt.Sprintf("weekly.quiz.answer.%d", w)
	ans, ok := cache.Get(key)
	if err := answers.UnmarshalBinary([]byte(ans)); !ok || err != nil {

		questions, err := quiz.Questions()
		if err != nil {
			ctx.StatusCode(500)
			ctx.JSON(internalServerError)
			return
		}

		for _, question := range questions {
			if question.QuestionTypeId == uint(models.OBJECTIVE) {
				answers[question.Id] = question.ShortAnswer
			} else {
				answers[question.Id] = question.FullAnswer
			}
		}
		logger.HandleError(cache.Set(key, answers, cache.Duration(time.Now().Add(time.Hour*72).Unix())))
	}

	var score uint = 0
	for id, ans := range answers {
		ua, ok := userAnswer[id]
		if !ok {
			continue
		}
		if ua == ans {
			score++
		}
	}

	// TODO : save user score
	ctx.JSON(apiResponse{
		"Status":  "success",
		"message": "Congratulations on completing the weekly quiz",
		"data": apiResponse{
			"score": 0,
			"quiz":  quiz,
		},
	})
}

func WeekLeaderBoard(ctx iris.Context) {

	cache.Get("weekly.quiz.scores")
}
