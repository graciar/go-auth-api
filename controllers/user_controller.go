package controllers

import (
	"bytes"
	"context"
	"fmt"
	"go-auth/helpers"
	"go-auth/models"
	"go-auth/services"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

var validate = validator.New()

type UserController struct {
	userservice services.UserService
}

func NewUserController(userservice services.UserService) UserController {
	return UserController{
		userservice: userservice,
	}
}

func (u *UserController) Signup(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 100*time.Second)
	defer cancel()

	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// checks all struct fields against their validate: tags
	if validationErr := validate.Struct(user); validationErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
		return
	}

	if err := u.userservice.Signup(ctx, &user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "user account created successfully!",
	})
}

func (u *UserController) Login(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 100*time.Second)
	defer cancel()

	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	token, refreshToken, foundUser, err := u.userservice.Login(ctx, user.Email, user.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.SetCookie(
		"refresh_token",
		refreshToken,
		3600*24*7,
		"/",
		os.Getenv("COOKIE_DOMAIN"),
		true,
		true,
	)

	c.JSON(http.StatusOK, gin.H{
		"message":       "login successful",
		"token":         token,
		"refresh_token": refreshToken,
		"user":          foundUser,
	})
}

func (u *UserController) GetAll(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 100*time.Second)
	defer cancel()

	userType := c.GetString("user_type")

	if err := helpers.CheckUserType(userType, "ADMIN"); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	recordPerPage, err := strconv.Atoi(c.DefaultQuery("recordPerPage", "10"))
	if err != nil || recordPerPage < 1 {
		recordPerPage = 10
	}

	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	startIndex := (page - 1) * recordPerPage

	users, err := u.userservice.GetAll(ctx, page, recordPerPage, startIndex)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"users": users})
}

func (u *UserController) GetUser(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 100*time.Second)
	defer cancel()
	userId := c.Param("user_id")

	if err := helpers.MatchUserTypeToUid(c, userId); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	foundUser, err := u.userservice.GetUser(ctx, &userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, foundUser)
}

func (u *UserController) UpdateUser(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 100*time.Second)
	defer cancel()

	var user *models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := u.userservice.UpdateUser(ctx, user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": "update successfuly"})
}

func (u *UserController) DeleteUser(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 100*time.Second)
	defer cancel()

	userId := c.Param("user_id")
	if err := u.userservice.DeleteUser(ctx, userId); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "account has been successfuly deleted"})
}

func (u *UserController) Refresh(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 100*time.Second)
	defer cancel()
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "no refresh token"})
		return
	}

	NewAccess, NewRefresh, _ := u.userservice.Refresh(ctx, refreshToken)
	c.SetCookie(
		"refresh_token",
		NewRefresh,
		3600*24*7,
		"/",
		os.Getenv("COOKIE_DOMAIN"),
		false,
		true,
	)
	c.JSON(http.StatusOK, gin.H{
		"message":           "Tokens are refreshed",
		"new_access_token":  NewAccess,
		"new_refresh_token": NewRefresh,
	})
}

func (u *UserController) ForgotPassword(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 100*time.Second)
	defer cancel()

	var req struct {
		Email string `json:"email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := u.userservice.EmailExists(ctx, req.Email)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := godotenv.Load(".env"); err != nil {
		log.Fatal("error loading .env file")
	}

	SENDGRID_FROM_EMAIL := os.Getenv("SENDGRID_FROM_EMAIL")
	if SENDGRID_FROM_EMAIL == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Server misconfiguration: sender email not set"})
		return
	}

	SENDGRID_API_KEY := os.Getenv("SENDGRID_API_KEY")
	if SENDGRID_API_KEY == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Server misconfiguration: SendGrid API key not set"})
		return
	}

	// generate random 6-digit OTP
	otp := fmt.Sprintf("%06d", rand.Intn(1000000))
	fmt.Println("Generated OTP:", otp)

	if err := u.userservice.SaveOTP(ctx, req.Email, otp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to send OTP. Please try again later"})
		return
	}

	var body bytes.Buffer
	t, err := template.ParseFiles("template/reset_password.html")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error. Please try again later"})
		return
	}

	data := struct {
		OTP string
	}{
		OTP: otp,
	}

	if err := t.Execute(&body, data); err != nil {
		log.Println("Error executing template:", err)
		return
	}

	from := mail.NewEmail("name", SENDGRID_FROM_EMAIL)
	subject := "subject"
	to := mail.NewEmail("", req.Email)
	plainTextContent := fmt.Sprintf("Your OTP code is: %s", otp)
	htmlContent := body.String()

	message := mail.NewSingleEmail(from, subject, to, plainTextContent, htmlContent)
	client := sendgrid.NewSendClient(SENDGRID_API_KEY)

	// send email
	response, err := client.Send(message)
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Println(response.StatusCode, response.Body, response.Headers)

	c.JSON(http.StatusOK, gin.H{
		"message": "Password reset email sent successfully",
		"email":   req.Email,
	})

}

func (u *UserController) VerifyOTP(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 100*time.Second)
	defer cancel()

	var req struct {
		Email string `json:"email"`
		OTP   string `json:"otp"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := u.userservice.VerifyOTP(ctx, req.Email, req.OTP)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	resetToken, _ := helpers.GenerateResetToken(req.Email)

	c.JSON(http.StatusOK, gin.H{
		"message":     "OTP verified successfully",
		"reset_token": resetToken,
	})
}

func (u *UserController) ResetPassword(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 100*time.Second)
	defer cancel()

	var req struct {
		Email       string `json:"email"`
		NewPassword string `json:"new_password"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := helpers.IsEmailMatch(c, req.Email); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	if err := u.userservice.ResetPassword(ctx, req.Email, req.NewPassword); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password updated successfully"})
}

func Logout(c *gin.Context) {
	_, cancel := context.WithTimeout(c.Request.Context(), 100*time.Second)
	defer cancel()

	c.SetCookie(
		"refresh_token",
		"",
		-1,
		"/",
		os.Getenv("COOKIE_DOMAIN"),
		true,
		true,
	)

	c.JSON(http.StatusOK, gin.H{"message": "logged out successfully"})
}
