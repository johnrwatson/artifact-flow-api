package main

import (
	"bytes"
	//"context"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	//"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	//"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetArtifacts(t *testing.T) {

	// Test for listing all artifacts

	setupMongoDbClient()

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

	// Check if the artifact slice was returned correctly
	assert.True(t, len(artifacts) >= 0)

	if len(artifacts) > 0 {
		// Assert the Name as a test artifat if any returned from GET Endpoint
		assert.Equal(t, "Sample Artifact", artifacts[0].Name)
	} else {
		fmt.Println("Info: Length of artifacts slice was returned as zero")
	}

}

func TestArtifactCRUD(t *testing.T) {

	setupMongoDbClient()

	// --------------------------------------------------------------------
	// [C] CREATE a new artifact

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
	// [R] Then READ the record using the GET endpoint

	// Create a mock request to POST the artifact into the database
	req, err = http.NewRequest("GET", "/artifacts/" + artifactId.Hex(), nil)
	if err != nil {
		t.Fatal(err)
	}

	// Set the content type header
	req.Header.Set("Content-Type", "application/json")

	// Refresh the response recorder to record the response
	rr = httptest.NewRecorder()

	// Call the getArtifacts function
	getArtifact(rr, req)

	// Assert the ID value using testify/assert
	assert.Equal(t, "Sample Artifact", artifact.Name)

    // --------------------------------------------------------------------
	// [U] Then attempt to UPDATE the record with a new name

	artifact = Artifact{
		ID:          artifactId,
		Name:        "Sample Artifact Updated",
		Description: "This is an updated sample artifact.",
		Category:    "Miscellaneous",
	}

	// Convert artifact to JSON
	body, err = json.Marshal(artifact)
	if err != nil {
		t.Fatal(err)
	}

	// Create a mock request to POST the artifact into the database
	req, err = http.NewRequest("PUT", "/artifacts/" + artifactId.Hex(), bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}

	// Set the content type header
	req.Header.Set("Content-Type", "application/json")

	// Create a response recorder to record the response
	rr = httptest.NewRecorder()

	// Call the getArtifacts function
	updateArtifact(rr, req)

	// Assert the ID value using testify/assert
	assert.Equal(t, "Sample Artifact Updated", artifact.Name)

    // --------------------------------------------------------------------
	// [D] Then attempt to DELETE the record with a new name

	// Convert artifact to JSON
	body, err = json.Marshal(artifact)
	if err != nil {
		t.Fatal(err)
	}

	// Create a mock request to POST the artifact into the database
	req, err = http.NewRequest("DELETE", "/artifacts/" + artifactId.Hex(), bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}

	// Set the content type header
	req.Header.Set("Content-Type", "application/json")

	// Create a response recorder to record the response
	rr = httptest.NewRecorder()

	// Call the getArtifacts function
	deleteArtifact(rr, req)

    // --------------------------------------------------------------------
   	// [R] Then READ the - should be - missing record using the GET endpoint

	// Create a mock request to POST the artifact into the database
	req, err = http.NewRequest("GET", "/artifacts/" + artifactId.Hex(), nil)
	if err != nil {
		t.Fatal(err)
	}

	// Set the content type header
	req.Header.Set("Content-Type", "application/json")

	// Refresh the response recorder to record the response
	rr = httptest.NewRecorder()

	// Call the getArtifacts function
	getArtifact(rr, req)

	assert.Equal(t, "Invalid Artifact ID\n", rr.Body.String())

}
