package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func DBinstance() *mongo.Client{
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	DBUserName := os.Getenv("DB_USERNAME")
	DBPassword := os.Getenv("DB_PASSWORD")

	if DBUserName == "" || DBPassword == "" {
		log.Fatal("Database credentials are missing. Set DB_USERNAME and DB_PASSWORD in environment variables.")
	}

	MongoDb := fmt.Sprintf(
		"mongodb+srv://%s:%s@restaurant-management-g.gbdz0.mongodb.net/?retryWrites=true&w=majority&appName=Restaurant-Management-Golang", 
		DBUserName, DBPassword,
	)  // you can use the online link as well
	// fmt.Println(MongoDb) // remove in production

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(MongoDb))
	if err != nil{
		log.Fatal(err)
	}


	fmt.Println("Connected to mongodb")
	return client
}

var Client *mongo.Client = DBinstance()

func OpenCollection(client *mongo.Client, collectionName string) *mongo.Collection{
	var collection *mongo.Collection = client.Database("restaurant").Collection(collectionName)

	return collection
}