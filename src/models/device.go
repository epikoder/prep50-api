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
		Id         uuid.UUID `sql:"primary_key" gorm:"unique;index;notnull"`
		UserID     uuid.UUID `gorm:"unique;index;notnull;type:varchar(36)"`
		Identifier string    `gorm:"unique;index;notnull"`
		Name       string
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
	return database.DB()
}

func (d *Device) Migrate() dbmodel.Migration {
	return dbmodel.NewMigration(d)
}
