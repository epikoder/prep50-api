package models

import (
	"github.com/Prep50mobileApp/prep50-api/src/pkg/dbmodel"
	"github.com/Prep50mobileApp/prep50-api/src/services/database"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	Provider struct {
		Id   uuid.UUID `sql:"primary_key;index;type:uuid;default:uuid_generate_v4()" json:"-"`
		Name string    `gorm:"type:varchar(50);column:name" json:"name"`
	}
)

func (p *Provider) ID() uuid.UUID {
	return p.Id
}

func (u *Provider) Tag() string {
	return "providers"
}

func (p *Provider) Database() *gorm.DB {
	return database.UseDB("app")
}

func (p *Provider) Migrate() dbmodel.Migration {
	return dbmodel.NewMigration(p)
}
