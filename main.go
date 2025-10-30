package main

import (
	"context"
	"go-auth/controllers"
	"go-auth/database"
	"go-auth/routes"
	"go-auth/services"
	"log"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	ctx := context.TODO()
	client := database.DBConnect()
	defer client.Disconnect(ctx)

	userCollectionName := os.Getenv("MONGO_USER_COLLECTION")
	otpCollectionName := os.Getenv("MONGO_OTP_COLLECTION")
	if userCollectionName == "" || otpCollectionName == "" {
		log.Fatal("MongoDB collection names not set in environment variables")
	}

	usercollection := database.OpenCollection(client, userCollectionName)
	otpcollection := database.OpenCollection(client, otpCollectionName)

	userservice := services.NewUserService(usercollection, otpcollection)
	usercontroller := controllers.NewUserController(userservice)

	server := gin.Default()
	basepath := server.Group("/v1")
	routes.AuthRoutes(basepath, &usercontroller)
	routes.UserRoutes(basepath, &usercontroller)

	log.Println("Server running on :9090")
	log.Fatal(server.Run(":9090"))
}
