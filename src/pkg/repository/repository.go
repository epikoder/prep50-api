package repository

import "github.com/google/uuid"

type (
	Repository interface {
		Get(uuid.UUID) interface{}
	}
)

func Get() {

}
