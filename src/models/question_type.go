package models

import (
	"time"

	"github.com/Prep50mobileApp/prep50-api/src/pkg/dbmodel"
	"github.com/Prep50mobileApp/prep50-api/src/services/database"
	"gorm.io/gorm"
)

type (
	QuestionType struct {
		Id        uint      `sql:"primary_key;" json:"id"`
		Title     string    `json:"-"`
		Details   string    `json:"-"`
		CreatedAt time.Time `json:"-"`
		UpdatedAt time.Time `json:"-"`
	}
)

func (u *QuestionType) ID() interface{} {
	return u.Id
}

func (u *QuestionType) Tag() string {
	return "question_types"
}

func (u *QuestionType) Database() *gorm.DB {
	return database.UseDB("core")
}

func (u *QuestionType) Migrate() dbmodel.Migration {
	return dbmodel.NewMigration(u)
}
