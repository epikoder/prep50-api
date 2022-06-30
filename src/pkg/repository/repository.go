package repository

import (
	"github.com/Prep50mobileApp/prep50-api/src/pkg/dbmodel"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	repository struct {
		Model dbmodel.DBModel
		DB    *gorm.DB
	}
)

func NewRepository(model dbmodel.DBModel) (r *repository) {
	r = &repository{model, model.Database()}
	return
}

func (r *repository) Preload(p ...string) *repository {
	for _, v := range p {
		r.DB = r.DB.Preload(v)
	}
	return r
}

func (r *repository) Get(i interface{}, condition ...string) (err error) {
	return
}

func (r *repository) FindOne(i ...interface{}) (ok bool) {
	tx := r.DB.First(r.Model, i...)
	uid, ok := r.Model.ID().(uuid.UUID)
	if !ok {
		return tx.Error == nil
	}
	return uid != uuid.Nil
}

func (r *repository) FindMany(i ...interface{}) error {
	return r.DB.Find(r.Model, i...).Error
}

func (r *repository) Create() error {
	return r.DB.Create(r.Model).Error
}

func (r *repository) Save() error {
	return r.DB.Save(r.Model).Error
}
