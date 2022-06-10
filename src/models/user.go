package models

import (
	"github.com/Prep50mobileApp/prep50-api/src/pkg/dbmodel"
	"github.com/Prep50mobileApp/prep50-api/src/services/database"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	User struct {
		Id        uuid.UUID  `sql:"primary_key;index;type:uuid;default:uuid_generate_v4()" json:"-"`
		UserName  string     `gorm:"type:varchar(50);column:username" json:"username"`
		Email     string     `gorm:"type:varchar(50);column:email;unique;index;" json:"email"`
		Phone     string     `gorm:"type:varchar(50);column:phone;unique;index;" json:"phone"`
		Password  string     `gorm:"type:varchar(50);column:password" json:"password"`
		Providers []Provider `gorm:"many2many:user_providers;"`
	}

	UserProvider struct {
		Token string
	}

	UserRegisterFormStruct struct {
		UserName string `validate:"required,string"`
		Email    string `validate:"required,email"`
		Phone    string `validate:"required,numeric,min=8"`
		Password string `validate:"required,string"`
	}

	UserLoginFormStruct struct {
		UName    string
		Password string
	}
)

func (u *User) Database() *gorm.DB {
	return database.UseDB("app")
}

func (u *User) Migrate() dbmodel.Migration {
	return dbmodel.NewMigration(u)
}
