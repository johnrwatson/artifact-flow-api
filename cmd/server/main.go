package main 

import (
	auth "artifactflow.com/m/v2/cmd/auth"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net/http"
	"os"
	"context"
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

// Connection string for MongoDB

// Database Name
const dbName = "artifactdb"

// Collection name
const collName = "artifacts"

// MongoDB client
var client *mongo.Client

// Get connection string
func getConnectionString() string {

	mongoDbConnectionString := os.Getenv("DB_CONNECTION_STRING")

	// Check if the environment variable is set
	if mongoDbConnectionString != "" {
		return mongoDbConnectionString
	}

	// Return a default value if the environment variable is not set
	return "mongodb://localhost:27017"
}

// Create an artifact record
func createArtifact(w http.ResponseWriter, r *http.Request) {
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

	collection := client.Database(dbName).Collection(collName)
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
func getArtifacts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	fmt.Println("Info: Getting all Artifacts")
	var artifacts []Artifact

	collection := client.Database(dbName).Collection(collName)
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

func artifactMatchesFilter(artifact Artifact, filter struct {
	SearchKey   string `json:"searchKey"`
	SearchValue string `json:"searchValue"`
}) bool {
	if filter.SearchKey != "" && filter.SearchValue != "" {
		// Check if the search key exists in the artifactMetadata field
		if value, ok := artifact.ArtifactMetadata[filter.SearchKey]; ok {
			// Compare the search value with the value in the artifactMetadata field
			return value == filter.SearchValue
		}
	}

	return false
}

// Search all artifact records
func searchArtifacts(w http.ResponseWriter, r *http.Request) {
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

	collection := client.Database(dbName).Collection(collName)

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
func getArtifact(w http.ResponseWriter, r *http.Request) {
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

	collection := client.Database(dbName).Collection(collName)

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
func updateArtifact(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	fmt.Println("Info: Updating a specific artifact record")

	params := mux.Vars(r)
	id, _ := primitive.ObjectIDFromHex(params["id"])

	var artifact Artifact
	_ = json.NewDecoder(r.Body).Decode(&artifact)

	collection := client.Database(dbName).Collection(collName)
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
func deleteArtifact(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	fmt.Println("Info: Deleting a specific artifact record")

	params := mux.Vars(r)
	id, _ := primitive.ObjectIDFromHex(params["id"])

	collection := client.Database(dbName).Collection(collName)
	_, err := collection.DeleteOne(r.Context(), bson.M{"_id": id})
	if err != nil {
		http.Error(w, "Unable to purge selected record out of the database", http.StatusBadRequest)
		log.Println(err)
	}

	json.NewEncoder(w).Encode("Artifact record deleted successfully.")

}

func setupMongoDbClient() bool {

	connectionString := getConnectionString()

	clientOptions := options.Client().ApplyURI(connectionString)
	var err error
	client, err = mongo.Connect(nil, clientOptions)
	if err != nil {
		log.Fatal(err)
		return false
	}

	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Fatal("Error: Unable to connect to the MongoDB database.")
		return false
	}
	fmt.Println("Info: Connected to the MongoDB database successfully.")
	return true

}

func main() {

	// Initialize MongoDB client
	setupDatabaseClient := setupMongoDbClient()

	// Setup Oauth Provider
	setupOauthProvider := auth.SetupOauthProvider()

	if setupDatabaseClient == false || setupOauthProvider == false {
		fmt.Sprintf("Error: Prereqs unable to be initialised:\n - Database Available: %v\n - Oauth Provider: %v", setupDatabaseClient, setupOauthProvider)
		os.Exit(1)
	}

	// Initialize router
	router := mux.NewRouter()

	// Puiblic API endpoints - Need to introduce versioning at some point
	router.HandleFunc("/artifacts", createArtifact).Methods("POST")
	router.HandleFunc("/artifacts", getArtifacts).Methods("GET")
	router.HandleFunc("/artifacts/search", searchArtifacts).Methods("POST")
	router.HandleFunc("/artifacts/{id}", getArtifact).Methods("GET")
	router.HandleFunc("/artifacts/{id}", updateArtifact).Methods("PUT")
	router.HandleFunc("/artifacts/{id}", deleteArtifact).Methods("DELETE")

	// Auth Handlers
	router.HandleFunc("/auth/login", auth.LoginHandler).Methods("GET")
	router.HandleFunc("/auth/callback", auth.CallbackHandler).Methods("GET")

	// Need auth backlogic here into the other API endpoints
	router.HandleFunc("/protected", auth.ProtectedHandler).Methods("GET")

	// Start the server
	log.Fatal(http.ListenAndServe(":8000", router))
}