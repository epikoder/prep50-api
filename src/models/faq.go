package models

import (
	"github.com/Prep50mobileApp/prep50-api/src/pkg/dbmodel"
	"github.com/Prep50mobileApp/prep50-api/src/services/database"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Faq struct {
	Id      uuid.UUID `sql:"primary_key;unique;type:uuid;default:uuid_generate_v4()" gorm:"type:varchar(36);index;" json:"-"`
	Title   string    `json:"title"`
	Slug    string    `json:"slug"`
	Content string    `gorm:"type:longtext" json:"content"`
}

func (f *Faq) ID() interface{} {
	return f.Id
}

func (*Faq) Tag() string {
	return "faqs"
}

func (*Faq) Database() *gorm.DB {
	return database.DB()
}

func (f *Faq) Migrate() dbmodel.Migration {
	return dbmodel.NewMigration(f)
}
