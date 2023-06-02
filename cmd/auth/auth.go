package auth

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"net/http"
	"os"
	"strings"
	"log"
)


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

func SetupOauthProvider() bool {

	// Set up Google OAuth2 configuration
	oauthClientId := os.Getenv("OAUTH_CLIENT_ID")
	oauthClientSecret := os.Getenv("OAUTH_CLIENT_SECRET")
	oauthRedirectURL := os.Getenv("OAUTH_REDIRECT_URL")

	// Check if any of the variables are empty or not set
	if oauthClientId == "" || oauthClientSecret == "" || oauthRedirectURL == "" || os.Getenv("OAUTH_JWT_KEY") == "" {
		fmt.Println("Error: One or more OAuth variables are not set, these are listed below: \n - OAUTH_CLIENT_ID\n - OAUTH_CLIENT_SECRET\n - OAUTH_REDIRECT_URL\n - OAUTH_JWT_KEY")
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

func ValidateToken(tokenString string) (*Claims, error) {
	// Remove 'Bearer ' prefix from token string
	if len(tokenString) > 7 && strings.ToUpper(tokenString[0:7]) == "BEARER " {
		tokenString = tokenString[7:]
	}

	fmt.Println(tokenString)
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		return nil, errors.New("Failed to parse token")
	}

	fmt.Println(token)
	fmt.Println(token.Claims.(*Claims))
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("Invalid token")
	}
	fmt.Println(claims)
	return claims, nil
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
