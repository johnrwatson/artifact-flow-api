package validation

import (
	artifacts "artifactflow.com/m/v2/cmd/artifacts"
	database "artifactflow.com/m/v2/cmd/database"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"net/http"
)

type RuleLimit struct {
	Type  string       `json:"type" bson:"type,omitempty"`
	Value *interface{} `json:"value,omitempty" bson:"value,omitempty`
}

type ValidationRule struct {
	ID          primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Name        string             `json:"name,omitempty" bson:"name,omitempty"`               // 80percent_code_coverage
	Description string             `json:"description,omitempty" bson:"description,omitempty"` // All code must have at least 80% code coverage
	RuleFamily  string             `json:"ruleFamily,omitempty" bson:"ruleFamily,omitempty"`   // code
	RuleLimits  []RuleLimit        `json:"ruleLimits,omitempty" bson:"ruleLimits,omitempty"`   // { min: 5, max: 10 } / { value: 3 }
	RuleKey     string             `json:"ruleKey,omitempty" bson:"ruleKey,omitempty"`         // metadata.cve.high
}

type ValidationRuleMapping struct {
	ID                 primitive.ObjectID     `json:"id,omitempty" bson:"_id,omitempty"`
	RuleId             primitive.ObjectID     `json:"ruleId,omitempty" bson:"ruleId,omitempty"`                         // 647f85e6e9fd4a733a4c6b8b
	Environments       map[string]interface{} `json:"environments,omitempty" bson:"environments,omitempty"`             // { development: true }
	Enforced           bool                   `json:"enforced,omitempty" bson:"enforced,omitempty"`                     // false / true
}

type ValidationRequest struct {
	ArtifactID  string `json:"artifactId"`
	Environment string `json:"environment"`
}

// Database & Collection for Validation & Mappings
const validationDbName = "validationdb"
const validationRuleColName = "validationrules"
const validationRuleMappingColName = "validationmappings"

// MongoDB client
var client, _ = database.SetupMongoDbClient()

// --------------------------------------------
// Validation Rules
// --------------------------------------------

