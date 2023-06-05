package main

import (
	auth "artifactflow.com/m/v2/cmd/auth"
	supporting "artifactflow.com/m/v2/cmd/supporting"
	artifacts "artifactflow.com/m/v2/cmd/artifacts"
	database "artifactflow.com/m/v2/cmd/database"
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

	// Register a middleware function to be used for all requests
	router.Use(auth.Middleware)

	// API endpoints for Artifacts
	router.HandleFunc("/artifacts", artifacts.CreateArtifact).Methods("POST")        
	router.HandleFunc("/artifacts", artifacts.GetArtifacts).Methods("GET")           
	router.HandleFunc("/artifacts/search", artifacts.SearchArtifacts).Methods("POST")
	router.HandleFunc("/artifacts/{id}", artifacts.GetArtifact).Methods("GET")
	router.HandleFunc("/artifacts/{id}", artifacts.UpdateArtifact).Methods("PUT")    
	router.HandleFunc("/artifacts/{id}", artifacts.DeleteArtifact).Methods("DELETE") 

	// Protected Auth Handlers
	router.HandleFunc("/auth/apikey", auth.ApiKeyHandler).Methods("GET")

	// Unprotected Supporting Handlers
    router.HandleFunc("/health", supporting.Health).Methods("GET")          
	router.HandleFunc("/auth/login", auth.LoginHandler).Methods("GET")
	router.HandleFunc("/auth/callback", auth.CallbackHandler).Methods("GET")

	// Start the server
	log.Fatal(http.ListenAndServe(":80", router))
}
