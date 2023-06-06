// clear-database.go
// go run ./development-utils/clear-database.go
// & will return the number of entries from the DB it cleared, like
//
// go run ./development-utils/clear-database.go
// Deleted the following number of entries from the db: 36

package main

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"strconv"
)

// Connection string for MongoDB
const connectionString = "mongodb://localhost:27017"

// Database & Collection for Artifacts
const artifactsDbName = "artifactdb"
const artifactsColName = "artifacts"

// Database & Collection for Validation
const validationDbName = "validationdb"
const validationColName = "validationrules"

// MongoDB client
var client *mongo.Client

func ClearDatabaseArtifacts(dbName string, colName string) error {
	// Initialize MongoDB client
	clientOptions := options.Client().ApplyURI(connectionString)
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		return err
	}

	// Access the "artifacts" collection
	collection := client.Database(dbName).Collection(colName)

	// Delete all documents in the collection
	res, err := collection.DeleteMany(context.Background(), bson.D{{}})
	if err != nil {
		return err
	}

	fmt.Println("Deleted the following number of entries from the db: " + strconv.FormatInt(res.DeletedCount, 10))

	// Close the MongoDB client connection
	err = client.Disconnect(context.Background())
	if err != nil {
		return err
	}

	return nil
}

func main() {
	ClearDatabaseArtifacts(artifactsDbName, artifactsColName)
	ClearDatabaseArtifacts(validationDbName, validationColName)
}