// Create a Validation Rule
func CreateRule(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	fmt.Println("Info: Creating new Validation Rule")
	var validationRule ValidationRule
	err := json.NewDecoder(r.Body).Decode(&validationRule)

	if err != nil {
		http.Error(w, "Unable to decode json into validationRule", 422)
		log.Println(err)
		return
	}

	collection := client.Database(validationDbName).Collection(validationRuleColName)
	result, err := collection.InsertOne(r.Context(), validationRule)
	if err != nil {
		http.Error(w, "Unable to insert the validationRule record into the database", 417)
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

	collection := client.Database(validationDbName).Collection(validationRuleColName)
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

// Search all validationRule records
func SearchRules(w http.ResponseWriter, r *http.Request) {
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

	fmt.Println("Info: Searching Validation Rules by " + filter.SearchKey + " where the attribute is set to " + filter.SearchValue + " with verb set to: " + filter.SearchVerb)

	collection := client.Database(validationDbName).Collection(validationRuleColName)

	// Build the filter
	query := bson.M{}
	if filter.SearchKey != "" && filter.SearchValue != "" {
		if filter.SearchVerb == "contains" {
			query = bson.M{filter.SearchKey: primitive.Regex{Pattern: filter.SearchValue, Options: "i"}}
		} else {
			query = bson.M{filter.SearchKey: filter.SearchValue}
		}
	}

	// Print the query to the log
	_, err := json.Marshal(query)
	if err != nil {
		log.Println("Error marshaling query to JSON:", err)
		http.Error(w, "Unable to marshal database query response to JSON", 500)
		return
	}

	// Retrieve validationRules matching the query
	var validationRules []ValidationRule
	cursor, err := collection.Find(r.Context(), query)
	if err != nil {
		http.Error(w, "Unable to retrieve validationRules", http.StatusInternalServerError)
		log.Println(err)
		return
	}

	defer cursor.Close(r.Context())
	for cursor.Next(r.Context()) {
		var validationRule ValidationRule
		if err := cursor.Decode(&validationRule); err != nil {
			log.Println("Error decoding validationRule:", err)
			continue
		}
		validationRules = append(validationRules, validationRule)
	}

	json.NewEncoder(w).Encode(validationRules)
}

// Get a specific validationRule record
func GetRule(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	fmt.Println("Info: Getting a specific validationRule record")
	params := mux.Vars(r)

	// Access the value of the "id" parameter
	idStr := params["id"]

	// Convert the string ID to an ObjectID if not already
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		http.Error(w, "Invalid validationRule ID", http.StatusBadRequest)
		log.Println(err)
		return
	}

	var validationRule ValidationRule

	collection := client.Database(validationDbName).Collection(validationRuleColName)

	err = collection.FindOne(r.Context(), bson.M{"_id": id}).Decode(&validationRule)
	if err != nil {
		// Handle the error / return a response
		http.Error(w, "Unable to find validationRule with that ID", http.StatusBadRequest)
		log.Println(err)
		return
	}

	json.NewEncoder(w).Encode(validationRule)
}

// Update an validationRule record
func UpdateRule(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	fmt.Println("Info: Updating a specific validationRule record")

	params := mux.Vars(r)
	id, _ := primitive.ObjectIDFromHex(params["id"])

	var validationRule ValidationRule
	_ = json.NewDecoder(r.Body).Decode(&validationRule)

	collection := client.Database(validationDbName).Collection(validationRuleColName)

	// This logic needs improved to update only the fields passed within the PUT, rather than assuming they were all passed
	update := bson.M{
		"$set": bson.M{
			"name":        validationRule.Name,
			"description": validationRule.Description,
			"ruleFamily":  validationRule.RuleFamily,
			"ruleLimits":  validationRule.RuleLimits,
		},
	}

	_, err := collection.UpdateOne(r.Context(), bson.M{"_id": id}, update)
	if err != nil {
		fmt.Println("Error receieved")
		log.Println(err)
	}

	validationRule.ID = id
	json.NewEncoder(w).Encode(validationRule)
}

// Delete an validationRule record
func DeleteRule(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	fmt.Println("Info: Deleting a specific validationRule record")

	params := mux.Vars(r)
	id, _ := primitive.ObjectIDFromHex(params["id"])

	collection := client.Database(validationDbName).Collection(validationRuleColName)
	_, err := collection.DeleteOne(r.Context(), bson.M{"_id": id})
	if err != nil {
		http.Error(w, "Unable to purge selected record out of the database", http.StatusBadRequest)
		log.Println(err)
	}

	json.NewEncoder(w).Encode("validationRule record deleted successfully.")

}

// --------------------------------------------
// Validation Rule Mappings
// --------------------------------------------

// Create a validationRuleMapping
func CreateRuleMapping(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	fmt.Println("Info: Creating new Validation Mapping")
	var validationRuleMapping ValidationRuleMapping
	err := json.NewDecoder(r.Body).Decode(&validationRuleMapping)

	if err != nil {
		http.Error(w, "Unable to decode json into validationRuleMapping", 422)
		log.Println(err)
		return
	}

	// Check if there is a record for the chosen validationrule within the validationrules database
	fmt.Println(validationDbName, validationRuleColName, validationRuleMapping.RuleId)
	exists, err := existenceValidator(validationDbName, validationRuleColName, validationRuleMapping.RuleId)
	if err != nil {
		http.Error(w, "Error checking validationRule collection", 500)
		log.Println(err)
		return
	}
	if !exists {
		http.Error(w, "Validation Rule not found", 404)
		return
	}

	collection := client.Database(validationDbName).Collection(validationRuleMappingColName)
	result, err := collection.InsertOne(r.Context(), validationRuleMapping)
	if err != nil {
		http.Error(w, "Unable to insert the validationRuleMapping record into the database", 417)
		log.Println(err)
		return
	}

	validationRuleMapping.ID = result.InsertedID.(primitive.ObjectID)
	json.NewEncoder(w).Encode(validationRuleMapping)
}

// Get all validationRuleMapping records
func GetRuleMappings(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	fmt.Println("Info: Getting all Validation Rules")
	var validationRuleMappings []ValidationRuleMapping

	collection := client.Database(validationDbName).Collection(validationRuleMappingColName)
	cursor, err := collection.Find(r.Context(), bson.M{})

	if err != nil {
		http.Error(w, "Unable to check Validation Rule collection with unset ID", 500)
		log.Println(err)
		return
	}

	defer cursor.Close(r.Context())
	for cursor.Next(r.Context()) {
		var validationRuleMapping ValidationRuleMapping
		cursor.Decode(&validationRuleMapping)
		validationRuleMappings = append(validationRuleMappings, validationRuleMapping)
	}

	json.NewEncoder(w).Encode(validationRuleMappings)
}

// Search all validationRuleMapping records
func SearchRuleMappings(w http.ResponseWriter, r *http.Request) {
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

	fmt.Println("Info: Searching Validation Rule Mappings by " + filter.SearchKey + " where the attribute is set to " + filter.SearchValue + " with verb set to: " + filter.SearchVerb)

	collection := client.Database(validationDbName).Collection(validationRuleMappingColName)

	// Build the filter
	query := bson.M{}
	if filter.SearchKey != "" && filter.SearchValue != "" {
		if filter.SearchVerb == "contains" {
			query = bson.M{filter.SearchKey: primitive.Regex{Pattern: filter.SearchValue, Options: "i"}}
		} else {
			query = bson.M{filter.SearchKey: filter.SearchValue}
		}
	}

	// Print the query to the log
	_, err := json.Marshal(query)
	if err != nil {
		log.Println("Error marshaling query to JSON:", err)
		http.Error(w, "Unable to marshal database query response to JSON", 500)
		return
	}

	// Retrieve validationRuleMappings matching the query
	var validationRuleMappings []ValidationRuleMapping
	cursor, err := collection.Find(r.Context(), query)
	if err != nil {
		http.Error(w, "Unable to retrieve validationRuleMappings", http.StatusInternalServerError)
		log.Println(err)
		return
	}

	defer cursor.Close(r.Context())
	for cursor.Next(r.Context()) {
		var validationRuleMapping ValidationRuleMapping
		if err := cursor.Decode(&validationRuleMapping); err != nil {
			log.Println("Error decoding validationRuleMapping:", err)
			continue
		}
		validationRuleMappings = append(validationRuleMappings, validationRuleMapping)
	}

	json.NewEncoder(w).Encode(validationRuleMappings)
}

// Get a specific validationRuleMapping record
func GetRuleMapping(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	fmt.Println("Info: Getting a specific validationRuleMapping record")
	params := mux.Vars(r)

	// Access the value of the "id" parameter
	idStr := params["id"]

	// Convert the string ID to an ObjectID if not already
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		http.Error(w, "Invalid validationRuleMapping ID", http.StatusBadRequest)
		log.Println(err)
		return
	}

	var validationRuleMapping ValidationRuleMapping

	collection := client.Database(validationDbName).Collection(validationRuleMappingColName)

	err = collection.FindOne(r.Context(), bson.M{"_id": id}).Decode(&validationRuleMapping)
	if err != nil {
		// Handle the error / return a response
		http.Error(w, "Unable to find validationRuleMapping with that ID", http.StatusBadRequest)
		log.Println(err)
		return
	}

	json.NewEncoder(w).Encode(validationRuleMapping)
}

