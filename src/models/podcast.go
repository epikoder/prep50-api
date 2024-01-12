package models

import (
	"time"

	"github.com/Prep50mobileApp/prep50-api/src/pkg/dbmodel"
	"github.com/Prep50mobileApp/prep50-api/src/services/database"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	Podcast struct {
		Id        uuid.UUID `sql:"primary_key;unique;type:uuid;default:uuid_generate_v4()" gorm:"type:varchar(36);index;" json:"id"`
		SubjectId uint      `json:"subject_id"`
		TopicId   uint      `json:"topic_id"`
		Title     string    `json:"title"`
		Url       string    `json:"url"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"update_at"`
	}

	PodcastForm struct {
		Title string
		Topic uint
	}

	PodcastUpdateForm struct {
		Id        uuid.UUID
		Title     string
		Subject   uint
		Objective uint
	}
)

func (p *Podcast) ID() interface{} {
	return p.Id
}

func (u *Podcast) Tag() string {
	return "podcasts"
}

func (p *Podcast) Database() *gorm.DB {
	return database.DB()
}

func (p *Podcast) Migrate() dbmodel.Migration {
	return dbmodel.NewMigration(p)
}
