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
		Id          uuid.UUID `sql:"primary_key;unique;type:uuid;default:uuid_generate_v4()" gorm:"type:varchar(36);index;" json:"-"`
		SubjectId   uint      `json:"subject_id"`
		ObjectiveId uint      `json:"objective_id"`
		Title       string    `json:"title"`
		Url         string    `json:"url"`
		CreatedAt   time.Time `json:"created_at"`
		UpdatedAt   time.Time `json:"update_at"`
	}

	PodcastForm struct {
		Title     string
		Objective uint
	}

	PodcastUpdateForm struct {
		Id        uuid.UUID
		Title     string
		Subject   uint
		Objective uint
	}

	PodcastTopic struct {
		Topic
		Objectives []PodcastObjective `json:"objectives"`
	}

	UserPodcastTopicProgress struct {
		Id         uint                           `sql:"primary_key;" json:"id"`
		SubjectId  int                            `json:"subject_id"`
		Title      string                         `json:"title"`
		Details    string                         `json:"details"`
		Objectives []UserPodcastObjectiveProgress `json:"objectives"`
		Progress   uint                           `json:"-"`
	}

	PodcastObjective struct {
		Objective
		Podcasts []Podcast `json:"podcasts"`
	}

	UserPodcastObjectiveProgress struct {
		PodcastObjective
		Progress uint `json:"progress"`
	}
)

func (p *Podcast) ID() interface{} {
	return p.Id
}

func (u *Podcast) Tag() string {
	return "podcasts"
}

func (p *Podcast) Database() *gorm.DB {
	return database.UseDB("app")
}

func (p *Podcast) Migrate() dbmodel.Migration {
	return dbmodel.NewMigration(p)
}

func (po *PodcastObjective) FilterPodcast(podcasts []Podcast) *PodcastObjective {
	for _, p := range podcasts {
		if p.ObjectiveId == po.Id {
			po.Podcasts = append(po.Podcasts, p)
		}
	}
	return po
}
