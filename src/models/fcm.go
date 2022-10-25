package models

import (
	"time"

	"github.com/Prep50mobileApp/prep50-api/src/pkg/dbmodel"
	"github.com/Prep50mobileApp/prep50-api/src/services/database"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Fcm struct {
	Id        uuid.UUID `sql:"primary_key;unique;type:varchar(36);index;type:uuid;default:uuid_generate_v4()" json:"id"`
	UserId    uuid.UUID `sql:"type:varchar(36);index;type:uuid;default:uuid_generate_v4()" json:"user_id"`
	Token     string
	Timestamp time.Time
}

func (n *Fcm) ID() interface{} {
	return n.Id
}

func (*Fcm) Tag() string {
	return "notifications"
}

func (*Fcm) Database() *gorm.DB {
	return database.UseDB("app")
}

func (n *Fcm) Migrate() dbmodel.Migration {
	return dbmodel.NewMigration(n)
}

func (*Fcm) OverrideMigration() bool {
	return true
}
