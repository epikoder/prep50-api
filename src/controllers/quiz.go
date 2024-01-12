package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/Prep50mobileApp/prep50-api/src/models"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/cache"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/config"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/list"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/logger"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/repository"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/settings"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/validation"
	"github.com/Prep50mobileApp/prep50-api/src/services/database"
	"github.com/kataras/iris/v12"
	"gorm.io/gorm"
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
		Started   bool              `json:"started"`
		Questions []models.Question `json:"questions"`
	}
	quiz := &models.WeeklyQuiz{}
	year, week := time.Now().ISOWeek()
	session := settings.Get("exam.session", year)
	if ok := repository.NewRepository(quiz).FindOne("week = ? AND session = ?", week, session); !ok {
		c.Ctx.JSON(apiResponse{
			"status":  "failed",
			"message": "No quizz avaliable for current week",
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
		if questions, err = quiz.QuestionsWithAnswer(); !logger.HandleError(err) {
			c.Ctx.StatusCode(http.StatusInternalServerError)
			c.Ctx.JSON(internalServerError)
			return
		}
	}

	c.Ctx.JSON(apiResponse{
		"status": "success",
		"data":   WQ{*quiz, started, list.Shuffle(questions)},
	})
}

func (c *WeeklyQuizController) Post() {
	var userAnswer Answer
	if err := c.Ctx.ReadJSON(&userAnswer); !logger.HandleError(err) {
		c.Ctx.StatusCode(400)
		c.Ctx.JSON(validation.Errors(err))
		return
	}

	answers := Answer{}
	year, week := time.Now().ISOWeek()
	session := settings.Get("exam.session", year)
	quiz := &models.WeeklyQuiz{}
	if ok := repository.NewRepository(quiz).Preload("Questions", func(db *gorm.DB) *gorm.DB {
		return db.Table(fmt.Sprintf("%s.questions", config.Conf.Database.Name))
	}).FindOne("week = ? AND session = ?", week, session); !ok {
		c.Ctx.JSON(apiResponse{
			"status":  "failed",
			"message": "No quiz available for this week",
		})
		return
	}

	if env := os.Getenv("APP_ENV"); env != "" && env != "production" {
		goto SKIP
	}
	if quiz.StartTime.Before(time.Now()) || quiz.StartTime.Add(time.Minute*time.Duration(quiz.Duration+10)).Before(time.Now()) {
		c.Ctx.JSON(apiResponse{
			"status":  "failed",
			"message": "Quiz is not active",
		})
		return
	}

SKIP:
	user, _ := getUser(c.Ctx)
	_time := time.Now()
	key := fmt.Sprintf("weekly.quiz.answer.%d.%d", session, week)
	ans, ok := cache.Get(key)
	if err := answers.UnmarshalBinary([]byte(ans)); !ok || !logger.HandleError(err) {
		questions := quiz.Questions

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
		if err := database.DB().First(result, "user_id = ? AND weekly_quiz_id = ?", user.Id, quiz.Id).Error; err == nil {
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
		if ua, ok := userAnswer[id]; ok && ua == ans {
			score++
		}
	}

	if err := database.DB().Create(models.WeeklyQuizResult{
		WeeklyQuizId: quiz.Id,
		UserId:       user.Id,
		Score:        score,
		Duration:     uint(quiz.StartTime.Sub(_time).Minutes()),
	}).Error; !logger.HandleError(err) {
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
	session := settings.Get("exam.session", year)
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
	if err := database.DB().
		Order("score DESC").
		Table("weekly_quiz_results as wr").
		Select("wr.score, u.username, u.email, u.photo").
		Joins("LEFT JOIN users as u ON u.id = wr.user_id").
		Find(&results, "weekly_quiz_id = ?", quiz.Id).Error; !logger.HandleError(err) {
		ctx.StatusCode(500)
		return
	}
	ctx.JSON(apiResponse{
		"status": "success",
		"data":   results,
	})
}
