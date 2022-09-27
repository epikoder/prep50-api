package models

import (
	"time"

	"github.com/Prep50mobileApp/prep50-api/src/pkg/dbmodel"
	"github.com/Prep50mobileApp/prep50-api/src/services/database"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	PasswordReset struct {
		Id        uuid.UUID `sql:"primary_key;unique;type:varchar(36);index;type:uuid;default:uuid_generate_v4()" json:"-"`
		Code      int       `gorm:"type:int(4);column:code;notnull;"`
		User      string    `gorm:"type:varchar(255);notnull;"`
		CreatedAt time.Time
	}
)

func (p *PasswordReset) ID() interface{} {
	return p.Id
}

func (u *PasswordReset) Tag() string {
	return "password_resets"
}

func (p *PasswordReset) Database() *gorm.DB {
	return database.UseDB("app")
}

func (p *PasswordReset) Migrate() dbmodel.Migration {
	return dbmodel.NewMigration(p)
}
