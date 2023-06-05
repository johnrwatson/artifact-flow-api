package auth

import (
	database "artifactflow.com/m/v2/cmd/database"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// Database & Collection for Auth
const authDbName = "authdb"
const authTokenColName = "tokens"

var (
	googleConfig *oauth2.Config
	jwtKey       = []byte(os.Getenv("OAUTH_JWT_KEY"))
	refreshToken string // Store the refresh token
)

type ResponseStruct struct {
	Token        string `json:"api_token"`
	RefreshToken string `json:"refresh_token"`
}

// Claims represents the custom JWT claims structure
type Claims struct {
	Email        string `json:"email"`
	ExpiresAt    int64  `json:"exp"`
	RefreshToken string `json:"refresh_token"` // Add this field
	jwt.StandardClaims
}

// API Key Record
type ApiKey struct {
	ID            primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	UserID        string             `json:"userID,omitempty" bson:"userID,omitempty"`
	Key           string             `json:"apikey,omitempty" bson:"apikey,omitempty"`
	GeneratedDate time.Time          `json:"generatedDate,omitempty" bson:"generatedDate,omitempty"`
}

// Session Storage
var (
	Store  *sessions.CookieStore
	Secret = []byte(os.Getenv("OAUTH_SESSION_SECRET"))
)

// MongoDB client
var client, _ = database.SetupMongoDbClient()

func SetupOauthProvider() bool {

	// Set up Google OAuth2 configuration
	oauthClientId := os.Getenv("OAUTH_CLIENT_ID")
	oauthClientSecret := os.Getenv("OAUTH_CLIENT_SECRET")
	oauthRedirectURL := os.Getenv("OAUTH_REDIRECT_URL")
	oauthSessionSecret := os.Getenv("OAUTH_SESSION_SECRET")

	// Check if any of the variables are empty or not set
	if oauthClientId == "" || oauthClientSecret == "" || oauthRedirectURL == "" || oauthSessionSecret == "" || os.Getenv("OAUTH_JWT_KEY") == "" {
		fmt.Println("Error: One or more OAuth variables are not set, these are listed below: \n - OAUTH_CLIENT_ID\n - OAUTH_CLIENT_SECRET\n - OAUTH_REDIRECT_URL\n - OAUTH_SESSION_SECRET\n - OAUTH_JWT_KEY")
		return false
	}

	googleConfig = &oauth2.Config{
		ClientID:     oauthClientId,
		ClientSecret: oauthClientSecret,
		RedirectURL:  oauthRedirectURL,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}

	return true

}

func LoginHandler(w http.ResponseWriter, r *http.Request) {

	if googleConfig.RedirectURL != "http://localhost:80" {
		// Before making the request, disable SSL certificate validation as the ca won't be valid
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	url := googleConfig.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func CallbackHandler(w http.ResponseWriter, r *http.Request) {
	code := r.FormValue("code")

	if googleConfig.RedirectURL != "http://localhost:80" {
		// Before making the request, disable SSL certificate validation as the ca won't be valid
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	token, err := googleConfig.Exchange(oauth2.NoContext, code)
	if err != nil {
		http.Error(w, "Failed to exchange token", http.StatusBadRequest)
		fmt.Println(err)
		return
	}

	client := googleConfig.Client(oauth2.NoContext, token)
	userinfo, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		http.Error(w, "Failed to get userinfo", http.StatusBadRequest)
		return
	}

	defer userinfo.Body.Close()

	var claims Claims
	if err := json.NewDecoder(userinfo.Body).Decode(&claims); err != nil {
		http.Error(w, "Failed to decode userinfo", http.StatusBadRequest)
		return
	}

	// Set the expiration time for the claims
	claims.ExpiresAt = token.Expiry.Unix()

	// Store the refresh token
	refreshToken = token.RefreshToken

	// Set the new refresh token
	claims.RefreshToken = refreshToken

	// Create a new session for the user
	session, err := Store.Get(r, "session-name")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Store the user ID in the session
	session.Values["emailID"] = claims.Email
	session.Values["uuid"], err = generateAPIKey()
	if err != nil {
		http.Error(w, "Error generating uuid session key", 500)
		log.Println(err)
		return
	}
	session.Save(r, w)

	log.Println("Lodged user in session store:", session.Values["emailID"], session.Values["uuid"])

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := jwtToken.SignedString(jwtKey)
	if err != nil {
		http.Error(w, "Failed to sign token", http.StatusInternalServerError)
		return
	}

	responseStruct := ResponseStruct{
		Token:        tokenString,
		RefreshToken: refreshToken,
	}

	jsonData, err := json.Marshal(responseStruct)
	if err != nil {
		log.Println("Error marshaling JSON:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)

}

func RefreshAccessToken(refreshToken string) (string, error) {
	if refreshToken == "" {
		return "", errors.New("Refresh token not provided in original token")
	}

	// Create a new OAuth2 config using the existing Google config and the refresh token
	conf := &oauth2.Config{
		ClientID:     googleConfig.ClientID,
		ClientSecret: googleConfig.ClientSecret,
		RedirectURL:  googleConfig.RedirectURL,
		Scopes:       googleConfig.Scopes,
		Endpoint:     google.Endpoint,
	}

	token := &oauth2.Token{
		RefreshToken: refreshToken,
	}

	if conf.RedirectURL == "http://localhost:8000" {
		// Before making the request, disable SSL certificate validation as the ca mightn't be valid locally
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	// Use the refresh token to obtain a new access token
	newToken, err := conf.TokenSource(oauth2.NoContext, token).Token()
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	// Store the new refresh token (if provided)
	if newToken.RefreshToken != "" {
		refreshToken = newToken.RefreshToken
	}

	// Convert the claims to jwt.MapClaims
	claims := jwt.MapClaims{
		"email":         newToken.Extra("email"),
		"exp":           newToken.Expiry.Unix(),
		"refresh_token": refreshToken,
	}

	// Generate a new JWT token with the updated claims
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := jwtToken.SignedString(jwtKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func ValidateToken(w http.ResponseWriter, r *http.Request) (*Claims, error) {

	tokenString := r.Header.Get("Authorization")
	if tokenString == "" {
		return nil, errors.New("Not able to process token from http headers")
	}

	// Remove 'Bearer ' prefix from token string
	if len(tokenString) > 7 && strings.ToUpper(tokenString[0:7]) == "BEARER " {
		tokenString = tokenString[7:]
	}

	claims, err := getTokenClaims(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
	}

	// Check if the token is expired + if so generate a new one
	if time.Now().Unix() > claims.ExpiresAt {
		tokenString, err = RefreshAccessToken(claims.RefreshToken)
		if err != nil {
			http.Error(w, "Failed to create refresh token. Likely due to missing token in the claim from the original or the original is invalid.", http.StatusUnauthorized)
			fmt.Println(err)
		}
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		return nil, errors.New("Failed to parse token")
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("Invalid token")
	}

	return claims, nil
}

func getTokenClaims(w http.ResponseWriter, r *http.Request) (*Claims, error) {

	claims, err := ValidateToken(w, r)

	if err != nil {
		return nil, err
	}

	return claims, nil

}

// When requesting a oauth token we need to skip the token validation inside the middleware
// & complete the full redirect without handling it
// All other requests should continue through the middelware process as normal
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Check if authentication is disabled
		if os.Getenv("OPEN_ENDPOINTS") == "true" {
			log.Println("OPEN_ENDPOINTS set to true, authentication disabled:", r.Method, r.URL.Path)
			next.ServeHTTP(w, r)
			return
		}

		if r.URL.Path != "/health" && r.URL.Path != "/auth/login" && r.URL.Path != "/auth/callback" {
			log.Println("Received request in Authentication:", r.Method, r.URL.Path)
			_, err := getTokenClaims(w, r)

			if err.Error() == "Not able to process token from http headers" {

				// Check if there is a session for this user
				session, err := Store.Get(r, "session-name")
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				// Retrieve the user ID from the session
				_, ok := session.Values["emailID"].(string)
				if !ok {
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
					return
				}

			} else if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}

		}

		next.ServeHTTP(w, r)

	})
}

func generateAPIKey() (string, error) {
	apiKey := uuid.New().String()
	return apiKey, nil
}

func ApiKeyHandler(w http.ResponseWriter, r *http.Request) {

	claims, err := getTokenClaims(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
	}

	var apiKey ApiKey
	apiKey.UserID = claims.Email
	apiKey.Key, err = generateAPIKey()
	if err != nil {
		http.Error(w, "Error generating API key", 500)
		log.Println(err)
		return
	}
	apiKey.GeneratedDate = time.Now() // Set the Generated Date

	collection := client.Database(authDbName).Collection(authTokenColName)
	result, err := collection.InsertOne(r.Context(), apiKey)
	if err != nil {
		http.Error(w, "Unable to insert the record into the database", 417)
		log.Println(err)
		return
	}

	apiKey.ID = result.InsertedID.(primitive.ObjectID)
	json.NewEncoder(w).Encode(apiKey)
}
