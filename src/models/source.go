package models

import (
	"time"

	"github.com/Prep50mobileApp/prep50-api/src/pkg/dbmodel"
	"github.com/Prep50mobileApp/prep50-api/src/services/database"
	"gorm.io/gorm"
)

type (
	Source struct {
		Id        uint      `sql:"primary_key;" json:"-"`
		Name      string    `json:"-"`
		Level     string    `json:"-"`
		Year      uint      `json:"-"`
		CreatedAt time.Time `json:"-"`
		UpdatedAt time.Time `json:"-"`
	}
)

func (u *Source) ID() interface{} {
	return u.Id
}

func (u *Source) Tag() string {
	return "sources"
}

func (u *Source) Database() *gorm.DB {
	return database.UseDB("core")
}

func (u *Source) Migrate() dbmodel.Migration {
	return dbmodel.NewMigration(u)
}
