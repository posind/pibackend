package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/xdg-go/pbkdf2"
	"google.golang.org/api/idtoken"
)

func VerifyIDToken(idToken string, audience string) (*idtoken.Payload, error) {
	payload, err := idtoken.Validate(context.Background(), idToken, audience)
	if err != nil {
		return nil, fmt.Errorf("id token validation failed: %v", err)
	}
	return payload, nil
}

func HashPassword(password, salt string, iterations int) string {
	passwordBytes := []byte(password)
	saltBytes := []byte(salt)
	hash := pbkdf2.Key(passwordBytes, saltBytes, iterations, 32, sha256.New)
	return hex.EncodeToString(hash)
}
