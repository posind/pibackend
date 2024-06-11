package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"google.golang.org/api/idtoken"
)

func VerifyIDToken(idToken string, audience string) (*idtoken.Payload, error) {
	payload, err := idtoken.Validate(context.Background(), idToken, audience)
	if err != nil {
		return nil, fmt.Errorf("id token validation failed: %v", err)
	}
	return payload, nil
}

func GetIDTokenFromHeader(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("missing Authorization header")
	}

	splitToken := strings.Split(authHeader, "Bearer ")
	if len(splitToken) != 2 {
		return "", fmt.Errorf("invalid token format")
	}

	return splitToken[1], nil
}
