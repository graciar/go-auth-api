package routes

import (
	"go-auth/controllers"
	"go-auth/middleware"

	"github.com/gin-gonic/gin"
)

func AuthRoutes(incomingRoutes *gin.RouterGroup, uc *controllers.UserController) {
	incomingRoutes.POST("/signup", uc.Signup)
	incomingRoutes.POST("/login", uc.Login)
	incomingRoutes.POST("/forgotpassword", uc.ForgotPassword)
	incomingRoutes.POST("/verify_otp", uc.VerifyOTP)
	incomingRoutes.POST("/password/reset", middleware.ResetTokenMiddleware(), uc.ResetPassword)
	incomingRoutes.POST("/refresh", uc.Refresh)
}
