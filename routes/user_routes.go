package routes

import (
	"go-auth/controllers"
	"go-auth/middleware"

	"github.com/gin-gonic/gin"
)

func UserRoutes(incomingRoutes *gin.RouterGroup, uc *controllers.UserController) {
	userRoutes := incomingRoutes.Group("/user")
	userRoutes.Use(middleware.Authenticate())
	userRoutes.GET("/getuser/:user_id", uc.GetUser)
	userRoutes.GET("/getall", uc.GetAll)
	userRoutes.PATCH("/update_user", uc.UpdateUser)
	userRoutes.POST("/delete/:user_id", uc.DeleteUser)
	userRoutes.POST("/logout", controllers.Logout)
}
