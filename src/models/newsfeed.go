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
		Content      string    `gorm:"type:longtext" json:"content"`
		Interactions []User    `gorm:"many2many:newsfeed_interactions" json:"-"`
		CreatedAt    time.Time `json:"created_at"`
		UpdatedAt    time.Time `json:"updated_at"`
	}

	NewsfeedForm struct {
		Title   string `json:"title"`
		Content string `gorm:"type:longtext" json:"content"`
	}

	NewsfeedUpdateForm struct {
		Slug    string `json:"slug"`
		Title   string `json:"title"`
		Content string `gorm:"type:longtext" json:"content"`
	}

	NewsfeedInteraction struct {
		NewsfeedId   uuid.UUID
		UserId       uuid.UUID
		Liked        bool
		IsBookmarked bool
	}

	NewsfeedComments struct {
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

func (*Newsfeed) Relations() []interface{ Join() string } {
	return []interface{ Join() string }{
		NewsfeedInteraction{},
	}
}

func (NewsfeedInteraction) Join() string {
	return "Interactions"
}

//++++++++++++++++++++++++++++++++++++++++++++++++++++//
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
