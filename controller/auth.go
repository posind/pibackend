package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gocroot/config"
	"github.com/gocroot/helper/atdb"
	"github.com/gocroot/helper/auth"
	"github.com/gocroot/helper/watoken"
	"github.com/gocroot/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
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
	var request model.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		respondWithJSON(w, http.StatusBadRequest, map[string]string{"message": "Invalid request"})
		return
	}

	// Find user in the database
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := config.Mongoconn.Collection("stp")
	filter := bson.M{"phonenumber": request.PhoneNumber}

	var user model.Stp
	err := collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		fmt.Printf("Error finding user: %v\n", err)
		respondWithJSON(w, http.StatusUnauthorized, map[string]string{"message": "Invalid phone number or password"})
		return
	}

	// Verify password using bcrypt
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(request.Password))
	if err != nil {
		var respn model.Response
		respn.Status = "Error: Passwords are not the same"
		respn.Info = "Additional info"
		respn.Location = "Password Verification Error"
		respn.Response = fmt.Sprintf("Password verification failed for phone number: %s", request.PhoneNumber)
		respondWithJSON(w, http.StatusUnauthorized, respn)
		return
	}

	// Generate token
	token, err := watoken.EncodeforHours(user.PhoneNumber, user.PasswordHash, config.PrivateKey, 18)
	if err != nil {
		respondWithJSON(w, http.StatusInternalServerError, map[string]string{"message": "Token generation failed"})
		return
	}

	response := map[string]interface{}{
		"message": "Authenticated successfully",
		"user":    user,
		"token":   token,
	}
	respondWithJSON(w, http.StatusOK, response)
}



func respondWithJSON(w http.ResponseWriter, status int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("HTTP 500: Internal Server Error"))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(response)
}


