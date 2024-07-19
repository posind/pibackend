package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"regexp"
	"time"

	"github.com/gocroot/config"
	"github.com/gocroot/helper/at"
	"github.com/gocroot/helper/atapi"
	"github.com/gocroot/helper/atdb"
	"github.com/gocroot/helper/auth"
	"github.com/gocroot/helper/watoken"
	"github.com/gocroot/helper/whatsauth"
	"github.com/gocroot/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

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

	respondWithJSON(w, http.StatusOK, response)
}

func GeneratePasswordHandler(w http.ResponseWriter, r *http.Request) {
	var request struct {
		PhoneNumber string `json:"phonenumber"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Validate phone number
	re := regexp.MustCompile(`^\+62\d{9,15}$`)
	if !re.MatchString(request.PhoneNumber) {
		http.Error(w, "Invalid phone number format", http.StatusBadRequest)
		return
	}

	// Generate random password
	randomPassword, err := auth.GenerateRandomPassword(12)
	if err != nil {
		http.Error(w, "Failed to generate password", http.StatusInternalServerError)
		return
	}

	// Hash the password
	hashedPassword, err := auth.HashPassword(randomPassword)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	// Update or insert the user in the database
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := config.Mongoconn.Collection("stp")
	filter := bson.M{"phonenumber": request.PhoneNumber}

	update := bson.M{
		"$set": model.Stp{
			PhoneNumber:  request.PhoneNumber,
			PasswordHash: hashedPassword,
			CreatedAt:    time.Now(),
		},
	}
	opts := options.Update().SetUpsert(true)
	_, err = collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		http.Error(w, "Failed to save user info", http.StatusInternalServerError)
		return
	}

	// Send the random password via WhatsApp
	dt := &whatsauth.TextMessage{
		To:       request.PhoneNumber + "@s.whatsapp.net", // Format the phone number for WhatsApp
		IsGroup:  false,
		Messages: "Your login password is: " + randomPassword + ". This password will expire in 4 minutes.",
	}
	_, resp, err := atapi.PostStructWithToken[model.Response]("Token", config.WAAPIToken, dt, config.WAAPIMessage)
	if err != nil {
		resp.Info = "Tidak berhak"
		resp.Response = err.Error()
		at.WriteJSON(w, http.StatusUnauthorized, resp)
		return
	}

	// Respond with success and the generated password
	response := map[string]interface{}{
		"message":       "Password generated and saved successfully",
		"password":      randomPassword,
		"hashedPassword": hashedPassword,
	}

	respondWithJSON(w, http.StatusOK, response)
}

func VerifyPasswordHandler(w http.ResponseWriter, r *http.Request) {
	var request model.VerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
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
		http.Error(w, "Invalid phone number or password", http.StatusUnauthorized)
		return
	}

	// Check if the password is expired
	if time.Since(user.CreatedAt) > 4*time.Minute {
		http.Error(w, "Password has expired", http.StatusUnauthorized)
		return
	}

	// Verify the password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(request.Password))
	if err != nil {
		http.Error(w, "Invalid phone number or password", http.StatusUnauthorized)
		return
	}

	// Respond with success
	response := map[string]interface{}{
		"message": "Password verified successfully",
	}
	respondWithJSON(w, http.StatusOK, response)
}