// Update an validationRuleMapping record
func UpdateRuleMapping(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	fmt.Println("Info: Updating a specific validationRuleMapping record")

	params := mux.Vars(r)
	id, _ := primitive.ObjectIDFromHex(params["id"])

	var validationRuleMapping ValidationRuleMapping
	_ = json.NewDecoder(r.Body).Decode(&validationRuleMapping)

	collection := client.Database(validationDbName).Collection(validationRuleMappingColName)

	// This logic needs improved to update only the fields passed within the PUT, rather than assuming they were all passed
	update := bson.M{
		"$set": bson.M{
			"ruleId":       validationRuleMapping.ID,
			"environments": validationRuleMapping.Environments,
			"enforced":     validationRuleMapping.Enforced,
		},
	}

	type ValidationRuleMapping struct {
		ID                 primitive.ObjectID     `json:"id,omitempty" bson:"_id,omitempty"`
		RuleId             primitive.ObjectID     `json:"ruleId,omitempty" bson:"ruleId,omitempty"`             // 647f85e6e9fd4a733a4c6b8b
		Environments       map[string]interface{} `json:"environments,omitempty" bson:"environments,omitempty"` // { development: true, preproduction: false, production: false }
		Enforced           bool                   `json:"enforced,omitempty" bson:"enforced,omitempty"`         // false / true
	}

	_, err := collection.UpdateOne(r.Context(), bson.M{"_id": id}, update)
	if err != nil {
		fmt.Println("Error receieved")
		log.Println(err)
	}

	validationRuleMapping.ID = id
	json.NewEncoder(w).Encode(validationRuleMapping)
}

