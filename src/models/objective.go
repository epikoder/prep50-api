package models

import (
	"github.com/Prep50mobileApp/prep50-api/src/pkg/dbmodel"
	"github.com/Prep50mobileApp/prep50-api/src/services/database"
	"gorm.io/gorm"
)

type (
	Objective struct {
		Id        uint       `sql:"primary_key;" json:"id"`
		SubjectId int        `json:"subject_id"`
		Title     string     `json:"title"`
		Details   string     `json:"details"`
		Questions []Question `gorm:"many2many:objective_questions; foreignKey:Id; joinForeignKey:ObjectiveId; joinReferences:Id" json:"questions,omitempty"`
		Lessons   []Lesson   `gorm:"many2many:objective_lessons; foreignKey:Id; joinForeignKey:ObjectiveId; joinReferences:Id" json:"lessons"`
		Progress  uint       `gorm:"-:migration" json:"progress"`
	}

	// UserObjectiveProgress struct {
	// 	Objective
	// 	Progress uint `json:"progress"`
	// }
)

func (o *Objective) Database() *gorm.DB {
	return database.DB()
}

func (u *Objective) ID() interface{} {
	return u.Id
}

func (u *Objective) Tag() string {
	return "objectives"
}

func (u *Objective) Migrate() dbmodel.Migration {
	return dbmodel.NewMigration(u)
}

func (o Objective) GetPodcasts() (p []Podcast) {
	p = make([]Podcast, 0)
	database.DB().Find(&p, "objective_id = ?", o.Id)
	return
}
