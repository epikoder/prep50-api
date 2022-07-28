package models

import (
	"time"

	"github.com/Prep50mobileApp/prep50-api/src/pkg/dbmodel"
	"github.com/Prep50mobileApp/prep50-api/src/services/database"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	WeeklyQuiz struct {
		Id            uuid.UUID `sql:"primary_key;unique;type:uuid;default:uuid_generate_v4()" gorm:"type:varchar(36);index;" json:"-"`
		Prize         uint      `json:"prize"`
		Message       string    `json:"message"`
		QuestionCount uint      `json:"question_count"`
		Duration      uint      `json:"duration"`
		Completed     bool      `json:"completed"`
		StartTime     time.Time `json:"start_time"`
		CreatedAt     time.Time `json:"-"`
		UpdatedAt     time.Time `json:"-"`
		CreatedBy     string    `json:"-"`
	}

	WeeklyQuestion struct {
		QuizId     uuid.UUID `gorm:"type:varchar(36);index;"`
		QuestionId uint      `gorm:"index;"`
		CreatedBy  string
	}

	WeeklyQuizFormStruct struct {
		Prize          uint `validate:"required,number"`
		Message        string
		Question_Count uint `validate:"required,number"`
		Duration       uint `validate:"required,number"`
		Start_Time     time.Time
	}
)

func (u *WeeklyQuiz) ID() interface{} {
	return u.Id
}

func (u *WeeklyQuiz) Tag() string {
	return "weekly_quizzes"
}

func (u *WeeklyQuiz) Database() *gorm.DB {
	return database.UseDB("app")
}

func (u *WeeklyQuiz) Migrate() dbmodel.Migration {
	return dbmodel.NewMigration(u)
}

func (w *WeeklyQuiz) WeeklyQuestions() (q []WeeklyQuestion, err error) {
	if err = w.Database().Find(&q, "quiz_id = ?", w.Id).Error; err != nil {
		return
	}
	return
}

func (w *WeeklyQuiz) Questions() (q []Question, err error) {
	wq, err := w.WeeklyQuestions()
	if err != nil {
		return nil, err
	}
	ids := []uint{}
	for _, q := range wq {
		ids = append(ids, q.QuestionId)
	}
	err = database.UseDB("core").Find(&q, "id IN ?", ids).Error
	return
}

func (u *WeeklyQuestion) ID() interface{} {
	return u.QuizId
}

func (u *WeeklyQuestion) Tag() string {
	return "weekly_quiz_questions"
}

func (u *WeeklyQuestion) Database() *gorm.DB {
	return database.UseDB("app")
}

func (u *WeeklyQuestion) Migrate() dbmodel.Migration {
	return dbmodel.NewMigration(u)
}
