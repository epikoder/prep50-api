package models

import (
	"time"

	"github.com/Prep50mobileApp/prep50-api/src/pkg/dbmodel"
	"github.com/Prep50mobileApp/prep50-api/src/services/database"
	"gorm.io/gorm"
)

type (
	Question struct {
		Id                 uint `sql:"primary_key;" json:"id"`
		SubjectId          uint
		SourceId           uint
		QuestionTypeId     uint
		Question           string
		QuestionDetails    string
		QuestionImage      string
		Option1            string
		Option2            string
		Option3            string
		Option4            string
		ShortAnswer        string
		FullAnswer         string
		AnswerImage        string
		AnswerDetails      string
		QuestionYear       uint
		QuestionYearNumber uint
		CreatedAt          time.Time `json:"-"`
		UpdatedAt          time.Time `json:"-"`
	}
)

func (u *Question) ID() interface{} {
	return u.Id
}

func (u *Question) Tag() string {
	return "questions"
}

func (u *Question) Database() *gorm.DB {
	return database.UseDB("core")
}

func (u *Question) Migrate() dbmodel.Migration {
	return dbmodel.NewMigration(u)
}
