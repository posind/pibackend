package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gocroot/helper/atdb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

// Database and collection names
const credCol = "credentials"
const tokenCol = "tokens"

var mongoinfo = atdb.DBInfo{
	DBString: os.Getenv("MONGOSTRINGTEST"),
	DBName:   "domyid",
}

var Mongoconn, ErrorMongoconn = atdb.MongoConnect(mongoinfo)

// Struct to hold the credentials data from MongoDB
type Credentials struct {
	ClientID     string   `json:"client_id"`
	ClientSecret string   `json:"client_secret"`
	RefreshToken string   `json:"refresh_token"`
	TokenURI     string   `json:"token_uri"`
	Scopes       []string `json:"scopes"`
	Expiry       string   `json:"expiry"`
	Token        string   `json:"token"`
}

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

// Retrieves a token source from MongoDB.
func tokenSourceFromMongo(db *mongo.Database, config *oauth2.Config) (oauth2.TokenSource, error) {
	token, err := tokenFromMongo(db)
	if err != nil {
		return nil, err
	}
	return config.TokenSource(context.Background(), token), nil
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
	var credData Credentials
	err := credCollection.FindOne(context.TODO(), bson.M{}).Decode(&credData)
	if err != nil {
		log.Fatalf("Unable to retrieve credentials from MongoDB: %v", err)
	}

	// Debug: Print the credentials retrieved from MongoDB
	fmt.Printf("Credentials retrieved from MongoDB: %+v\n", credData)

	// Create the OAuth2 config from the credentials data
	config := &oauth2.Config{
		ClientID:     credData.ClientID,
		ClientSecret: credData.ClientSecret,
		Endpoint:     google.Endpoint,
		Scopes:       credData.Scopes,
		RedirectURL:  "http://localhost",
	}

	// Initialize token with retrieved values
	expiryTime, err := time.Parse(time.RFC3339, credData.Expiry)
	if err != nil {
		log.Fatalf("Unable to parse token expiry time: %v", err)
	}
	token := &oauth2.Token{
		AccessToken:  credData.Token,
		TokenType:    "Bearer",
		RefreshToken: credData.RefreshToken,
		Expiry:       expiryTime,
	}

	// Save the token to MongoDB
	saveTokenToMongo(Mongoconn, token)

	// Create token source from credentials and config
	tokenSource, err := tokenSourceFromMongo(Mongoconn, config)
	if err != nil {
		log.Fatalf("Unable to create token source: %v", err)
	}
	// Create OAuth2 client using token source
	httpClient := oauth2.NewClient(context.Background(), tokenSource)

	srv, err := calendar.NewService(context.TODO(), option.WithHTTPClient(httpClient))
	if err != nil {
		// If token expired, try refreshing token and creating service again
		if strings.Contains(err.Error(), "token expired") {
			tokenSource := oauth2.ReuseTokenSource(nil, tokenSource)
			httpClient := oauth2.NewClient(context.Background(), tokenSource)
			srv, err = calendar.NewService(context.Background(), option.WithHTTPClient(httpClient))
			if err != nil {
				log.Fatalf("Unable to retrieve Calendar client: %v", err)
			}
		} else {
			log.Fatalf("Unable to retrieve Calendar client: %v", err)
		}
	}

	// Create event...

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
