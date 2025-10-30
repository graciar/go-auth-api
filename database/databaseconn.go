package database

import (
	"context"
	"log"
	"os"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func DBConnect() *mongo.Client {
	if err := godotenv.Load(".env"); err != nil {
		log.Fatal("error loading .env file")
	}

	MongoURI := os.Getenv("MONGODB_URL")

	ctx := context.TODO()
	Conn := options.Client().ApplyURI(MongoURI)

	client, err := mongo.Connect(ctx, Conn)
	if err != nil {
		log.Fatal(err)
	}

	return client
}

func OpenCollection(client *mongo.Client, collectionName string) *mongo.Collection {
	collection := client.Database("jwtauth").Collection(collectionName)
	return collection
}
