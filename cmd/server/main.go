package main

import (
	artifacts "artifactflow.com/m/v2/cmd/artifacts"
	auth "artifactflow.com/m/v2/cmd/auth"
	database "artifactflow.com/m/v2/cmd/database"
	supporting "artifactflow.com/m/v2/cmd/supporting"
	validation "artifactflow.com/m/v2/cmd/validation"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"log"
	"net/http"
	"os"
)

// MongoDB client
var client, _ = database.SetupMongoDbClient()

func main() {

	auth.Store = sessions.NewCookieStore(auth.Secret)

	// Initialize MongoDB client
	_, setupDatabaseClient := database.SetupMongoDbClient()

	// Setup Oauth Provider
	setupOauthProvider := auth.SetupOauthProvider()

	if setupDatabaseClient == false || setupOauthProvider == false {
		fmt.Sprintf("Error: Prereqs unable to be initialised:\n - Database Available: %v\n - Oauth Provider: %v", setupDatabaseClient, setupOauthProvider)
		os.Exit(1)
	}

	// Initialize router
	router := mux.NewRouter()

	// Register the authentication middleware function
	router.Use(auth.Middleware)

	// API endpoints for Artifacts
	router.HandleFunc("/artifacts", artifacts.CreateArtifact).Methods("POST")
	router.HandleFunc("/artifacts", artifacts.GetArtifacts).Methods("GET")
	router.HandleFunc("/artifacts/search", artifacts.SearchArtifacts).Methods("POST")
	router.HandleFunc("/artifacts/{id}", artifacts.GetArtifact).Methods("GET")
	router.HandleFunc("/artifacts/{id}", artifacts.UpdateArtifact).Methods("PUT")
	router.HandleFunc("/artifacts/{id}", artifacts.DeleteArtifact).Methods("DELETE")

	// API endpoints for Validation Rules
	router.HandleFunc("/validation/rules", validation.CreateRule).Methods("POST")
	router.HandleFunc("/validation/rules", validation.GetRules).Methods("GET")
	router.HandleFunc("/validation/rules/{id}", validation.GetRule).Methods("GET")
	router.HandleFunc("/validation/rules/search", validation.SearchRules).Methods("POST")
	router.HandleFunc("/validation/rules/{id}", validation.UpdateRule).Methods("PUT")
	router.HandleFunc("/validation/rules/{id}", validation.DeleteRule).Methods("DELETE")

	// API endpoints for Validation Rule Mappings
	router.HandleFunc("/validation/mappings", validation.CreateRuleMapping).Methods("POST")
	router.HandleFunc("/validation/mappings", validation.GetRuleMappings).Methods("GET")
	router.HandleFunc("/validation/mappings/{id}", validation.GetRuleMapping).Methods("GET")
	router.HandleFunc("/validation/mappings/search", validation.SearchRuleMappings).Methods("POST")
	router.HandleFunc("/validation/mappings/{id}", validation.UpdateRuleMapping).Methods("PUT")
	router.HandleFunc("/validation/mappings/{id}", validation.DeleteRuleMapping).Methods("DELETE")

	// Generate a Static API Key for Artifact-Flow
	router.HandleFunc("/auth/apikey", auth.ApiKeyHandler).Methods("GET")

	// Unprotected Supporting Handlers
	router.HandleFunc("/health", supporting.Health).Methods("GET")
	router.HandleFunc("/auth/login", auth.LoginHandler).Methods("GET")
	router.HandleFunc("/auth/callback", auth.CallbackHandler).Methods("GET")

	// Start the server
	log.Fatal(http.ListenAndServe(":80", router))
}
