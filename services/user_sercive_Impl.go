package services

import (
	"context"
	"errors"
	"go-auth/helpers"
	"go-auth/models"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserServiceImpl struct {
	usercollection *mongo.Collection
	otpcollection  *mongo.Collection
}

func NewUserService(usercollection *mongo.Collection, otpcollection *mongo.Collection) UserService {
	return &UserServiceImpl{
		usercollection: usercollection,
		otpcollection:  otpcollection,
	}
}

type PaginatedUsers struct {
	TotalCount int            `bson:"total_count"`
	UserItems  []*models.User `bson:"user_items"`
}

func (u *UserServiceImpl) Signup(c context.Context, user *models.User) error {
	password := helpers.HashPassword(*user.Password)
	user.Password = &password

	// check for existing email
	emailCount, err := u.usercollection.CountDocuments(c, bson.M{"email": user.Email})
	if err != nil {
		return err
	}
	if emailCount > 0 {
		return errors.New("email already exists")
	}

	user.ID = primitive.NewObjectID()
	user.User_id = user.ID.Hex()
	user.Created_at = time.Now()
	user.Updated_at = time.Now()

	_, insertErr := u.usercollection.InsertOne(c, user)
	if insertErr != nil {
		return insertErr
	}
	return nil
}

func (u *UserServiceImpl) Login(c context.Context, email *string, password *string) (string, string, *models.User, error) {
	var foundUser models.User
	if err := u.usercollection.FindOne(c, bson.M{"email": email}).Decode(&foundUser); err != nil {
		return "", "", nil, errors.New("email is not found")
	}

	if foundUser.Email == nil {
		return "", "", nil, errors.New("user not found")
	}
	passwordIsValid, err := helpers.VerifyPassword(*password, *foundUser.Password)
	if !passwordIsValid {
		return "", "", nil, err
	}

	token, refreshToken, _ := helpers.GenerateAllTokens(
		*foundUser.Email,
		*foundUser.Username,
		*foundUser.User_type,
		foundUser.User_id,
	)

	foundUser.Password = nil

	return token, refreshToken, &foundUser, nil
}

func (u *UserServiceImpl) GetAll(c context.Context, page, recordPerPage, startIndex int) ([]*models.User, error) {
	matchStage := bson.D{{Key: "$match", Value: bson.D{{}}}}

	groupStage := bson.D{
		{Key: "$group", Value: bson.D{
			{Key: "_id", Value: bson.D{{Key: "_id", Value: "null"}}},
			{Key: "total_count", Value: bson.D{{Key: "$sum", Value: 1}}},
			{Key: "data", Value: bson.D{{Key: "$push", Value: "$$ROOT"}}},
		}},
	}

	projectStage := bson.D{
		{Key: "$project", Value: bson.D{
			{Key: "_id", Value: 0},
			{Key: "total_count", Value: 1},
			{Key: "user_items", Value: bson.D{{Key: "$slice", Value: []interface{}{"$data", startIndex, recordPerPage}}}},
		}},
	}

	result, err := u.usercollection.Aggregate(c, mongo.Pipeline{
		matchStage, groupStage, projectStage,
	})
	if err != nil {
		return nil, err
	}
	var allUsers []PaginatedUsers
	if err = result.All(c, &allUsers); err != nil {
		return nil, err
	}
	return allUsers[0].UserItems, nil
}

func (u *UserServiceImpl) GetUser(c context.Context, userId *string) (*models.User, error) {
	var user models.User

	err := u.usercollection.FindOne(c, bson.M{"user_id": userId}).Decode(&user)
	if err != nil {
		return nil, errors.New("email or password is incorrect")
	}

	return &user, nil
}

func (u *UserServiceImpl) UpdateUser(c context.Context, user *models.User) error {
	filter := bson.M{"email": user.Email}
	if *user.Email == "" {
		return errors.New("user_id is not found")
	}
	update := bson.M{
		"$set": bson.M{
			"email":      user.Email,
			"username":   user.Username,
			"updated_at": time.Now(),
		},
	}

	_, err := u.usercollection.UpdateOne(c, filter, update)
	return err
}

func (u *UserServiceImpl) DeleteUser(c context.Context, userId string) error {
	filter := bson.D{bson.E{Key: "user_id", Value: userId}}
	result, _ := u.usercollection.DeleteOne(c, filter)
	if result.DeletedCount != 1 {
		return errors.New("no matched document found for update")
	}
	return nil
}

func (u *UserServiceImpl) Refresh(c context.Context, refreshToken string) (string, string, error) {
	claims, msg := helpers.ValidateToken(refreshToken)
	if msg != "" {
		return "", "", errors.New("error while validating token")
	}

	var user models.User
	err := u.usercollection.FindOne(c, bson.M{"user_id": claims.Uid}).Decode(&user)
	if err != nil {
		return "", "", errors.New("user not found")
	}

	NewAccess, NewRefresh, err := helpers.GenerateAllTokens(
		*user.Email,
		*user.Username,
		*user.User_type,
		user.User_id,
	)
	if err != nil {
		return "", "", err
	}

	return NewAccess, NewRefresh, err
}

func (u *UserServiceImpl) EmailExists(c context.Context, email string) (bool, error) {
	filter := bson.M{"email": email}
	count, err := u.usercollection.CountDocuments(c, filter)
	if err != nil {
		return false, err
	}
	if count < 1 {
		return false, errors.New("email doesnt exist")
	}
	return count > 0, nil
}

func (u *UserServiceImpl) SaveOTP(c context.Context, email string, otp string) error {

	record := models.OTP{
		Email:     email,
		OTP:       otp,
		ExpiresAt: time.Now().Add(5 * time.Minute),
		Used:      false,
	}

	// Remove any old OTPs for the same email
	_, _ = u.otpcollection.DeleteMany(c, bson.M{"email": email})

	_, err := u.otpcollection.InsertOne(c, record)
	return err
}

func (u *UserServiceImpl) VerifyOTP(c context.Context, email string, otp string) error {
	var record models.OTP
	err := u.otpcollection.FindOne(c, bson.M{"email": email, "otp": otp, "used": false}).Decode(&record)
	if err == mongo.ErrNoDocuments {
		return errors.New("invalid OTP")
	}

	if time.Now().After(record.ExpiresAt) {
		return errors.New("OTP expired")
	}

	// Mark OTP as used
	_, err = u.otpcollection.UpdateOne(c,
		bson.M{"_id": record.ID},
		bson.M{"$set": bson.M{"used": true}},
	)
	if err != nil {
		return err
	}

	return nil
}

func (u *UserServiceImpl) ResetPassword(c context.Context, email string, password string) error {
	hashedPassword := helpers.HashPassword(password) // implement bcrypt
	filter := bson.M{"email": email}
	update := bson.M{"$set": bson.M{"password": hashedPassword}}

	_, err := u.usercollection.UpdateOne(c, filter, update)
	if err != nil {
		return err
	}
	return nil
}
