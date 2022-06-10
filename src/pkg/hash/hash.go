package hash

import (
	"github.com/Prep50mobileApp/prep50-api/src/pkg/logger"
	"golang.org/x/crypto/bcrypt"
)

func MakeHash(i string) (r string, err error) {
	b, err := bcrypt.GenerateFromPassword([]byte(i), 12)
	if !logger.HandleError(err) {
		return "", err
	}
	return string(b), nil
}

func CheckHash(h string, s string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(h), []byte(s))
	return err == nil
}
