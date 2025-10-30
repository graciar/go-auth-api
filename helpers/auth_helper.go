package helpers

import (
	"errors"
	"log"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Panic(err)
	}
	return string(bytes)
}

func VerifyPassword(userPass string, providedPass string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(providedPass), []byte(userPass))
	check := true
	if err != nil {
		check = false
		return check, errors.New("password is incorrect")
	}
	return check, nil
}

func CheckUserType(userType string, role string) (err error) {
	err = nil
	if userType != role {
		err = errors.New("unauthorized to access this resource")
		return err
	}
	return err
}

func MatchUserTypeToUid(c *gin.Context, userId string) (err error) {
	uid := c.GetString("uid")

	err = nil
	if uid != userId {
		err = errors.New("unauthorized to access this resource")
		return err
	}
	return err
}

func IsEmailMatch(c *gin.Context, email string) error {
	jwtEmail := c.GetString("email")

	if email != jwtEmail {
		return errors.New("unauthorized to access this resource")
	}
	return nil
}
