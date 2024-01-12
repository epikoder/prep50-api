package models

import (
	"math"

	"github.com/Prep50mobileApp/prep50-api/src/pkg/dbmodel"
	"github.com/Prep50mobileApp/prep50-api/src/services/database"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserProgress struct {
	UserId      uuid.UUID `gorm:"primaryKey;type:varchar(36)" json:"-"`
	ObjectiveId uint      `gorm:"primaryKey;autoIncrement:false" json:"-"`
	SubjectId   uint      `gorm:"primaryKey;autoIncrement:false" json:"-"`
	Score       uint      `json:"score"`
}

func (u *UserProgress) ID() interface{} {
	return u.UserId
}

func (u *UserProgress) Tag() string {
	return "user_progress"
}

func (u *UserProgress) Database() *gorm.DB {
	return database.DB()
}

func (u *UserProgress) Migrate() dbmodel.Migration {
	return dbmodel.NewMigration(u)
}

func FindObjectiveProgressFromList(ups []UserProgress, objectiveId uint) uint {
	for _, _up_ := range ups {
		if _up_.ObjectiveId == objectiveId {
			return _up_.Score
		}
	}
	return 0
}

func FindSubjectProgressFromList(ups []UserProgress, subjectId uint) (s uint) {
	s = 0
	var n float64 = 0
	for _, _up_ := range ups {
		if _up_.SubjectId == subjectId {
			s += _up_.Score
			n++
		}
	}
	if n != 0 {
		var count int64 = 1
		database.DB().Model(&Objective{}).Where("subject_id = ?", subjectId).Count(&count)
		return uint(math.Round(float64(s) / float64(count)))
	}
	return
}
