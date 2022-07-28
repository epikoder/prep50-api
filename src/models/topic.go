package models

import (
	"github.com/Prep50mobileApp/prep50-api/src/pkg/dbmodel"
	"github.com/Prep50mobileApp/prep50-api/src/services/database"
	"gorm.io/gorm"
)

type (
	Topic struct {
		Id         uint             `sql:"primary_key;" json:"id"`
		SubjectId  int              `json:"subject_id"`
		Title      string           `json:"title"`
		Details    string           `json:"details"`
		Objectives []TopicObjective `json:"objectives"`
	}

	TopicObjective struct {
		Id          uint     `sql:"primary_key;" json:"id"`
		TopicId     int      `json:"topic_id"`
		ObjectiveId int      `json:"objective_id"`
		Title       string   `json:"title"`
		Details     string   `json:"details"`
		Lessons     []Lesson `gorm:"many2many:objective_lessons; foreignKey:ObjectiveId; joinForeignKey:ObjectiveId; references:Id; joinReferences:LessonId" json:"lessons"`
	}
)

func (t *Topic) ID() interface{} {
	return t.Id
}

func (t *Topic) Tag() string {
	return "topics"
}

func (*Topic) Database() *gorm.DB {
	return database.UseDB("core")
}

func (t *Topic) Migrate() dbmodel.Migration {
	return dbmodel.NewMigration(t)
}
