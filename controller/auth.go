package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"regexp"
	"time"

	"github.com/gocroot/config"
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
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(map[string]string{"message": "Invalid request"})
        return
    }

    // Ambil kredensial dari database
    creds, err := atdb.GetOneDoc[auth.GoogleCredential](config.Mongoconn, "credentials", bson.M{})
    if err != nil {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusBadGateway)
        json.NewEncoder(w).Encode(map[string]string{"message": "Database Connection Problem: Unable to fetch credentials"})
        return
    }

    // Verifikasi ID token menggunakan client_id
    payload, err := auth.VerifyIDToken(request.Token, creds.ClientID)
    if err != nil {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusUnauthorized)
        json.NewEncoder(w).Encode(map[string]string{"message": "Invalid token: Token verification failed"})
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
        response := map[string]interface{}{
            "message": "Please scan the QR code to provide your phone number",
            "user":    userInfo,
            "token":   "",
        }
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusUnauthorized)
        json.NewEncoder(w).Encode(response)
        return
    } else if existingUser.PhoneNumber != "" {
        token, err := watoken.EncodeforHours(existingUser.PhoneNumber, existingUser.Name, config.PrivateKey, 18) // Generating a token for 18 hours
        if err != nil {
            w.Header().Set("Content-Type", "application/json")
            w.WriteHeader(http.StatusInternalServerError)
            json.NewEncoder(w).Encode(map[string]string{"message": "Token generation failed"})
            return
        }
        response := map[string]interface{}{
            "message": "Authenticated successfully",
            "user":    userInfo,
            "token":   token,
        }
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(response)
        return
    }

    update := bson.M{
        "$set": userInfo,
    }
    opts := options.Update().SetUpsert(true)
    _, err = collection.UpdateOne(ctx, filter, update, opts)
    if err != nil {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(map[string]string{"message": "Failed to save user info: Database update failed"})
        return
    }

    response := map[string]interface{}{
        "user": userInfo,
    }
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(response)
}


func GeneratePasswordHandler(w http.ResponseWriter, r *http.Request) {
	var request struct {
		PhoneNumber string `json:"phonenumber"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": "Invalid request"})
		return
	}

	// Validate phone number
	re := regexp.MustCompile(`^62\d{9,15}$`)
	if !re.MatchString(request.PhoneNumber) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": "Invalid phone number format"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if the phone number is registered
	var user struct {
		Name string `bson:"name"`
	}
	userCollection := config.Mongoconn.Collection("user")
	err := userCollection.FindOne(ctx, bson.M{"phonenumber": request.PhoneNumber}).Decode(&user)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"message": "Phone number not registered"})
		return
	}

	// Generate random password
	randomPassword, err := auth.GenerateRandomPassword(12)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Failed to generate password"})
		return
	}

	// Hash the password
	hashedPassword, err := auth.HashPassword(randomPassword)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Failed to hash password"})
		return
	}

	// Update or insert the user in the database
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
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Failed to save user info"})
		return
	}

	// Send password via WhatsApp
	dt := &whatsauth.TextMessage{
		To:      request.PhoneNumber,
		IsGroup: false,
		Messages: "Hi! Your login password is: " + randomPassword +
			". Enter this password on the STP page within 4 minutes. The password will expire after that.",
	}
	_, resp, err := atapi.PostStructWithToken[model.Response]("Token", config.WAAPIToken, dt, config.WAAPIMessage)
	if err != nil {
		resp.Info = "Unauthorized"
		resp.Response = err.Error()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(resp)
		return
	}

	// Respond with success and the generated password
	response := map[string]interface{}{
		"message":        "Password generated and saved successfully",
		"password":       randomPassword,
		"hashedPassword": hashedPassword,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}



func VerifyPasswordHandler(w http.ResponseWriter, r *http.Request) {
    var request struct {
        PhoneNumber string `json:"phonenumber"`
        Password    string `json:"password"`
    }
    if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(map[string]string{"message": "Invalid request"})
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
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusUnauthorized)
        json.NewEncoder(w).Encode(map[string]string{"message": "Invalid phone number or password"})
        return
    }

    // Verify password and expiry
    if time.Now().After(user.CreatedAt.Add(4 * time.Minute)) {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusUnauthorized)
        json.NewEncoder(w).Encode(map[string]string{"message": "Password expired"})
        return
    }

    err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(request.Password))
    if err != nil {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusUnauthorized)
        json.NewEncoder(w).Encode(map[string]string{"message": "Invalid phone number or password"})
        return
    }
	token, err := watoken.EncodeforHours(user.PhoneNumber, user.PhoneNumber, config.PrivateKey, 18)
	if err != nil {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(map[string]string{"message": "Token generation failed"})
        return
    }
	response := map[string]interface{}{
        "message": "Authenticated successfully",
        "token":   token,
    }
    // Respond with success
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(response)
}






