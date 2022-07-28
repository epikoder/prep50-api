package models

import (
	"strings"

	"github.com/Prep50mobileApp/prep50-api/src/pkg/dbmodel"
	"github.com/Prep50mobileApp/prep50-api/src/services/database"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	Role struct {
		Id          uuid.UUID    `sql:"primary_key;unique;type:uuid;default:uuid_generate_v4()" gorm:"type:varchar(36);index;" json:"-"`
		Name        string       `json:"name"`
		Disabled    bool         `json:"disabled"`
		Permissions []Permission `gorm:"many2many:role_permissions" json:"permissions"`
	}

	Permission struct {
		Id   uuid.UUID `sql:"primary_key;type:uuid;unique;default:uuid_generate_v4()" gorm:"type:varchar(36);index;" json:"-"`
		Name string    `json:"name"`
	}

	RolePermission struct {
		RoleId       uuid.UUID `sql:"type:uuid;" gorm:"type:varchar(36);index;" json:"role_id"`
		PermissionId uuid.UUID `sql:"type:uuid;" gorm:"type:varchar(36);index;" json:"permission_id"`
		CreatedBy    string    `json:"-"`
	}
)

func (u *Role) ID() interface{} {
	return u.Id
}

func (u *Role) Tag() string {
	return "roles"
}

func (u *Role) Database() *gorm.DB {
	return database.UseDB("app")
}

func (u *Role) Migrate() dbmodel.Migration {
	return dbmodel.NewMigration(u)
}

func (u *Role) Relations() []interface{ Join() string } {
	return []interface{ Join() string }{
		RolePermission{},
	}
}

func (r *Role) MatchRole(role string) (ok bool) {
	arr := strings.Split(r.Name, "-")
	for _, a := range arr {
		if ok = a == role; ok {
			return
		}
	}
	return strings.EqualFold(r.Name, strings.ToLower(role))
}

func (u *Permission) ID() interface{} {
	return u.Id
}

func (u *Permission) Tag() string {
	return "permissions"
}

func (u *Permission) Database() *gorm.DB {
	return database.UseDB("app")
}

func (u *Permission) Migrate() dbmodel.Migration {
	return dbmodel.NewMigration(u)
}

func (p *Permission) MatchPerm(perm string) (ok bool) {
	if ok = p.Name == "*.*"; ok {
		return
	}
	permArr := strings.Split(perm, ".")
	arr := strings.Split(p.Name, ".")
	if len(permArr) != 2 || len(arr) != 2 {
		return
	}
	if arr[0] == permArr[0] {
		return arr[1] == permArr[1] || arr[1] == "*"
	}
	return strings.EqualFold(p.Name, strings.ToLower(perm))
}

func (RolePermission) Join() string {
	return "Permissions"
}
