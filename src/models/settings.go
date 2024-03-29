package models

import (
	"unicode"

	"github.com/Prep50mobileApp/prep50-api/src/pkg/dbmodel"
	"github.com/Prep50mobileApp/prep50-api/src/services/database"
	"gorm.io/gorm"
)

type GeneralSetting struct {
	Id       uint   `gorm:"primarykey"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Location string `json:"location"`
	Website  string `json:"website"`
	Terms    string `gorm:"type:longText" json:"terms"`
	Privacy  string `gorm:"type:longText" json:"privacy"`
}

func (p *GeneralSetting) ID() interface{} {
	return p.Id
}

func (u *GeneralSetting) Tag() string {
	return "general_settings"
}

func (p *GeneralSetting) Database() *gorm.DB {
	return database.DB()
}

func (p *GeneralSetting) Migrate() dbmodel.Migration {
	return dbmodel.NewMigration(p)
}

func (*GeneralSetting) Field(s string) string {
	r := ""
	useNext := false
	for i, v := range s {
		if unicode.IsLower(v) && i == 0 {
			r += string(unicode.ToUpper(v))
			continue
		}

		if v == '_' {
			useNext = true
			continue
		}

		if useNext {
			r += string(unicode.ToUpper(v))
			useNext = false
			continue
		}
		r += string(v)
	}
	return r
}
