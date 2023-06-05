package artifacts

import (
	database "artifactflow.com/m/v2/cmd/database"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"net/http"
)

// Artifact represents a basic artifact record
type Artifact struct {
	ID               primitive.ObjectID     `json:"id,omitempty" bson:"_id,omitempty"`
	Name             string                 `json:"name,omitempty" bson:"name,omitempty"`
	Description      string                 `json:"description,omitempty" bson:"description,omitempty"`
	ArtifactType     string                 `json:"artifactType,omitempty" bson:"artifactType,omitempty"`
	ArtifactFamily   string                 `json:"artifactFamily,omitempty" bson:"artifactFamily,omitempty"`
	ArtifactMetadata map[string]interface{} `json:"artifactMetadata,omitempty" bson:"artifactMetadata,omitempty"`
}

// Database & Collection for Artifacts
const artifactDbName = "artifactdb"
const artifactColName = "artifacts"

// MongoDB client
var client, _ = database.SetupMongoDbClient()

// Create an artifact record
func CreateArtifact(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	fmt.Println("Info: Creating new Artifact")
	var artifact Artifact
	err := json.NewDecoder(r.Body).Decode(&artifact)

	fmt.Println("Info: Parsed into JSON")

	if err != nil {
		http.Error(w, "Unable to decode json into artifact", 422)
		log.Println(err)
		return
	}

	collection := client.Database(artifactDbName).Collection(artifactColName)
	result, err := collection.InsertOne(r.Context(), artifact)
	if err != nil {
		http.Error(w, "Unable to insert the record into the database", 417)
		log.Println(err)
		return
	}

	artifact.ID = result.InsertedID.(primitive.ObjectID)
	json.NewEncoder(w).Encode(artifact)
}

// Get all artifact records
func GetArtifacts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	fmt.Println("Info: Getting all Artifacts")
	var artifacts []Artifact

	collection := client.Database(artifactDbName).Collection(artifactColName)
	cursor, err := collection.Find(r.Context(), bson.M{})

	if err != nil {
		http.Error(w, "Unable to check Artifact collection with unset ID", 500)
		log.Println(err)
		return
	}

	defer cursor.Close(r.Context())
	for cursor.Next(r.Context()) {
		var artifact Artifact
		cursor.Decode(&artifact)
		artifacts = append(artifacts, artifact)
	}

	json.NewEncoder(w).Encode(artifacts)
}

// Search all artifact records
func SearchArtifacts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Parse request body
	var filter struct {
		SearchKey   string `json:"searchKey"`
		SearchValue string `json:"searchValue"`
		SearchVerb  string `json:"searchVerb"`
	}

	if err := json.NewDecoder(r.Body).Decode(&filter); err != nil {
		http.Error(w, "Invalid request body", 422)
		return
	}

	// Close the request body
	defer r.Body.Close()

	fmt.Println("Info: Searching Artifacts by " + filter.SearchKey + " where the attribute is set to " + filter.SearchValue + " with verb set to: " + filter.SearchVerb)

	collection := client.Database(artifactDbName).Collection(artifactColName)

	// Build the filter
	query := bson.M{}
	if filter.SearchKey != "" && filter.SearchValue != "" {
		if filter.SearchVerb == "contains" {
			query["$or"] = []bson.M{
				{filter.SearchKey: primitive.Regex{Pattern: filter.SearchValue, Options: "i"}},
				{"artifactMetadata." + filter.SearchKey: primitive.Regex{Pattern: filter.SearchValue, Options: "i"}},
			}
		} else {
			query["$or"] = []bson.M{
				{filter.SearchKey: filter.SearchValue},
				{"artifactMetadata." + filter.SearchKey: filter.SearchValue},
			}
		}
	}

	// Print the query to the log
	_, err := json.Marshal(query)
	if err != nil {
		log.Println("Error marshaling query to JSON:", err)
		http.Error(w, "Unable to marshal database query response to JSON", 500)
		return
	}

	// Retrieve artifacts matching the query
	var artifacts []Artifact
	cursor, err := collection.Find(r.Context(), query)
	if err != nil {
		http.Error(w, "Unable to retrieve artifacts", http.StatusInternalServerError)
		log.Println(err)
		return
	}

	defer cursor.Close(r.Context())
	for cursor.Next(r.Context()) {
		var artifact Artifact
		if err := cursor.Decode(&artifact); err != nil {
			log.Println("Error decoding artifact:", err)
			continue
		}
		artifacts = append(artifacts, artifact)
	}

	json.NewEncoder(w).Encode(artifacts)
}

// Get a specific artifact record
func GetArtifact(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	fmt.Println("Info: Getting a specific artifact record")
	params := mux.Vars(r)

	// Access the value of the "id" parameter
	idStr := params["id"]

	// Convert the string ID to an ObjectID if not already
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		http.Error(w, "Invalid Artifact ID", http.StatusBadRequest)
		log.Println(err)
		return
	}

	var artifact Artifact

	collection := client.Database(artifactDbName).Collection(artifactColName)

	err = collection.FindOne(r.Context(), bson.M{"_id": id}).Decode(&artifact)
	if err != nil {
		// Handle the error / return a response
		http.Error(w, "Unable to find artifact with that ID", http.StatusBadRequest)
		log.Println(err)
		return
	}

	json.NewEncoder(w).Encode(artifact)
}

// Update an artifact record
func UpdateArtifact(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	fmt.Println("Info: Updating a specific artifact record")

	params := mux.Vars(r)
	id, _ := primitive.ObjectIDFromHex(params["id"])

	var artifact Artifact
	_ = json.NewDecoder(r.Body).Decode(&artifact)

	collection := client.Database(artifactDbName).Collection(artifactColName)
	update := bson.M{
		"$set": bson.M{
			"name":             artifact.Name,
			"description":      artifact.Description,
			"artifactType":     artifact.ArtifactType,
			"artifactFamily":   artifact.ArtifactFamily,
			"artifactMetadata": artifact.ArtifactMetadata,
		},
	}
	_, err := collection.UpdateOne(r.Context(), bson.M{"_id": id}, update)
	if err != nil {
		fmt.Println("Error receieved")
		log.Println(err)
	}

	artifact.ID = id
	json.NewEncoder(w).Encode(artifact)
}

// Delete an artifact record
func DeleteArtifact(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	fmt.Println("Info: Deleting a specific artifact record")

	params := mux.Vars(r)
	id, _ := primitive.ObjectIDFromHex(params["id"])

	collection := client.Database(artifactDbName).Collection(artifactColName)
	_, err := collection.DeleteOne(r.Context(), bson.M{"_id": id})
	if err != nil {
		http.Error(w, "Unable to purge selected record out of the database", http.StatusBadRequest)
		log.Println(err)
	}

	json.NewEncoder(w).Encode("Artifact record deleted successfully.")

}
