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
		Id        uuid.UUID
		Name      string `gorm:"unique;index;notnull"`
		CreatedAt time.Time
		UpdatedAt time.Time
	}
)

func (e *Exam) ID() uuid.UUID {
	return e.Id
}

func (e *Exam) Tag() string {
	return "exam_types"
}

func (e *Exam) Database() *gorm.DB {
	return database.UseDB("app")
}

func (e *Exam) Migrate() dbmodel.Migration {
	return dbmodel.NewMigration(e)
}
