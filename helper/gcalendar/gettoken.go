package gcalendar

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/oauth2"
)

// Request a token from the web, then returns the retrieved token
func GetTokenFromWeb(config *oauth2.Config) (*oauth2.Token, error) {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		return nil, err
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		return nil, err
	}
	return tok, nil
}

// Saves a token to MongoDB
func SaveToken(db *mongo.Database, token *oauth2.Token) (err error) {
	collection := db.Collection("tokens")
	tokenRecord := bson.M{
		"token":         token.AccessToken,
		"refresh_token": token.RefreshToken,
		"expiry":        token.Expiry,
	}

	_, err = collection.UpdateOne(
		context.TODO(),
		bson.M{},
		bson.M{"$set": tokenRecord},
		options.Update().SetUpsert(true),
	)
	if err != nil {
		return
	}
	return
}
