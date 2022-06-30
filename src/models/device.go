package models

import (
	"time"

	"github.com/Prep50mobileApp/prep50-api/src/pkg/dbmodel"
	"github.com/Prep50mobileApp/prep50-api/src/services/database"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	Device struct {
		Id         uuid.UUID
		UserID     uuid.UUID `sql:"primary_key" gorm:"unique;index;notnull"`
		Identifier string    `gorm:"unique;index;notnull"`
		Name       string
		Token      string `gorm:"type:text"`
		CreatedAt  time.Time
		UpdatedAt  time.Time
	}
)

func (d *Device) ID() interface{} {
	return d.Id
}

func (d *Device) Tag() string {
	return "devices"
}

func (d *Device) Database() *gorm.DB {
	return database.UseDB("app")
}

func (d *Device) Migrate() dbmodel.Migration {
	return dbmodel.NewMigration(d)
}
