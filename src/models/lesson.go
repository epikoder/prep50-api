package models

import (
	"time"

	"github.com/Prep50mobileApp/prep50-api/src/pkg/dbmodel"
	"github.com/Prep50mobileApp/prep50-api/src/services/database"
	"gorm.io/gorm"
)

type (
	Lesson struct {
		Id        int       `sql:"primary_key;" json:"id"`
		SubjectId int       `json:"subject_id"`
		Title     string    `json:"title"`
		Content   string    `sql:"type:longtext" json:"content"`
		DocUrl    string    `json:"docs_url"`
		SlideUrl  string    `json:"slide_url"`
		VideoUrl  string    `json:"video_url"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"update_at"`
	}
)

func (l *Lesson) ID() interface{} {
	return l.Id
}

func (l *Lesson) Tag() string {
	return "lessons"
}

func (*Lesson) Database() *gorm.DB {
	return database.UseDB("core")
}

func (l *Lesson) Migrate() dbmodel.Migration {
	return dbmodel.NewMigration(l)
}