// Delete an validationRuleMapping record
func DeleteRuleMapping(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	fmt.Println("Info: Deleting a specific validationRuleMapping record")

	params := mux.Vars(r)
	id, _ := primitive.ObjectIDFromHex(params["id"])

	collection := client.Database(validationDbName).Collection(validationRuleMappingColName)
	_, err := collection.DeleteOne(r.Context(), bson.M{"_id": id})
	if err != nil {
		http.Error(w, "Unable to purge selected record out of the database", http.StatusBadRequest)
		log.Println(err)
	}

	json.NewEncoder(w).Encode("validationRuleMapping record deleted successfully.")

}

// --------------------------------------------
// Validation of Artifacts
// --------------------------------------------

// Validate whether an Artifact is suitable for Environment
func ValidateArtifact(w http.ResponseWriter, r *http.Request) {
	var req ValidationRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Retrieve the artifact and validation rules from MongoDB
	artifact, err := getArtifactByID(req.ArtifactID)
	if err != nil {
		http.Error(w, "Failed to retrieve artifact", http.StatusInternalServerError)
		return
	}

	//fmt.Println("Artifact Record:")
	//fmt.Printf("%+v\n", artifact)

	rules, err := getValidationRulesForEnvironment(req.Environment)
	if err != nil {
		http.Error(w, "Failed to retrieve validation rules", http.StatusInternalServerError)
		fmt.Println(err)
		return
	}

	//fmt.Println("Rules below:")
	//fmt.Printf("%+v\n", rules)

	// Perform validation check
	passesValidation, errorMap  := validateArtifactAgainstRules(artifact, rules)

	// Return the result
	result := struct {
		ArtifactId        string            `json:"artifactId"`
		PassesValidation  bool              `json:"passesValidation"`
		Environment       string            `json:"environment"`
		Violations        map[string]string `json:"violations,omitempty"`
	}{
		ArtifactId:        req.ArtifactID,
		PassesValidation:  passesValidation,
		Environment:       req.Environment,
	}

	if !passesValidation {
		violations := make(map[string]string)
		for ruleKey, err := range errorMap {
			violations[ruleKey] = err.Error()
		}
		result.Violations = violations
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func getArtifactByID(artifactID string) (*artifacts.Artifact, error) {
	// Implementation to retrieve the artifact by ID from MongoDB
	collection := client.Database(artifacts.ArtifactDbName).Collection(artifacts.ArtifactColName)

	// Convert the string ID to an ObjectID if not already
	id, err := primitive.ObjectIDFromHex(artifactID)
	if err != nil {
		return nil, err
	}

	var artifact artifacts.Artifact

	err = collection.FindOne(context.TODO(), bson.M{"_id": id}).Decode(&artifact)
	if err != nil {
		return nil, err
	}

	return &artifact, nil
}

func getValidationRulesForEnvironment(environment string) ([]ValidationRule, error) {
	// Implementation to retrieve the validation rules for the given environment from MongoDB
	collection := client.Database(validationDbName).Collection(validationRuleMappingColName)

	// Build the filter
	query := bson.M{"environments." + environment: true}

	var validationRules []ValidationRule

	cursor, err := collection.Find(context.TODO(), query)
	if err != nil {
		return nil, err
	}

	defer cursor.Close(context.TODO())

	for cursor.Next(context.TODO()) {
		var validationRuleMapping ValidationRuleMapping
		if err := cursor.Decode(&validationRuleMapping); err != nil {
			return nil, err
		}

		ruleID := validationRuleMapping.RuleId

		rule, err := getRuleByID(ruleID)
		if err != nil {
			return nil, err
		}

		validationRules = append(validationRules, *rule)
	}

	return validationRules, nil
}

func getRuleByID(ruleID primitive.ObjectID) (*ValidationRule, error) {
	// Implementation to retrieve the validation rule by ID from MongoDB
	collection := client.Database(validationDbName).Collection(validationRuleColName)

	var validationRule ValidationRule

	err := collection.FindOne(context.TODO(), bson.M{"_id": ruleID}).Decode(&validationRule)
	if err != nil {
		return nil, err
	}

	return &validationRule, nil
}

func validateArtifactAgainstRules(artifact *artifacts.Artifact, rules []ValidationRule) (bool, map[string]error) {

	errors := make(map[string]error)

	for _, rule := range rules {
		for _, lim := range rule.RuleLimits {
			fmt.Println(rule, ":", lim)
			if err := lim.Evaluate(*artifact, rule.RuleKey); len(err.Problems) != 0 {
				errors[rule.RuleKey] = err
			}
		}
	}

	return len(errors) == 0, errors

}

// ------------------------------------------------------------------------------------------
// Supporting Functions
// ------------------------------------------------------------------------------------------
// Function to check the existence of a record in a collection
func existenceValidator(dbName string, colName string, id primitive.ObjectID) (bool, error) {
	collection := client.Database(dbName).Collection(colName)
	//fmt.Println("searching for")
	//fmt.Println(id)
	count, err := collection.CountDocuments(context.Background(), bson.M{"_id": id})
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func lookup(artifact artifacts.Artifact, metadata map[string]any, keys []string) (any, bool) {

	//fmt.Println("------------SEARCHING FOR-----------------------------------------")
	//fmt.Println(metadata, keys)

	// If the key exists against the root object & isn't within the metadata lookup
	if len(keys) == 1 && keys[0] != "artifactMetadata" {

		// First check if child key in metadata
		val, ok := metadata[keys[0]] // As it needs to skip the first key in this case
		//fmt.Println("Metadata record found to be:")
		//fmt.Println(metadata)
		//fmt.Println("Is the key found?", val, ok)
		if len(keys) == 1 && ok {
			return val, ok
		}

		// If not, then it must be a root key
		// Create a temporary map to check if the requested key exists in the struct
		// This isn't great, but it'll work for now
		rootFields := map[string]interface{}{
			"id":             artifact.ID,
			"name":           artifact.Name,
			"description":    artifact.Description,
			"artifactType":   artifact.ArtifactType,
			"artifactFamily": artifact.ArtifactFamily,
		}

		val, ok = rootFields[keys[0]]
		//fmt.Println("Scanning root keys", val, ok, "for", keys[0])
		if ok {
			return val, ok
		}

		return nil, false

		// If the key is proposed to be within the metadata block, pass back it's children
	} else if keys[0] == "artifactMetadata" {

		//fmt.Println("Searching inside artifactmetadata", keys[0], metadata)

		// base cases - this is a nil map, or no keys were given:
		if metadata == nil || len(keys) == 0 {
			return nil, false
		}

		// base cases - this is the last key, or this key is not found:
		val, ok := metadata[keys[1]] // As it needs to skip the first key in this case
		//fmt.Println("Is the key found?", val, ok)
		//fmt.Println(len(keys), keys)
		if len(keys) == 1 || !ok {
			return val, ok
		}

		// base case: the child is not a map:
		child, ok := val.(map[string]any)
		if !ok {
			return nil, false
		}
		//fmt.Println("Passing the lookup onto the next metadata key section", child, ok)
		//fmt.Println("----------PASSING ONTO NEXT LOOP SECTION------------------------------------------")
		// recursive case: ask the child to get the value
		return lookup(artifact, child, keys[2:])
	} else {
		//fmt.Println("Searching deeper inside artifactmetadata", keys[0], metadata)

		// base cases - this is a nil map, or no keys were given:
		if metadata == nil || len(keys) == 0 {
			return nil, false
		}

		// base cases - this is the last key, or this key is not found:
		val, ok := metadata[keys[0]] // As it needs to skip the first key in this case

		// -----LOGGING HELPER--------------------
		// --------------------------------------------
		subkeys := make([]string, len(metadata))
		i := 0
		for k := range metadata {
			subkeys[i] = k
		  i++
		}
		//fmt.Println("All keys shown below:")
		//fmt.Println(subkeys)
		//fmt.Println("Looking for key", keys[0])
		// --------------------------------------------

		//fmt.Println("Is the key found?", val, ok)
		//fmt.Println(len(keys), keys)
		if len(keys) == 1 || !ok {
			return val, ok
		}

		// base case: the child is not a map:
		child, ok := val.(map[string]any)
		if !ok {
			return nil, false
		}
		//fmt.Println("Passing the lookup onto the next metadata key section", child, ok)
		//fmt.Println("----------PASSING ONTO NEXT LOOP SECTION------------------------------------------")
		// recursive case: ask the child to get the value
		return lookup(artifact, child, keys[1:])
	}

	return nil, false

}
