package models

import (
	"github.com/Prep50mobileApp/prep50-api/src/pkg/dbmodel"
	"github.com/Prep50mobileApp/prep50-api/src/services/database"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	Subject struct {
		Id uuid.UUID `sql:"primary_key;type:uuid;default:uuid_generate_v4()" gorm:"type:varchar(36);index;" json:"-"`
	}
)

func (s *Subject) ID() interface{} {
	return s.Id
}

func (s *Subject) Tag() string {
	return "subjects"
}

func (u *Subject) Database() *gorm.DB {
	return database.UseDB("app")
}

func (u *Subject) Migrate() dbmodel.Migration {
	return dbmodel.NewMigration(u)
}

// func (u *User) Relations() []interface{ Join() string } {
// 	return []interface{ Join() string }{
// 		UserProvider{},
// 	}
// }

// func (u UserProvider) Join() string {
// 	return "Providers"
// }
