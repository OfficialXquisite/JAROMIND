package config

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

var Client *mongo.Client
var DB *mongo.Database

func InitDatabase() {
	_ = godotenv.Load(".env")
	_ = godotenv.Load("../.env")

	connectMongoDB()
}

func connectMongoDB() {
	uri := os.Getenv("MONGO_URI")
	dbName := os.Getenv("DB_NAME")

	mongoClient, err := mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal("Failed to create mongo client:", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := mongoClient.Connect(ctx); err != nil {
		log.Fatal("Failed to connect to mongodb:", err)
	}

	if err := mongoClient.Ping(ctx, nil); err != nil {
		log.Fatal("❌ MongoDB ping failed:", err)
	}

	Client = mongoClient
	DB = mongoClient.Database(dbName)

	fmt.Println("✅ Connected to MongoDB successfully!")
}

func GetCollection(collectionName string) *mongo.Collection {
	if Client == nil {
		InitDatabase()
	}
	return DB.Collection(collectionName)
}
