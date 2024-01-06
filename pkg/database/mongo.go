package database

import (
	"context"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Client *mongo.Client

func ConnectDB() {
	var err error // define error here to prevent overshadowing the global DB

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	env := os.Getenv("DATABASE_URL")

	Client, err = mongo.Connect(ctx, options.Client().ApplyURI(env))
	if err != nil {
		log.Fatal(err)
	}
}
