package database

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"os"
)

// MongoDB client
var client *mongo.Client

// Database Names
const artifactDbName = "artifactdb"
const authDbName = "authdb"

// Collection names
const artifactColName = "artifacts"
const authTokenColName = "tokens"

// Get connection string
func GetConnectionString() string {
	username := os.Getenv("DB_USERNAME")
	password := os.Getenv("DB_PASSWORD")

	mongoDbConnectionString := os.Getenv("DB_CONNECTION_STRING")

	// Check if the environment variables are set
	if mongoDbConnectionString != "" {
		return mongoDbConnectionString
	}

	// Build the connection string using the username and password if available
	if username != "" && password != "" {
		return fmt.Sprintf("mongodb://%s:%s@localhost:27017", username, password)
	}

	// Return a default value if the environment variables are not set
	return "mongodb://localhost:27017"
}

func SetupMongoDbClient() (*mongo.Client, bool) {

	connectionString := GetConnectionString()

	clientOptions := options.Client().ApplyURI(connectionString)
	var err error
	client, err = mongo.Connect(nil, clientOptions)
	if err != nil {
		log.Fatal(err)
		return client, false
	}

	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Fatal("Error: Unable to connect to the MongoDB database.")
		return client, false
	}
	fmt.Println("Info: Connected to the MongoDB database successfully.")
	return client, true

}
