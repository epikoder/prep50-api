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
	"github.com/Prep50mobileApp/prep50-api/src/pkg/settings"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/validation"
	"github.com/Prep50mobileApp/prep50-api/src/services/database"
	"github.com/kataras/iris/v12"
)

type (
	WeeklyQuizController struct {
		Ctx iris.Context
	}
	Answer map[uint]string
)

func (a Answer) MarshalBinary() (data []byte, err error) {
	return json.Marshal(a)
}

func (a Answer) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, &a)
}

func (c *WeeklyQuizController) Get() {
	type WQ struct {
		models.WeeklyQuiz
		Started   bool                            `json:"started"`
		Questions []models.QuestionsWithoutAnswer `json:"questions"`
	}
	quiz := &models.WeeklyQuiz{}
	_, w := time.Now().ISOWeek()
	if ok := repository.NewRepository(quiz).FindOne("week", w); !ok {
		c.Ctx.JSON(apiResponse{
			"status":  "failed",
			"message": "no quizz avaliable for current week",
		})
		return
	}

	var questions []models.QuestionsWithoutAnswer = []models.QuestionsWithoutAnswer{}
	started := quiz.StartTime.After(time.Now())
	if env := os.Getenv("APP_ENV"); env != "" && env != "production" {
		started = true
	}
	if started {
		var err error
		if questions, err = quiz.QuestionsWithoutAnswer(); err != nil {
			c.Ctx.StatusCode(http.StatusInternalServerError)
			c.Ctx.JSON(internalServerError)
			return
		}
	}

	c.Ctx.JSON(apiResponse{
		"status": "success",
		"data":   WQ{*quiz, started, models.RandomizeQuestionWithoutAnswer(questions)},
	})
}

func (c *WeeklyQuizController) Post() {
	var userAnswer Answer
	if err := c.Ctx.ReadJSON(&userAnswer); err != nil {
		c.Ctx.StatusCode(400)
		c.Ctx.JSON(validation.Errors(err))
		return
	}

	answers := Answer{}
	year, week := time.Now().ISOWeek()
	session := settings.Get("examSession", year)
	quiz := &models.WeeklyQuiz{}
	if ok := repository.NewRepository(quiz).FindOne("week = ? AND session = ?", week, session); !ok {
		c.Ctx.JSON(apiResponse{
			"status":  "failed",
			"message": "No quiz available for this week",
		})
		return
	}

	if env := os.Getenv("APP_ENV"); env != "" && env != "production" {
		goto SKIP
	}
	if quiz.StartTime.After(time.Now()) {
		c.Ctx.JSON(apiResponse{
			"status":  "failed",
			"message": "Quiz is not running",
		})
		return
	}

SKIP:
	user, _ := getUser(c.Ctx)
	_time := time.Now()
	key := fmt.Sprintf("weekly.quiz.answer.%d.%d", session, week)
	ans, ok := cache.Get(key)
	if err := answers.UnmarshalBinary([]byte(ans)); !ok || err != nil {
		questions, err := quiz.Questions()
		if err != nil {
			c.Ctx.StatusCode(500)
			c.Ctx.JSON(internalServerError)
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
	{
		result := &models.WeeklyQuizResult{}
		if err := database.UseDB("app").First(result, "user_id = ? AND weekly_quiz_id = ?", user.Id, quiz.Id).Error; err == nil {
			c.Ctx.JSON(apiResponse{
				"Status":  "success",
				"message": "Congratulations on completing the weekly quiz",
				"data": apiResponse{
					"score":  result.Score,
					"quiz":   quiz,
					"answer": answers,
				},
			})
			return
		}
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

	if err := database.UseDB("app").Create(models.WeeklyQuizResult{
		WeeklyQuizId: quiz.Id,
		UserId:       user.Id,
		Score:        score,
		Duration:     uint(quiz.StartTime.Sub(_time).Minutes()),
	}).Error; err != nil {
		c.Ctx.StatusCode(500)
		c.Ctx.JSON(internalServerError)
		return
	}

	c.Ctx.JSON(apiResponse{
		"Status":  "success",
		"message": "Congratulations on completing the weekly quiz",
		"data": apiResponse{
			"score":   score,
			"quiz":    quiz,
			"answers": answers,
		},
	})
}

type (
	ResultQuery struct {
		Week    uint
		Session uint
	}
	Result struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Photo    string `json:"photo"`
		Score    uint   `json:"score"`
	}
)

func LeaderBoard(ctx iris.Context) {
	query := &ResultQuery{}
	ctx.ReadBody(query)

	year, week := time.Now().ISOWeek()
	session := settings.Get("examSession", year)
	if query.Week == 0 {
		query.Week = uint(week)
	}
	if query.Session == 0 {
		query.Session = (uint)(session.(int))
	}
	quiz := &models.WeeklyQuiz{}
	if ok := repository.NewRepository(quiz).FindOne("week = ? AND session = ?", query.Week, query.Session); !ok {
		ctx.JSON(apiResponse{
			"status":  "failed",
			"message": "No quiz available for this week",
		})
		return
	}

	results := []Result{}
	if err := database.UseDB("app").
		Order("score DESC").
		Table("weekly_quiz_results as wr").
		Select("wr.score, u.username, u.email, u.photo").
		Joins("LEFT JOIN users as u ON u.id = wr.user_id").
		Find(&results, "weekly_quiz_id = ?", quiz.Id).Error; err != nil {
		ctx.StatusCode(500)
		return
	}
	ctx.JSON(apiResponse{
		"status": "success",
		"data":   results,
	})
}
