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
		Id            uuid.UUID      `sql:"primary_key;unique;type:uuid;default:uuid_generate_v4()" gorm:"type:varchar(36);index;" json:"-"`
		UserName      string         `gorm:"type:varchar(50);unique;column:username;notnull" json:"username"`
		Email         string         `gorm:"type:varchar(255);column:email;unique;index;notnull;" json:"email"`
		Phone         string         `gorm:"type:varchar(20);column:phone;" json:"phone"`
		Photo         string         `gorm:"type:varchar(255);column:photo" json:"photo"`
		Password      string         `gorm:"type:varchar(64);column:password" json:"-"`
		Address       string         `json:"address"`
		Gender        Gender         `gorm:"type:varchar(1)" json:"gender"`
		Referral      string         `json:"referral"`
		ReferralBonus uint           `json:"-"`
		IsProvider    bool           `gorm:"type:tinyint(1);" json:"-"`
		Locked        bool           `gorm:"type:tinyint(1);" json:"-"`
		IsAdmin       bool           `json:"-"`
		CreatedAt     time.Time      `json:"created_at"`
		UpdatedAt     time.Time      `json:"-"`
		DeletedAt     gorm.DeletedAt `json:"-"`
		Device        Device         `json:"-"`
		Fcm           Fcm            `json:"-"`
		Mock          []Mock         `gorm:"many2many:user_mocks;" json:"mocks,omitempty"`
		Providers     []Provider     `gorm:"many2many:user_providers;" json:"-"`
		Exams         []Exam         `gorm:"many2many:user_exams;" json:"exams"`
		Roles         []Role         `gorm:"many2many:user_roles" json:"-"`
		Permissions   []Permission   `gorm:"many2many:user_permissions" json:"-"`
		Transactions  []Transaction  `json:"-"`
		Notifications []Notification `json:"-"`
	}

	UserExam struct {
		Id            uuid.UUID     `sql:"primary_key;type:uuid;default:uuid_generate_v4()" gorm:"type:varchar(36);notnull;index;" json:"-"`
		UserId        uuid.UUID     `sql:"type:uuid;" gorm:"type:varchar(36);notnull;index;" json:"-"`
		ExamId        uuid.UUID     `sql:"type:uuid;" gorm:"type:varchar(36);notnull;" json:"-"`
		TransactionId uuid.UUID     `sql:"type:uuid;" gorm:"type:varchar(36);index;" json:"-"`
		Session       uint          `json:"session"`
		PaymentStatus PaymentStatus `json:"payment_status"`
		CreatedAt     time.Time     `json:"created_at"`
		ExpiresAt     time.Time     `json:"expires_at"`
	}

	UserProvider struct {
		UserId     uuid.UUID `sql:"primary_key;type:uuid;default:uuid_generate_v4()" gorm:"type:varchar(36);index;" json:"-"`
		ProviderId uuid.UUID `sql:"type:uuid;" gorm:"type:varchar(36);index;" json:"-"`
		IsLoggedIn bool      `gorm:"type:tinyint(1);" json:"-"`
		Token      string
	}

	UserRole struct {
		UserId    uuid.UUID `sql:"primary_key;type:uuid;default:uuid_generate_v4()" gorm:"type:varchar(36);index;" json:"-"`
		RoleId    uuid.UUID `sql:"type:uuid;" gorm:"type:varchar(36);index;" json:"-"`
		CreatedBy string    `json:"-"`
	}

	UserPermission struct {
		UserId       uuid.UUID `sql:"primary_key;type:uuid;default:uuid_generate_v4()" gorm:"type:varchar(36);index;" json:"-"`
		PermissionId uuid.UUID `sql:"type:uuid;" gorm:"type:varchar(36);index;" json:"-"`
		CreatedBy    string    `json:"-"`
	}

	UserRegisterFormStruct struct {
		UserName string `validate:"required,alphanum"`
		Email    string `validate:"required,email"`
		Phone    string `validate:"required,numeric,min=8"`
		Password string `validate:"required,alphanum,min=6"`
		Referral string
	}

	UserLoginFormStruct struct {
		UserName string
		User     string
		Password string `validate:"required,alphanum"`
	}

	PaymentStatus string
	Gender        string

	AdminUser struct {
		User
		Permissions []string `json:"permisions"`
		Roles       []string `json:"roles"`
	}
)

const (
	Pending   PaymentStatus = "pending"
	Completed PaymentStatus = "completed"

	Male   Gender = "M"
	Female Gender = "F"
)

func GetGender(s string) Gender {
	switch t := Gender(s); t {
	case Male, Female:
		return t
	default:
		return Gender("")
	}
}

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

func (*User) Relations() []interface{ Join() string } {
	return []interface{ Join() string }{
		UserProvider{},
		UserExam{},
		UserRole{},
	}
}

func (u *User) HasPermission(perm string) (ok bool) {
	if u.Permissions == nil || u.Roles == nil {
		if err := u.Database().
			Preload("Roles.Permissions").
			Preload("Permissions").
			Find(u).Error; err != nil {
			return
		}
	}
	permissions := []Permission{}
	permissions = append(permissions, u.Permissions...)
	for _, role := range u.Roles {
		permissions = append(permissions, role.Permissions...)
	}

	for _, p := range permissions {
		if ok = p.MatchPerm(perm); ok {
			return
		}
	}
	return
}

func (u *User) HasRole(role string) (ok bool) {
	if u.Roles == nil {
		if err := u.Database().Preload("Roles").Find(u).Error; err != nil {
			return
		}
	}
	for _, r := range u.Roles {
		if ok = r.MatchRole(role); ok {
			return
		}
	}
	return
}

func (UserProvider) Join() string {
	return "Providers"
}

func (UserExam) Join() string {
	return "Exams"
}

func (UserRole) Join() string {
	return "Roles"
}

func (UserPermission) Join() string {
	return "Permissions"
}

func (*User) OverrideMigration() bool {
	return true
}
