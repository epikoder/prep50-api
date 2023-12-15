package models

import (
	"time"

	"github.com/Prep50mobileApp/prep50-api/src/pkg/dbmodel"
	"github.com/Prep50mobileApp/prep50-api/src/services/database"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	Transaction struct {
		Id        uuid.UUID `sql:"primary_key;unique;type:uuid;default:uuid_generate_v4()" gorm:"type:varchar(36);index;" json:"id"`
		UserId    uuid.UUID `gorm:"type:varchar(36);index" json:"user_id"`
		Item      string    `json:"-"`
		Amount    uint      `json:"amount"`
		Reference string    `json:"reference"`
		Provider  string    `json:"provider"`
		Status    string    `json:"status"`
		Session   uint      `gorm:"notnull" json:"session"`
		Response  string    `gorm:"type:longtext" json:"response"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	}
)

func (t *Transaction) ID() interface{} {
	return t.Id
}

func (t *Transaction) Tag() string {
	return "transactions"
}

func (t *Transaction) Database() *gorm.DB {
	return database.DB()
}

func (t *Transaction) Migrate() dbmodel.Migration {
	return dbmodel.NewMigration(t)
}

func (t *Transaction) OverrideMigration() bool {
	return true
}
