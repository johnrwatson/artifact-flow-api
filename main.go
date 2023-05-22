// main.go

package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net/http"
)

// Artifact represents an artifact record
type Artifact struct {
	ID          primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Name        string             `json:"name,omitempty" bson:"name,omitempty"`
	Description string             `json:"description,omitempty" bson:"description,omitempty"`
	Category    string             `json:"category,omitempty" bson:"category,omitempty"`
}

// Connection string for MongoDB
const connectionString = "mongodb://localhost:27017"

// Database Name
const dbName = "artifactdb"

// Collection name
const collName = "artifacts"

// MongoDB client
var client *mongo.Client

// Create an artifact record
func createArtifact(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var artifact Artifact
	_ = json.NewDecoder(r.Body).Decode(&artifact)

	collection := client.Database(dbName).Collection(collName)
	result, err := collection.InsertOne(r.Context(), artifact)
	if err != nil {
		log.Fatal(err)
	}

	artifact.ID = result.InsertedID.(primitive.ObjectID)
	json.NewEncoder(w).Encode(artifact)
}

// Get all artifact records
func getArtifacts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var artifacts []Artifact

	collection := client.Database(dbName).Collection(collName)
	cursor, err := collection.Find(r.Context(), bson.M{})

    if err != nil {
        http.Error(w, "Unable to check Artifact collection with unset ID", 500)
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

// Get a specific artifact record
func getArtifact(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	params := mux.Vars(r)
	
    // Access the value of the "id" parameter
    idStr := params["id"]

    // Convert the string ID to an ObjectID if not already
    id, err := primitive.ObjectIDFromHex(idStr)
    if err != nil {
        http.Error(w, "Invalid Artifact ID", http.StatusBadRequest)
        return
    }

	var artifact Artifact

	collection := client.Database(dbName).Collection(collName)

	err = collection.FindOne(r.Context(), bson.M{"_id": id}).Decode(&artifact)
    if err != nil {
        // Handle the error / return a response
        http.Error(w, "Unable to find artifact with that ID", http.StatusBadRequest)
        return
    }

	json.NewEncoder(w).Encode(artifact)
}

// Update an artifact record
func updateArtifact(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	params := mux.Vars(r)
	id, _ := primitive.ObjectIDFromHex(params["id"])

	var artifact Artifact
	_ = json.NewDecoder(r.Body).Decode(&artifact)

	collection := client.Database(dbName).Collection(collName)
	update := bson.M{
		"$set": bson.M{
			"name":        artifact.Name,
			"description": artifact.Description,
			"category":    artifact.Category,
		},
	}
	_, err := collection.UpdateOne(r.Context(), bson.M{"_id": id}, update)
	if err != nil {
		log.Fatal(err)
	}

	artifact.ID = id
	json.NewEncoder(w).Encode(artifact)
}

// Delete an artifact record
func deleteArtifact(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	params :=

		mux.Vars(r)
	id, _ := primitive.ObjectIDFromHex(params["id"])

	collection := client.Database(dbName).Collection(collName)
	_, err := collection.DeleteOne(r.Context(), bson.M{"_id": id})
	if err != nil {
		log.Fatal(err)
	}

	json.NewEncoder(w).Encode("Artifact record deleted successfully.")

}

func setupMongoDbClient() {

	clientOptions := options.Client().ApplyURI(connectionString)
	var err error
	client, err = mongo.Connect(nil, clientOptions)
	if err != nil {
		log.Fatal(err)
	}

}

func main() {

	// Initialize MongoDB client
	setupMongoDbClient()

	// Initialize router
	router := mux.NewRouter()

	// API endpoints
	router.HandleFunc("/artifacts", createArtifact).Methods("POST")
	router.HandleFunc("/artifacts", getArtifacts).Methods("GET")
	router.HandleFunc("/artifacts/{id}", getArtifact).Methods("GET")
	router.HandleFunc("/artifacts/{id}", updateArtifact).Methods("PUT")
	router.HandleFunc("/artifacts/{id}", deleteArtifact).Methods("DELETE")

	// Start the server
	log.Fatal(http.ListenAndServe(":8000", router))
}
