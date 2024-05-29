package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gocroot/helper"
	"github.com/gocroot/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

// MongoDB URI
const uri = "mongodb://localhost:27017"

// Database and collection names
const dbName = "google_api"
const credCol = "credentials"
const tokenCol = "tokens"

var mongoinfo = model.DBInfo{
	DBString: "mongodb+srv://domyid:d2FqCsbjSS7hW2Xt@cluster0.fvazjna.mongodb.net/",
	DBName:   "domyid",
}

var Mongoconn, ErrorMongoconn = helper.MongoConnect(mongoinfo)

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config, db *mongo.Database) *http.Client {
	tok, err := tokenFromMongo(db)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveTokenToMongo(db, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// Retrieves a token from MongoDB.
func tokenFromMongo(db *mongo.Database) (*oauth2.Token, error) {
	collection := db.Collection(tokenCol)
	var token oauth2.Token
	err := collection.FindOne(context.TODO(), bson.M{}).Decode(&token)
	if err != nil {
		return nil, err
	}
	return &token, nil
}

// Saves a token to MongoDB.
func saveTokenToMongo(db *mongo.Database, token *oauth2.Token) {
	collection := db.Collection(tokenCol)
	// Remove any existing tokens
	_, err := collection.DeleteMany(context.TODO(), bson.M{})
	if err != nil {
		log.Fatalf("Unable to delete old tokens from MongoDB: %v", err)
	}
	_, err = collection.InsertOne(context.TODO(), token)
	if err != nil {
		log.Fatalf("Unable to save token to MongoDB: %v", err)
	}
}

func main() {
	if ErrorMongoconn != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", ErrorMongoconn)
	}

	// Read credentials.json from MongoDB
	credCollection := Mongoconn.Collection(credCol)
	var credData bson.M
	err := credCollection.FindOne(context.TODO(), bson.M{}).Decode(&credData)
	if err != nil {
		log.Fatalf("Unable to retrieve credentials from MongoDB: %v", err)
	}

	// Debug: Print the credentials retrieved from MongoDB
	fmt.Printf("Credentials retrieved from MongoDB: %v\n", credData)

	// Remove the MongoDB specific _id field
	delete(credData, "_id")

	// Debug: Print the credentials after removing _id
	fmt.Printf("Credentials after removing _id: %v\n", credData)

	credBytes, err := json.Marshal(credData)
	if err != nil {
		log.Fatalf("Unable to marshal credentials: %v", err)
	}

	// Debug: Print the marshaled JSON credentials
	fmt.Printf("Marshaled JSON credentials: %s\n", string(credBytes))

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(credBytes, calendar.CalendarScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}

	httpClient := getClient(config, Mongoconn)

	srv, err := calendar.NewService(context.TODO(), option.WithHTTPClient(httpClient))
	if err != nil {
		log.Fatalf("Unable to retrieve Calendar client: %v", err)
	}

	event := &calendar.Event{
		Summary:     "Google I/O 2024",
		Location:    "800 Howard St., San Francisco, CA 94103",
		Description: "A chance to hear more about Google's developer products.",
		Start: &calendar.EventDateTime{
			DateTime: "2024-06-28T09:00:00-07:00",
			TimeZone: "America/Los_Angeles",
		},
		End: &calendar.EventDateTime{
			DateTime: "2024-06-28T17:00:00-07:00",
			TimeZone: "America/Los_Angeles",
		},
		Attendees: []*calendar.EventAttendee{
			{Email: "awangga@gmail.com"},
			{Email: "awangga@ulbi.ac.id"},
		},
	}

	calendarId := "primary"
	event, err = srv.Events.Insert(calendarId, event).Do()
	if err != nil {
		log.Fatalf("Unable to create event: %v", err)
	}
	fmt.Printf("Event created: %s\n", event.HtmlLink)
}
