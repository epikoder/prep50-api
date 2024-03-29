package models

import (
	"github.com/Prep50mobileApp/prep50-api/src/pkg/dbmodel"
	"github.com/Prep50mobileApp/prep50-api/src/services/database"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	Subject struct {
		Id          uint        `sql:"primary_key;" json:"id"`
		Name        string      `json:"name"`
		Description string      `json:"description"`
		Objectives  []Objective `json:"objectives"`
		Lessons     []Lesson    `json:"-"`
	}

	UserSubject struct {
		Id         uuid.UUID `sql:"primary_key;type:uuid;unique;default:uuid_generate_v4()" gorm:"type:varchar(36);index;" json:"id"`
		UserId     uuid.UUID `sql:"type:uuid;default:uuid_generate_v4()" gorm:"type:varchar(36);index;" json:"-"`
		UserExamId uuid.UUID `json:"-"`
		SubjectId  uint      `json:"subject_id"`
	}

	UserSubjectProgress struct {
		Id          uint   `sql:"primary_key;" json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Progress    uint   `json:"progress"`
	}
)

func (s *Subject) ID() interface{} {
	return s.Id
}

func (s *Subject) Tag() string {
	return "subjects"
}

func (*Subject) Database() *gorm.DB {
	return database.DB()
}

func (s *Subject) Migrate() dbmodel.Migration {
	return dbmodel.NewMigration(s)
}

func (u *UserSubject) ID() interface{} {
	return u.Id
}

func (u *UserSubject) Tag() string {
	return "user_subjects"
}

func (u *UserSubject) Database() *gorm.DB {
	return database.DB()
}

func (u *UserSubject) Migrate() dbmodel.Migration {
	return dbmodel.NewMigration(u)
}
