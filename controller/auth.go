package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gocroot/config"
	"github.com/gocroot/helper/atdb"
	"github.com/gocroot/helper/auth"
	"github.com/gocroot/helper/watoken"
	"github.com/gocroot/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func Auth(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Ambil kredensial dari database
	creds, err := atdb.GetOneDoc[auth.GoogleCredential](config.Mongoconn, "credentials", bson.M{})
	if err != nil {
		http.Error(w, "Database Connection Problem: Unable to fetch credentials", http.StatusBadGateway)
		return
	}

	// Verifikasi ID token menggunakan client_id
	payload, err := auth.VerifyIDToken(request.Token, creds.ClientID)
	if err != nil {
		http.Error(w, "Invalid token: Token verification failed", http.StatusUnauthorized)
		return
	}

	userInfo := model.Userdomyikado{
		Name:                 payload.Claims["name"].(string),
		Email:                payload.Claims["email"].(string),
		GoogleProfilePicture: payload.Claims["picture"].(string),
	}

	// Simpan atau perbarui informasi pengguna di database
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := config.Mongoconn.Collection("user")
	filter := bson.M{"email": userInfo.Email}

	var existingUser model.Userdomyikado
	err = collection.FindOne(ctx, filter).Decode(&existingUser)
	if err != nil || existingUser.PhoneNumber == "" {
		// User does not exist or exists but has no phone number, request QR scan
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		response, _ := json.Marshal(map[string]interface{}{
			"message": "Please scan the QR code to provide your phone number",
			"user":    userInfo,
			"token" : "",
		})
		w.Write(response)
		return
	}else if existingUser.PhoneNumber != "" {
		token, err := watoken.EncodeforHours(existingUser.PhoneNumber, existingUser.Name, config.PrivateKey, 18) // Generating a token for 18 hours
		if err != nil {
			http.Error(w, "Token generation failed", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response, _ := json.Marshal(map[string]interface{}{
			"message": "Authenticated successfully",
			"user":    userInfo,
			"token":   token,
		})
		w.Write(response)
		return
	}

	update := bson.M{
		"$set": userInfo,
	}
	opts := options.Update().SetUpsert(true)
	_, err = collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		http.Error(w, "Failed to save user info: Database update failed", http.StatusInternalServerError)
		return
	}

	response, err := json.Marshal(userInfo)
	if err != nil {
		http.Error(w, "Internal server error: JSON marshaling failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
    var request model.Stp
    if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }

    // Salt and iterations - in a real application, store and retrieve these for each user
    salt := "randomSalt123"
    iterations := 100000

    hashedPassword := auth.HashPassword(request.PasswordHash, salt, iterations)

    // Find user in the database
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    collection := config.Mongoconn.Collection("user")
    filter := bson.M{"phonenumber": request.PhoneNumber}

    var user model.Userdomyikado
    err := collection.FindOne(ctx, filter).Decode(&user)
    if err != nil {
        http.Error(w, "Invalid phone number or password", http.StatusUnauthorized)
        return
    }

    // Verify password
    if user.PasswordHash != hashedPassword {
        http.Error(w, "Invalid phone number or password", http.StatusUnauthorized)
        return
    }

    // Generate token
    token, err := watoken.EncodeforHours(user.PhoneNumber, user.Name, config.PrivateKey, 18)
    if err != nil {
        http.Error(w, "Token generation failed", http.StatusInternalServerError)
        return
    }

    response, err := json.Marshal(map[string]interface{}{
        "message": "Authenticated successfully",
        "user":    user,
        "token":   token,
    })
    if err != nil {
        http.Error(w, "Internal server error: JSON marshaling failed", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.Write(response)
}

