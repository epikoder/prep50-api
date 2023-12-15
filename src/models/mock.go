package models

import (
	"time"

	"github.com/Prep50mobileApp/prep50-api/src/pkg/dbmodel"
	"github.com/Prep50mobileApp/prep50-api/src/services/database"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	Mock struct {
		Id        uuid.UUID `sql:"primary_key;unique;type:uuid;default:uuid_generate_v4()" gorm:"type:varchar(36);index;" json:"id"`
		Amount    uint      `json:"amount"`
		Duration  uint      `json:"duration"`
		StartTime time.Time `json:"start_time"`
		EndTime   time.Time `json:"end_time"`
		Session   uint      `json:"session"`
		Video     string    `json:"video"`
		CreatedAt time.Time `json:"-"`
		UpdatedAt time.Time `json:"-"`
		CreatedBy string    `json:"-"`
		UpdatedBy string    `json:"-"`
	}

	MockQuestion struct {
		Id         uuid.UUID `sql:"primary_key;unique;type:uuid;default:uuid_generate_v4()" gorm:"type:varchar(36);index;" json:"id"`
		MockId     uuid.UUID `gorm:"type:varchar(36);index;"`
		QuestionId uint      `gorm:"index;"`
		CreatedBy  string    `json:"-"`
	}

	MockForm struct {
		Amount     uint      `validate:"required"`
		Start_Time time.Time `validate:"required"`
		End_Time   time.Time `validate:"required"`
		Duration   uint      `validate:"required"`
	}

	MockUpdateForm struct {
		Id         uuid.UUID
		Amount     string
		Start_Time time.Time
		End_Time   time.Time
		Duration   uint
	}
)

func (p *Mock) ID() interface{} {
	return p.Id
}

func (u *Mock) Tag() string {
	return "providers"
}

func (p *Mock) Database() *gorm.DB {
	return database.DB()
}

func (p *Mock) Migrate() dbmodel.Migration {
	return dbmodel.NewMigration(p)
}

func (m *MockQuestion) ID() interface{} {
	return m.MockId
}

func (m *MockQuestion) Tag() string {
	return "weekly_quiz_questions"
}

func (m *MockQuestion) Database() *gorm.DB {
	return database.DB()
}

func (m *MockQuestion) Migrate() dbmodel.Migration {
	return dbmodel.NewMigration(m)
}

func (m *Mock) MockQuestions() (q []MockQuestion, err error) {
	if err = m.Database().Find(&q, "mock_id = ?", m.Id).Error; err != nil {
		return
	}
	return
}

func (m *Mock) Questions() (q []Question, err error) {
	wq, err := m.MockQuestions()
	if err != nil {
		return nil, err
	}
	ids := []uint{}
	for _, q := range wq {
		ids = append(ids, q.QuestionId)
	}
	err = database.DB().Find(&q, "id IN ?", ids).Error
	return
}
