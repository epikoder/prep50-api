package repository

import (
	"os"

	"github.com/Prep50mobileApp/prep50-api/src/pkg/dbmodel"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	repository struct {
		Model dbmodel.DBModel
		DB    *gorm.DB
	}

	Type = interface{}
)

func NewRepository(model dbmodel.DBModel) (r *repository) {
	r = &repository{model, model.Database()}
	if env := os.Getenv("APP_ENV"); env != "" && env != "production" {
		r.DB = r.DB.Debug()
	}
	return
}

func (r *repository) Preload(query string, args ...interface{}) *repository {
	r.DB = r.DB.Preload(query, args...)
	return r
}

func (r *repository) Get(i interface{}, condition ...string) (err error) {
	return
}

func (r *repository) First() error {
	return r.DB.First(r.Model).Error
}
func (r *repository) FindOne(i ...interface{}) (ok bool) {
	tx := r.DB.First(r.Model, i...)
	if tx.Error != nil {
		return
	}
	uid, ok := r.Model.ID().(uuid.UUID)
	if !ok {
		return tx.Error == nil
	}
	return uid != uuid.Nil
}

func (r *repository) FindOneDst(dst interface{}, i ...interface{}) error {
	return r.DB.First(dst, i...).Error
}

func (r *repository) FindMany(dst interface{}, i ...interface{}) error {
	return r.DB.Find(dst, i...).Error
}

func (r *repository) Create() error {
	return r.DB.Create(r.Model).Error
}

func (r *repository) Save(data ...Type) error {
	if len(data) > 0 {
		for _, v := range data {
			if err := r.DB.Model(r.Model).Save(v).Error; err != nil {
				return err
			}
		}
		return nil
	}
	return r.DB.Save(r.Model).Error
}

func (r *repository) Delete(v interface{}, args ...interface{}) error {
	return r.DB.Delete(v, args...).Error
}
