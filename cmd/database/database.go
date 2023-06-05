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

	mongoDbConnectionString := os.Getenv("DB_CONNECTION_STRING")

	// Check if the environment variable is set
	if mongoDbConnectionString != "" {
		return mongoDbConnectionString
	}

	// Return a default value if the environment variable is not set
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
