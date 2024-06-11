package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gocroot/config"
	"github.com/gocroot/helper/atdb"
	"github.com/gocroot/helper/auth"
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
		http.Error(w, "Database Connection Problem", http.StatusBadGateway)
		return
	}

	// Verifikasi ID token menggunakan client_id
	payload, err := auth.VerifyIDToken(request.Token, creds.ClientID)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	userInfo := model.Userdomyikado{
		Name:    payload.Claims["name"].(string),
		Email:   payload.Claims["email"].(string),
		GoogleProfilePicture: payload.Claims["googleprofilepicture"].(string),
	}

	// Simpan atau perbarui informasi pengguna di database
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := config.Mongoconn.Collection("user")
	filter := bson.M{"email": userInfo.Email}
	update := bson.M{
		"$set": userInfo,
	}
	opts := options.Update().SetUpsert(true)

	_, err = collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		http.Error(w, "Failed to save user info", http.StatusInternalServerError)
		return
	}

	response, err := json.Marshal(userInfo)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}