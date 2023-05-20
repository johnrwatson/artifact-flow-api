package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ClearArtifacts() error {
	// Initialize MongoDB client
	clientOptions := options.Client().ApplyURI(connectionString)
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		return err
	}

	// Access the "artifacts" collection
	collection := client.Database(dbName).Collection(collName)

	// Delete all documents in the collection
	_, err = collection.DeleteMany(context.Background(), bson.D{{}})
	if err != nil {
		return err
	}

	// Close the MongoDB client connection
	err = client.Disconnect(context.Background())
	if err != nil {
		return err
	}

	return nil
}

func TestGetArtifacts(t *testing.T) {

	clientOptions := options.Client().ApplyURI(connectionString)
	var err error
	client, err = mongo.Connect(nil, clientOptions)

	err = ClearArtifacts()
	if err != nil {
		log.Fatal(err)
	}

	// Create a mock request
	req, err := http.NewRequest("GET", "/artifacts", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a response recorder to record the response
	rr := httptest.NewRecorder()

	// Call the getArtifacts function
	getArtifacts(rr, req)

	// Check the response status code
	assert.Equal(t, http.StatusOK, rr.Code)

	// Check the response body
	var artifacts []Artifact
	err = json.Unmarshal(rr.Body.Bytes(), &artifacts)
	if err != nil {
		t.Fatal(err)
	}

	// Add assertions to validate the retrieved artifacts if needed
	assert.Len(t, artifacts, 0) // Assuming there are no artifacts in the database
}
