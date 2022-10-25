package models

import (
	"time"

	"github.com/Prep50mobileApp/prep50-api/src/pkg/dbmodel"
	"github.com/Prep50mobileApp/prep50-api/src/services/database"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	Exam struct {
		Id           uuid.UUID `sql:"type:uuid;" gorm:"unique" json:"-"`
		Name         string    `gorm:"unique;index;notnull" json:"name"`
		Description  string    `json:"description"`
		Amount       uint      `json:"price"`
		SubjectCount int       `json:"subject_count"`
		Status       bool      `json:"status"`
		CreatedAt    time.Time `json:"-"`
		UpdatedAt    time.Time `json:"-"`
	}

	ExamPackage struct{}
)

func (e *Exam) ID() interface{} {
	return e.Id
}

func (e *Exam) Tag() string {
	return "exams"
}

func (e *Exam) Database() *gorm.DB {
	return database.UseDB("app")
}

func (e *Exam) Migrate() dbmodel.Migration {
	return dbmodel.NewMigration(e)
}

func (e *Exam) OverrideMigration() bool {
	return true
}
