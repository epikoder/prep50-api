package models

import (
	"time"

	"github.com/Prep50mobileApp/prep50-api/src/pkg/dbmodel"
	"github.com/Prep50mobileApp/prep50-api/src/services/database"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	User struct {
		Id         uuid.UUID      `sql:"primary_key;type:uuid;default:uuid_generate_v4()" gorm:"type:varchar(36);index;" json:"-"`
		UserName   string         `gorm:"type:varchar(50);column:username;notnull" json:"username"`
		Email      string         `gorm:"type:varchar(255);column:email;unique;index;notnull;" json:"email"`
		Phone      string         `gorm:"type:varchar(20);column:phone;unique;index;notnull" json:"phone"`
		Photo      string         `gorm:"type:varchar(255);column:photo" json:"photo"`
		Password   string         `gorm:"type:varchar(64);column:password" json:"-"`
		IsProvider bool           `gorm:"type:tinyint(1);" json:"-"`
		Locked     bool           `gorm:"type:tinyint(1);" json:"-"`
		Providers  []Provider     `gorm:"many2many:user_providers;" json:"-"`
		Exams      []Exam         `gorm:"many2many:user_exams;" json:"exams"`
		Device     Device         `json:"-"`
		CreatedAt  time.Time      `json:"-"`
		UpdatedAt  time.Time      `json:"-"`
		DeletedAt  gorm.DeletedAt `json:"-"`
	}

	UserProvider struct {
		UserId     uuid.UUID `sql:"primary_key;type:uuid;default:uuid_generate_v4()" gorm:"type:varchar(36);index;" json:"-"`
		ProviderId uuid.UUID `sql:"type:uuid;default:uuid_generate_v4()" gorm:"type:varchar(36);index;" json:"-"`
		Token      string
	}

	UserRegisterFormStruct struct {
		UserName string `validate:"required,alphanum"`
		Email    string `validate:"required,email"`
		Phone    string `validate:"required,numeric,min=8"`
		Password string `validate:"required,alphanum,min=6"`
	}

	UserLoginFormStruct struct {
		UserName string `validate:"required,alphanum"`
		Password string `validate:"required,alphanum"`
	}
)

func (u *User) ID() interface{} {
	return u.Id
}

func (u *User) Tag() string {
	return "users"
}

func (u *User) Database() *gorm.DB {
	return database.UseDB("app")
}

func (u *User) Migrate() dbmodel.Migration {
	return dbmodel.NewMigration(u)
}

func (u *User) Relations() []interface{ Join() string } {
	return []interface{ Join() string }{
		UserProvider{},
	}
}

func (u UserProvider) Join() string {
	return "Providers"
}
