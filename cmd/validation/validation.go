package validation

import (
	database "artifactflow.com/m/v2/cmd/database"
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"net/http"
)

type ValidationRule struct {
	ID           primitive.ObjectID     `json:"id,omitempty" bson:"_id,omitempty"`
	Name         string                 `json:"name,omitempty" bson:"name,omitempty"`
	Description  string                 `json:"description,omitempty" bson:"description,omitempty"`
	RuleStrategy string                 `json:"ruleStrategy,omitempty" bson:"ruleStrategy,omitempty"`
	RuleFamily   string                 `json:"ruleFamily,omitempty" bson:"ruleFamily,omitempty"`
	Environments map[string]interface{} `json:environments,omitempty" bson:"environments,omitempty"`
}

// Database & Collection for Validation
const validationDbName = "validationdb"
const validationColName = "validationrules"

// MongoDB client
var client, _ = database.SetupMongoDbClient()

// Create a Validation Rule
func CreateRule(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	fmt.Println("Info: Creating new Validation Rule")
	var validationRule ValidationRule
	err := json.NewDecoder(r.Body).Decode(&validationRule)

	fmt.Println("Info: Parsed into JSON")

	if err != nil {
		http.Error(w, "Unable to decode json into validationRule", 422)
		log.Println(err)
		return
	}

	collection := client.Database(validationDbName).Collection(validationColName)
	result, err := collection.InsertOne(r.Context(), validationRule)
	if err != nil {
		http.Error(w, "Unable to insert the record into the database", 417)
		log.Println(err)
		return
	}

	validationRule.ID = result.InsertedID.(primitive.ObjectID)
	json.NewEncoder(w).Encode(validationRule)
}

// Get all Validation Rule records
func GetRules(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	fmt.Println("Info: Getting all Validation Rules")
	var validationRules []ValidationRule

	collection := client.Database(validationDbName).Collection(validationColName)
	cursor, err := collection.Find(r.Context(), bson.M{})

	if err != nil {
		http.Error(w, "Unable to check Validation Rule collection with unset ID", 500)
		log.Println(err)
		return
	}

	defer cursor.Close(r.Context())
	for cursor.Next(r.Context()) {
		var validationRule ValidationRule
		cursor.Decode(&validationRule)
		validationRules = append(validationRules, validationRule)
	}

	json.NewEncoder(w).Encode(validationRules)
}
