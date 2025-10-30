package helpers

import (
	"context"
	"go-auth/database"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type SignedDetails struct {
	Email     string
	Username  string
	TokenType string
	Uid       string
	User_type string
	jwt.RegisteredClaims
}

var userCollection *mongo.Collection = database.OpenCollection(database.DBConnect(), "user")
var SECRET_KEY string = os.Getenv("SECRET_KEY")

func GenerateAllTokens(email string, username string, userType string, uid string) (signedToken string, signedRefreshToken string, err error) {
	claims := &SignedDetails{
		Email:     email,
		Username:  username,
		TokenType: "access",
		User_type: userType,
		Uid:       uid,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * time.Minute)),
		},
	}

	refreshClaims := &SignedDetails{
		Uid: uid,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(168 * time.Hour)),
		},
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(SECRET_KEY))
	if err != nil {
		return "", "", err
	}

	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(SECRET_KEY))
	if err != nil {
		return "", "", err
	}

	return token, refreshToken, nil
}

func UpdateAllToken(c context.Context, signedToken string, signedRefreshToken string, userId string) {
	var updateObj primitive.D
	updated_at, _ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	updateObj = append(updateObj, bson.E{Key: "updated_at", Value: updated_at})

	filter := bson.M{"user_id": userId}
	opt := options.Update().SetUpsert(false)

	_, err := userCollection.UpdateOne(
		c, filter, bson.D{
			{Key: "$set", Value: updateObj},
		},
		opt,
	)

	if err != nil {
		log.Panic(err)
		return
	}
}

func ValidateToken(signedToken string) (claims *SignedDetails, msg string) {
	token, err := jwt.ParseWithClaims(
		signedToken,
		&SignedDetails{},
		func(token *jwt.Token) (any, error) {
			return []byte(SECRET_KEY), nil
		},
	)

	if err != nil {
		msg = "error"
		return
	}

	claims, ok := token.Claims.(*SignedDetails)

	if !ok {
		msg = "token is invalid"
		return
	}

	if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now().Local()) {
		msg = "token is expired"
		return
	}
	return claims, msg
}

func GenerateResetToken(email string) (resetToken string, err error) {
	resetclaims := &SignedDetails{
		Email:     email,
		TokenType: "reset",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * time.Minute)),
		},
	}

	resetToken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, resetclaims).SignedString([]byte(SECRET_KEY))
	if err != nil {
		return "", err
	}
	return resetToken, nil
}
