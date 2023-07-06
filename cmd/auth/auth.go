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
	"encoding/base64"

	"context"
	"google.golang.org/api/idtoken"

)

// Database & Collection for Auth
const authDbName = "authdb"
const authTokenColName = "tokens"

var (
	googleConfig *oauth2.Config
	jwtKey       = []uint8(os.Getenv("OAUTH_JWT_KEY"))
	refreshToken string // Store the refresh token
	Store  *sessions.CookieStore
	Secret       = []byte(os.Getenv("OAUTH_SESSION_SECRET"))
)

type ResponseStruct struct {
	Token        string `json:"api_token"`
	RefreshToken string `json:"refresh_token"`
}

// Claims represents the custom JWT claims structure
type CustomClaims struct {
	Email        string `json:"email"`
	ExpiresAt    int64  `json:"exp"`
	RefreshToken string `json:"refresh_token"`
	jwt.StandardClaims
}

// API Key Record
type ApiKey struct {
	ID            primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	UserID        string             `json:"userID,omitempty" bson:"userID,omitempty"`
	Key           string             `json:"apikey,omitempty" bson:"apikey,omitempty"`
	GeneratedDate time.Time          `json:"generatedDate,omitempty" bson:"generatedDate,omitempty"`
}

type PublicKeyInfo struct {
	Use       string `json:"use"`
	Kid       string `json:"kid"`
	Kty       string `json:"kty"`
	Alg       string `json:"alg"`
	N         string `json:"n"`
	E         string `json:"e"`
	Algorithm string
	PublicKey []byte  
}

type KeyResponse struct {
	Keys []PublicKeyInfo `json:"keys"`
}

// MongoDB client
var client, _ = database.SetupMongoDbClient()

func SetupOauthProvider() bool {

	// Set up Google OAuth2 configuration
	oauthClientId := os.Getenv("OAUTH_CLIENT_ID")
	oauthClientSecret := os.Getenv("OAUTH_CLIENT_SECRET")
	oauthRedirectURL := os.Getenv("OAUTH_REDIRECT_URL")
	oauthSessionSecret := os.Getenv("OAUTH_SESSION_SECRET")

	// Check if any of the variables are empty or not set
	if oauthClientId == "" || oauthClientSecret == "" || oauthRedirectURL == "" || oauthSessionSecret == "" || jwtKey == nil {
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

	if googleConfig.RedirectURL == "http://localhost:80/auth/callback" {
		// Before making the request, disable SSL certificate validation as the ca won't be valid
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	url := googleConfig.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func CallbackHandler(w http.ResponseWriter, r *http.Request) {
	code := r.FormValue("code")

	if googleConfig.RedirectURL == "http://localhost:80/auth/callback" {
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

	var claims CustomClaims
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

	log.Println("Lodged user in session store:", session.Values["GeneratedDate"], session.Values["uuid"])

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
		return "", errors.New("Refresh token not provided in original token, unable to refresh authentication token automatically")
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

	if conf.RedirectURL == "http://localhost:80/auth/callback" {
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

func getTokenClaims(w http.ResponseWriter, r *http.Request) (*CustomClaims, error) {

	if googleConfig.RedirectURL == "http://localhost:80/auth/callback" {
		// Before making the request, disable SSL certificate validation as the ca won't be valid
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	tokenString := r.Header.Get("Authorization")
	if tokenString == "" {
		return nil, errors.New("No token found in Authorization header")
	}

	// Remove 'Bearer ' prefix from token string
	if len(tokenString) > 7 && strings.ToUpper(tokenString[0:7]) == "BEARER " {
		tokenString = tokenString[7:]
	}

	// Getting the algorithm of the jwt signature
	fmt.Printf("Type of jwtKey: %T\n", jwtKey)
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		fmt.Println("Invalid token format")
	}
	
	// Decode the header
	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		fmt.Println("Error decoding token header:", err)
	}
	
	// Parse the header JSON
	var header map[string]interface{}
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		fmt.Println("Error parsing token header:", err)
	}
	
	// Extract the algorithm from the header
	alg, ok := header["alg"].(string)
	if !ok {
		fmt.Println("Invalid or missing signing algorithm in token header")
	}
	
	fmt.Println("Signing algorithm:", alg)
	var token *jwt.Token

	switch alg {
	case "HS256":
		fmt.Println("Handling HS256 algorithm")
		token, err = jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})
	case "RS256":
		fmt.Println("Handling RS256 algorithm")
		// Retrieve the public keys from the URL
		
		audience := "160220461475-r2h1nktt3nagkub9jdl6tds2ntoh8gsn.apps.googleusercontent.com" // This is the given audience of the FE
	
		// Create a context
		ctx := context.Background()
	
		// Verify and parse the JWT
		payload, err := idtoken.Validate(ctx, tokenString, audience)
		if err != nil {
			return nil, err
		}
		
		// Extract the claims from the payload
		standardClaims := payload.Claims
		for key, value := range standardClaims {
			fmt.Printf("Key: %s, Value: %v\n", key, value)
		}
		issuer := standardClaims["iss"].(string)
		subject := standardClaims["sub"].(string)
		email := standardClaims["email"].(string)
	
		fmt.Println("Issuer:", issuer)
		fmt.Println("Subject:", subject)
		fmt.Println("Email:", email)


		claims := &CustomClaims{
			Email:        standardClaims["email"].(string),
			ExpiresAt:    12345,
			RefreshToken: "", // Set the refresh token value as needed
			StandardClaims: jwt.StandardClaims{
				Audience:  standardClaims["aud"].(string),
				ExpiresAt: standardClaims["exp"].(int64),
				Id:        "123",
				IssuedAt:  standardClaims["iat"].(int64),
				Issuer:    standardClaims["iss"].(string),
				NotBefore: standardClaims["nbf"].(int64),
				Subject:   standardClaims["sub"].(string),
			},
		}

		if err != nil {
			return nil, errors.New("JWT parsing failed:")
		}

		return claims, nil


	default:
		fmt.Println("")
		return nil, errors.New("Unsupported JWT Algorithm: " + alg)
	}

	if err != nil {
		fmt.Println(err)
		return nil, errors.New("Failed to parse token")
	}

	fmt.Println("Checking the claims on the token")
	claims, ok := token.Claims.(*CustomClaims)
	fmt.Println("Finished checking claims on token")
	if !ok || !token.Valid {

		return nil, errors.New("Invalid token")
	}

	// Check if the token is expired + if so generate a new one
	//if time.Now().Unix() > claims.ExpiresAt {
	//	tokenString, err = RefreshAccessToken(claims.RefreshToken)
	//	if err != nil {
	//		return nil, err
	//	}
	//}

	return claims, nil
}

// When requesting a oauth token we need to skip the token validation inside the middleware
// & complete the full redirect without handling it
// All other requests should continue through the middelware process as normal
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// CORS Policy for Frontend Access
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")

		// Check if authentication is disabled
		if os.Getenv("OPEN_ENDPOINTS") == "true" {
			log.Println("Warning: INSECURE API - OPEN_ENDPOINTS variable set to true, authentication is disabled for all endpoints:", r.Method, r.URL.Path)
			next.ServeHTTP(w, r)
			return
		}

		if r.URL.Path != "/health" && r.URL.Path != "/auth/login" && r.URL.Path != "/auth/callback" {

			fmt.Println("Getting token claims")

			_, err := getTokenClaims(w, r)

			fmt.Println("Checking the output of the check", err)

			if err != nil {

				if err.Error() == "No token found in Authorization header" {

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
					return
						return
					}

					next.ServeHTTP(w, r)
					return

				}

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
