package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gocroot/config"
	"github.com/gocroot/helper/atapi"
	"github.com/gocroot/helper/whatsauth"
	"github.com/gocroot/model"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/api/idtoken"
)

func VerifyIDToken(idToken string, audience string) (*idtoken.Payload, error) {
	payload, err := idtoken.Validate(context.Background(), idToken, audience)
	if err != nil {
		return nil, fmt.Errorf("id token validation failed: %v", err)
	}
	return payload, nil
}

func GenerateRandomPassword(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func SendWhatsAppPassword(w http.ResponseWriter, phoneNumber string, password string) {
    // Prepare WhatsApp message
    dt := &whatsauth.TextMessage{
        To:      phoneNumber,
        IsGroup: false,
        Messages: "Hi! Your login password is: *" + password + "*.\n\n" +
        "Enter this password on the STP page within 4 minutes. The password will expire after that. " +
        "To copy the password, press and hold the password.",
    }

    // Send WhatsApp message
    _, resp, err := atapi.PostStructWithToken[model.Response]("Token", config.WAAPIToken, dt, config.WAAPIMessage)
    if err != nil {
        resp.Info = "Unauthorized"
        resp.Response = err.Error()
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusUnauthorized)
        json.NewEncoder(w).Encode(resp)
        return
    }

    if resp.Info == "Unauthorized" {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusUnauthorized)
        json.NewEncoder(w).Encode(map[string]string{"message": "Unauthorized access"})
        return
    }
}




