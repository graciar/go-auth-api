package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID         primitive.ObjectID `bson:"_id"`
	Username   *string            `json:"username" bson:"username" validate:"required,max=24"`
	Email      *string            `json:"email" bson:"email" validate:"email,required"`
	Password   *string            `json:"password" validate:"required,min=6"`
	User_type  *string            `json:"user_type" validate:"required,eq=ADMIN|eq=USER"`
	Created_at time.Time          `json:"created_at" bson:"created_at"`
	Updated_at time.Time          `json:"update_at" bson:"updated_at"`
	User_id    string             `json:"user_id"`
}
