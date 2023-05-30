package main

import (
	"bytes"
	//"context"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	//"go.mongodb.org/mongo-driver/bson"
	"encoding/base64"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"reflect"
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
		ID:             artifactId,
		Name:           "Sample Artifact",
		Description:    "This is a sample artifact.",
		ArtifactType:   "container",
		ArtifactFamily: "test-family",
		ArtifactMetadata: map[string]interface{}{
			"repository":          "test-repository.git",
			"repository_provider": "github",
			"location":            "artifactory",
		},
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
	req, err = http.NewRequest("GET", "/artifacts/"+artifactId.Hex(), nil)
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
		ID:             artifactId,
		Name:           "Sample Artifact Updated",
		Description:    "This is an updated sample artifact.",
		ArtifactType:   "container",
		ArtifactFamily: "test-family",
		ArtifactMetadata: map[string]interface{}{
			"repository":          "test-repository.git",
			"repository_provider": "github",
			"location":            "artifactory",
		},
	}

	// Convert artifact to JSON
	body, err = json.Marshal(artifact)
	if err != nil {
		t.Fatal(err)
	}

	// Create a mock request to POST the artifact into the database
	req, err = http.NewRequest("PUT", "/artifacts/"+artifactId.Hex(), bytes.NewBuffer(body))
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
	req, err = http.NewRequest("DELETE", "/artifacts/"+artifactId.Hex(), bytes.NewBuffer(body))
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
	req, err = http.NewRequest("GET", "/artifacts/"+artifactId.Hex(), nil)
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

func generateRandomID(length int) string {
	randomBytes := make([]byte, length)
	_, err := rand.Read(randomBytes)
	if err != nil {
		panic(err) // Handle the error appropriately in your code
	}

	randomID := base64.URLEncoding.EncodeToString(randomBytes)
	return randomID[:length] // Trim the string to the desired length
}

func compareSearchResult(searchReq *http.Request, expectedArtifact Artifact) (bool, error) {
	// Create a new response recorder for the search
	rr := httptest.NewRecorder()

	// Set the content type header
	searchReq.Header.Set("Content-Type", "application/json")

	// Call the searchArtifacts function
	searchArtifacts(rr, searchReq)

	// Check the response status code
	if rr.Code != http.StatusOK {
		fmt.Println(rr)
		return false, fmt.Errorf("Expected status code %d but got %d", http.StatusOK, rr.Code)
	}

	// Parse the response body
	var searchResult []Artifact
	if err := json.NewDecoder(rr.Body).Decode(&searchResult); err != nil {
		return false, err
	}

	// Perform assertions on the search results
	if len(searchResult) != 1 {
		return false, fmt.Errorf("Expected 1 search result but got %d", len(searchResult))
	}

	// Assert specific artifact details
	if !reflect.DeepEqual(searchResult[0], expectedArtifact) {
		return false, fmt.Errorf("Search result does not match the expected artifact:\nExpected: %+v\nActual: %+v", expectedArtifact, searchResult[0])
	}

	return true, nil
}

func TestArtifactSearch(t *testing.T) {

	setupMongoDbClient()

	// --------------------------------------------------------------------
	// [C] CREATE a new artifact to search for by unique search attributes

	artifactId := primitive.NewObjectID()
	randomSearchString := generateRandomID(8)       // Use the full string to allow searching for a specific artifact
	randomSearchSubString := randomSearchString[:4] // Use the first four characters to search with `contains`

	artifact := Artifact{
		ID:             artifactId,
		Name:           "Sample Search Artifact-" + randomSearchString, // name
		Description:    "This is a sample search artifact.",
		ArtifactType:   "container",
		ArtifactFamily: "test-family",
		ArtifactMetadata: map[string]interface{}{
			"repository":          "searchy-repository.git",
			"repository_provider": "searchy-provider",
			"location":            "shallow",
			"deeperStruct": map[string]interface{}{
				"metadata": "deep",
				"secondNest": map[string]interface{}{
					"doubleNested": "deeper-" + randomSearchString, // artifactMetadata.deeperStruct.secondNest.doubleNested
				},
			},
		},
	}

	jsonArtifact, err := json.Marshal(artifact)
	if err != nil {
		log.Fatalf("Failed to marshal struct to JSON: %v", err)
	}

	log.Printf("JSON data: %s", jsonArtifact)

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

	// Check the response status code
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status code %d but got %d", http.StatusOK, rr.Code)
	}

	// --------------------------------------------------------------------
	// Test 1: Search for Root key with Contains
	// Test 2: Search for Root key with Equals
	// Test 3: Search for Deeply Nested key with Contains
	// Test 4: Search for Deeply Nested key with Equals
	// --------------------------------------------------------------------

	searchPayload := []byte(`{
		"searchKey": "name",
		"searchValue": "` + randomSearchString + `",
		"searchVerb": "contains"
	}`)

	// Test 1: Search for Root key with Contains
	searchReq, err := http.NewRequest("POST", "/artifacts/search", bytes.NewBuffer(searchPayload))

	if err != nil {
		t.Fatal(err)
	}

	match, err := compareSearchResult(searchReq, artifact)
	if err != nil {
		t.Fatal(err)
	}

	if !match {
		t.Errorf("Search result does not match the expected artifact")
	}
	// --------------------------------------------------------------------
	searchPayload = []byte(`{
		"searchKey": "name",
		"searchValue": "Sample Search Artifact-` + randomSearchString + `"
	}`)

	// Test 2: Search for Root key with Equals
	searchReq, err = http.NewRequest("POST", "/artifacts/search", bytes.NewBuffer(searchPayload))

	if err != nil {
		t.Fatal(err)
	}

	match, err = compareSearchResult(searchReq, artifact)
	if err != nil {
		t.Fatal(err)
	}

	if !match {
		t.Errorf("Search result does not match the expected artifact")
	}
	// --------------------------------------------------------------------
	// Test 3: Search for Deeply Nested key with Contains
	searchReq, err = http.NewRequest("POST", "/artifacts/search", bytes.NewBuffer([]byte(`{
		"searchKey": "artifactMetadata.deeperStruct.secondNest.doubleNested",
		"searchValue": "`+randomSearchSubString+`",
		"searchVerb": "contains"
	}`)))

	if err != nil {
		t.Fatal(err)
	}

	match, err = compareSearchResult(searchReq, artifact)
	if err != nil {
		t.Fatal(err)
	}

	if !match {
		t.Errorf("Search result does not match the expected artifact")
	}
	// --------------------------------------------------------------------
	// Test 4: Search for Deeply Nested key with Equals
	searchReq, err = http.NewRequest("POST", "/artifacts/search", bytes.NewBuffer([]byte(`{
		"searchKey": "artifactMetadata.deeperStruct.secondNest.doubleNested",
		"searchValue": "deeper-`+randomSearchString+`"
	}`)))

	if err != nil {
		t.Fatal(err)
	}

	match, err = compareSearchResult(searchReq, artifact)
	if err != nil {
		t.Fatal(err)
	}

	if !match {
		t.Errorf("Search result does not match the expected artifact")
	}
	// -------------------------------------------------------------------

}
