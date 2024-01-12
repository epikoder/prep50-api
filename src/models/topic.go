package models

import (
	"github.com/Prep50mobileApp/prep50-api/src/pkg/dbmodel"
	"github.com/Prep50mobileApp/prep50-api/src/services/database"
	"gorm.io/gorm"
)

type (
	Topic struct {
		Id         uint        `sql:"primary_key;" json:"id"`
		SubjectId  int         `json:"subject_id"`
		Title      string      `json:"title"`
		Details    string      `json:"details"`
		Objectives []Objective `gorm:"many2many:topic_objectives; foreignKey:Id; joinForeignKey:TopicId; joinReferences:Id" json:"objectives"`
	}

	PodcastTopic struct {
		Id        uint      `sql:"primary_key;" json:"id"`
		SubjectId int       `json:"subject_id"`
		Title     string    `json:"title"`
		Details   string    `json:"details"`
		Podcast   []Podcast `gorm:"foreignKey:TopicId;references:Id" json:"podcasts"`
	}
)

func (t *Topic) ID() interface{} {
	return t.Id
}

func (t *Topic) Tag() string {
	return "topics"
}

func (*Topic) Database() *gorm.DB {
	return database.DB()
}

func (t *Topic) Migrate() dbmodel.Migration {
	return dbmodel.NewMigration(t)
}
