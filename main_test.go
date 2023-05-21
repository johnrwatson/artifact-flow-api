package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetArtifacts(t *testing.T) {

	setupMongoDbClient()

	// Create a mock request
	req, err := http.NewRequest("GET", "/artifacts", nil)
	fmt.Println(err)
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

func TestCreateArtifact(t *testing.T) {

	setupMongoDbClient()

	// --------------------------------------------------------------------
	// Create a new artifact

	artifactId := primitive.NewObjectID()

	artifact := Artifact{
		ID:          artifactId,
		Name:        "Sample Artifact",
		Description: "This is a sample artifact.",
		Category:    "Miscellaneous",
	}

	// Convert artifact to JSON
	body, err := json.Marshal(artifact)
	if err != nil {
		t.Fatal(err)
	}

	// Create a mock request to POST the artifact into the database
	req, err := http.NewRequest("POST", "/artifacts", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}

	// Set the content type header
	req.Header.Set("Content-Type", "application/json")

	// Create a response recorder to record the response
	rr := httptest.NewRecorder()

	// Call the getArtifacts function
	createArtifact(rr, req)

	// --------------------------------------------------------------------
	// Then pull the record out of the DB
	// Could potentially just use the getArtifact endpoint but that
	// is a more complex method of testing the same thing

	ctx := context.Background()

	// Manual Test
	collection := client.Database(dbName).Collection(collName)

	err = collection.FindOne(ctx, bson.M{"_id": artifactId}).Decode(&artifact)
	if err != nil {
		log.Fatal(err)
	}

	// Assert the ID value using testify/assert
	assert.Equal(t, artifactId, artifact.ID)

}
