package models

import (
	"fmt"
	"reflect"
	"time"

	"github.com/Prep50mobileApp/prep50-api/src/pkg/dbmodel"
	"github.com/Prep50mobileApp/prep50-api/src/services/database"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	Newsfeed struct {
		Id           uuid.UUID `sql:"primary_key;unique;type:uuid;default:uuid_generate_v4()" gorm:"type:varchar(36);index;" json:"-"`
		UserId       uuid.UUID `gorm:"index" json:"-"`
		Slug         string    `gorm:"index;unique" json:"slug"`
		Title        string    `json:"title"`
		Photo        string    `json:"photo"`
		Content      string    `gorm:"type:longtext" json:"content"`
		Interactions []User    `gorm:"many2many:newsfeed_interactions" json:"-"`
		Comments     []User    `gorm:"many2many:newsfeed_comments;foreignKey:Id" json:"comments"`
		Reports      []User    `gorm:"many2many:newsfeed_reports" json:"reports"`
		CreatedAt    time.Time `json:"created_at"`
		UpdatedAt    time.Time `json:"updated_at"`
	}

	NewsfeedForm struct {
		Title   string `validate:"required"`
		Content string `validate:"required"`
		Photo   string
	}

	NewsfeedUpdateForm struct {
		Slug    string `json:"slug"`
		Title   string `json:"title"`
		Content string `gorm:"type:longtext" json:"content"`
	}

	NewsfeedInteraction struct {
		NewsfeedId   uuid.UUID `json:"newsfeed_id"`
		UserId       uuid.UUID `json:"user_id"`
		Liked        bool      `json:"is_liked"`
		IsBookmarked bool      `json:"is_bookmarked"`
	}

	NewsfeedComment struct {
		Id         uuid.UUID `sql:"primary_key;unique;type:uuid;default:uuid_generate_v4()" gorm:"type:varchar(36);index;" json:"id"`
		NewsfeedId uuid.UUID `gorm:"type:varchar(36)" json:"-"`
		UserId     uuid.UUID `gorm:"type:varchar(36)" json:"-"`
		Comment    string    `gorm:"type:mediumtext" json:"comment"`
		Reports    []User    `gorm:"many2many:newsfeed_comment_reports" json:"-"`
		CreatedAt  time.Time `json:"created_at"`
		UpdatedAt  time.Time `json:"updated_at"`
	}

	NewsfeedReport struct {
		Id         uuid.UUID `sql:"primary_key;unique;type:uuid;default:uuid_generate_v4()" gorm:"type:varchar(36);index;" json:"id"`
		NewsfeedId uuid.UUID `json:"newsfeed_id"`
		UserId     uuid.UUID `json:"user_id"`
		Type       string    `json:"type"`
		Message    string    `gorm:"type:longtext" json:"message"`
		CreatedAt  time.Time `json:"created_at"`
		UpdatedAt  time.Time `json:"updated_at"`
	}

	NewsfeedCommentReport struct {
		Id                uuid.UUID `sql:"primary_key;unique;type:uuid;default:uuid_generate_v4()" gorm:"type:varchar(36);index;" json:"id"`
		NewsfeedCommentId uuid.UUID `gorm:"type:varchar(36)" json:"newsfeed_comment_id"`
		UserId            uuid.UUID `gorm:"type:varchar(36)" json:"user_id"`
		Type              string    `json:"type"`
		Message           string    `gorm:"type:longtext" json:"message"`
		CreatedAt         time.Time `json:"created_at"`
		UpdatedAt         time.Time `json:"updated_at"`
	}
)

func (m *Newsfeed) ID() interface{} {
	return m.Id
}

func (m *Newsfeed) Tag() string {
	return "newsfeed"
}

func (m *Newsfeed) Database() *gorm.DB {
	return database.UseDB("app")
}

func (m *Newsfeed) Migrate() dbmodel.Migration {
	return dbmodel.NewMigration(m)
}

func (m *NewsfeedComment) ID() interface{} {
	return m.Id
}

func (m *NewsfeedComment) Tag() string {
	return "newsfeed_comments"
}

func (m *NewsfeedComment) Database() *gorm.DB {
	return database.UseDB("app")
}

func (m *NewsfeedComment) Migrate() dbmodel.Migration {
	return dbmodel.NewMigration(m)
}

func (*Newsfeed) Relations() []interface{ Join() string } {
	return []interface{ Join() string }{
		NewsfeedInteraction{},
		NewsfeedComment{},
		NewsfeedReport{},
	}
}

func (*NewsfeedComment) Relations() []interface{ Join() string } {
	return []interface{ Join() string }{
		NewsfeedCommentReport{},
	}
}

func (NewsfeedInteraction) Join() string {
	return "Interactions"
}

func (NewsfeedComment) Join() string {
	return "Comments"
}

func (NewsfeedReport) Join() string {
	return "Reports"
}

func (NewsfeedCommentReport) Join() string {
	return "Reports"
}

// ++++++++++++++++++++++++++++++++++++++++++++++++++++//
func UniqueSlug(model dbmodel.DBModel, slug string) (s string, err error) {
	s = slug
	tries := 1
	for {
		m := reflect.New(reflect.TypeOf(model).Elem()).Interface().(dbmodel.DBModel)
		if err = model.Database().Find(m, "slug = ?", s).Error; err != nil {
			return
		}
		uid, ok := m.ID().(uuid.UUID)
		fmt.Println(uid, ok)
		if !ok {
			if id := m.ID().(int); id == 0 {
				fmt.Println(id)
				return
			}
		} else if uid == uuid.Nil {
			return
		}
		s = fmt.Sprintf("%s-%d", slug, tries)
		tries++
	}
}
