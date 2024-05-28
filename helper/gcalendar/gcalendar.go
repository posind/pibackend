package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
)

// MongoDB URI
const uri = "mongodb://localhost:27017"

// Database and collection names
const dbName = "google_api"
const credCol = "credentials"
const tokenCol = "tokens"

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config, client *mongo.Client) *http.Client {
	tok, err := tokenFromMongo(client)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveTokenToMongo(client, tok)
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
func tokenFromMongo(client *mongo.Client) (*oauth2.Token, error) {
	collection := client.Database(dbName).Collection(tokenCol)
	var token oauth2.Token
	err := collection.FindOne(context.TODO(), bson.M{}).Decode(&token)
	if err != nil {
		return nil, err
	}
	return &token, nil
}

// Saves a token to MongoDB.
func saveTokenToMongo(client *mongo.Client, token *oauth2.Token) {
	collection := client.Database(dbName).Collection(tokenCol)
	_, err := collection.InsertOne(context.TODO(), token)
	if err != nil {
		log.Fatalf("Unable to save token to MongoDB: %v", err)
	}
}

func main() {
	// MongoDB client setup
	clientOpts := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(context.TODO(), clientOpts)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer client.Disconnect(context.TODO())

	// Read credentials.json from MongoDB
	credCollection := client.Database(dbName).Collection(credCol)
	var credData bson.M
	err = credCollection.FindOne(context.TODO(), bson.M{}).Decode(&credData)
	if err != nil {
		log.Fatalf("Unable to retrieve credentials from MongoDB: %v", err)
	}

	credBytes, err := json.Marshal(credData)
	if err != nil {
		log.Fatalf("Unable to marshal credentials: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(credBytes, calendar.CalendarScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}

	httpClient := getClient(config, client)

	srv, err := calendar.New(httpClient)
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
			{Email: "lpage@example.com"},
			{Email: "sbrin@example.com"},
		},
	}

	calendarId := "primary"
	event, err = srv.Events.Insert(calendarId, event).Do()
	if err != nil {
		log.Fatalf("Unable to create event: %v", err)
	}
	fmt.Printf("Event created: %s\n", event.HtmlLink)
}
