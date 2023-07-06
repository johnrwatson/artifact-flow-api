package main

import (
	"context"
	"fmt"
	"log"

	"google.golang.org/api/idtoken"
)

func main() {
	jwt := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6ImpvaG5AZmFtaWx5d2F0c29uLmNvLnVrIiwiZXhwIjoxNjg4NTg3ODUwLCJyZWZyZXNoX3Rva2VuIjoiIiwic3ViIjoiMTA1NDMwNTkxMTM4OTM2OTU0ODE0In0.JObJnyHZH1KSnHtHRFMQgwC3S-yzBgCB3qAGsozVPig"
	audience := "160220461475-r2h1nktt3nagkub9jdl6tds2ntoh8gsn.apps.googleusercontent.com"

	// Create a context
	ctx := context.Background()

	// Verify and parse the JWT
	payload, err := idtoken.Validate(ctx, jwt, audience)
	if err != nil {
		log.Fatalf("Failed to validate JWT: %v", err)
	}

	// Extract the claims from the payload
	claims := payload.Claims
	issuer := claims["iss"].(string)
	subject := claims["sub"].(string)
	email := claims["email"].(string)

	fmt.Println("Issuer:", issuer)
	fmt.Println("Subject:", subject)
	fmt.Println("Email:", email)
}