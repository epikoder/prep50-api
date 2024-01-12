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
		Id        uuid.UUID  `sql:"primary_key;unique;type:uuid;default:uuid_generate_v4()" gorm:"type:varchar(36);index;" json:"id"`
		Prize     uint       `json:"prize"`
		Message   string     `json:"message"`
		Duration  uint       `json:"duration"`
		StartTime time.Time  `json:"start_time"`
		Week      uint       `json:"week"`
		Session   uint       `json:"session"`
		Results   []User     `gorm:"many2many:weekly_quizz_results;foreignKey:Id;" json:"results"`
		CreatedAt time.Time  `json:"-"`
		UpdatedAt time.Time  `json:"-"`
		CreatedBy string     `json:"-"`
		UpdatedBy string     `json:"-"`
		Questions []Question `gorm:"-:migration; many2many:weekly_questions; foreignKey:Id; joinForeignKey:QuizId;" json:"-"`
	}

	WeeklyQuestion struct {
		QuizId     uuid.UUID `gorm:"type:varchar(36);index;"`
		QuestionId uint      `gorm:"index;"`
		CreatedAt  time.Time
	}

	WeeklyQuizResult struct {
		WeeklyQuizId uuid.UUID `gorm:"notnull;type:varchar(36);index;" json:"quiz_id"`
		UserId       uuid.UUID `gorm:"notnull;type:varchar(36);index;" json:"user_id"`
		Score        uint      `json:"score"`
		Duration     uint      `json:"duration"`
	}

	WeeklyQuizFormStruct struct {
		Prize      uint `validate:"required,number"`
		Message    string
		Duration   uint `validate:"required,number"`
		Start_Time time.Time
	}

	WeeklyQuizUpdateForm struct {
		Id         uuid.UUID `validate:"required"`
		Prize      uint
		Message    string
		Duration   uint
		Start_Time time.Time
	}
)

func (u *WeeklyQuiz) ID() interface{} {
	return u.Id
}

func (u *WeeklyQuiz) Tag() string {
	return "weekly_quizzes"
}

func (u *WeeklyQuiz) Database() *gorm.DB {
	return database.DB()
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

func (w *WeeklyQuiz) QuestionsWithAnswer() (q []Question, err error) {
	wq, err := w.WeeklyQuestions()
	if err != nil {
		return nil, err
	}
	ids := []uint{}
	for _, q := range wq {
		ids = append(ids, q.QuestionId)
	}
	err = database.DB().Table((&Question{}).Tag()).Find(&q, "id IN ?", ids).Error
	return
}

func (u *WeeklyQuestion) ID() interface{} {
	return u.QuizId
}

func (u *WeeklyQuestion) Tag() string {
	return "weekly_quiz_questions"
}

func (u *WeeklyQuestion) Database() *gorm.DB {
	return database.DB()
}

func (u *WeeklyQuestion) Migrate() dbmodel.Migration {
	return dbmodel.NewMigration(u)
}

func (*WeeklyQuiz) Relations() []interface{ Join() string } {
	return []interface{ Join() string }{
		WeeklyQuizResult{},
	}
}

func (WeeklyQuizResult) Join() string {
	return "Results"
}
