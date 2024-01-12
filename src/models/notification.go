package models

import (
	"time"

	"github.com/Prep50mobileApp/prep50-api/src/pkg/dbmodel"
	"github.com/Prep50mobileApp/prep50-api/src/services/database"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	Notification struct {
		Id        uuid.UUID `sql:"primary_key;unique;type:varchar(36);index;type:uuid;default:uuid_generate_v4()" json:"id"`
		UserId    uuid.UUID `sql:"type:uuid;default:uuid_generate_v4()" gorm:"type:varchar(36);index" json:"-"`
		Title     string    `json:"title" validate:"required"`
		Body      string    `gorm:"type:mediumtext" json:"body" validate:"required"`
		ImageUrl  string    `json:"image_url"`
		CreatedAt time.Time `gorm:"created_at" json:"created_at"`
	}
)

func (n *Notification) ID() interface{} {
	return n.Id
}

func (*Notification) Tag() string {
	return "notifications"
}

func (*Notification) Database() *gorm.DB {
	return database.DB()
}

func (n *Notification) Migrate() dbmodel.Migration {
	return dbmodel.NewMigration(n)
}

func (*Notification) OverrideMigration() bool {
	return true
}
