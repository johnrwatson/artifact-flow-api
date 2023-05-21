package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

)

func TestGetArtifacts(t *testing.T) {

	assertMongoDbClient()

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
