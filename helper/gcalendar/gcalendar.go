package gcalendar

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/gocroot/helper/atdb"
	"golang.org/x/oauth2"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Retrieve a token from MongoDB
func tokenFromDB(db *mongo.Database) (*oauth2.Token, error) {
	tokenRecord, err := atdb.GetOneDoc[CredentialRecord](db, "tokens", bson.M{})
	if err != nil {
		return nil, err
	}

	token := &oauth2.Token{
		AccessToken:  tokenRecord.Token,
		RefreshToken: tokenRecord.RefreshToken,
		TokenType:    "Bearer",
		Expiry:       tokenRecord.Expiry,
	}
	if tokenRecord.Token == "" {
		return nil, errors.New("token tidak ada")
	}

	return token, nil
}

// Saves a token to MongoDB
func saveToken(db *mongo.Database, token *oauth2.Token) {
	collection := db.Collection("tokens")
	tokenRecord := bson.M{
		"token":         token.AccessToken,
		"refresh_token": token.RefreshToken,
		"expiry":        token.Expiry,
	}

	_, err := collection.UpdateOne(
		context.TODO(),
		bson.M{},
		bson.M{"$set": tokenRecord},
		options.Update().SetUpsert(true),
	)
	if err != nil {
		log.Fatalf("Unable to save oauth token: %v", err)
	}
}

// Retrieve credentials.json from MongoDB
func credentialsFromDB() (*oauth2.Config, error) {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb+srv://ulbi:k0dGfeYgAorMKDAz@cluster0.fvazjna.mongodb.net/"))
	if err != nil {
		return nil, err
	}
	defer client.Disconnect(context.TODO())

	collection := client.Database("domyid").Collection("credentials")
	var credentialRecord CredentialRecord
	err = collection.FindOne(context.TODO(), bson.M{}).Decode(&credentialRecord)
	if err != nil {
		return nil, err
	}

	if len(credentialRecord.RedirectURIs) == 0 {
		return nil, fmt.Errorf("no redirect URIs found in credentials")
	}

	config := &oauth2.Config{
		ClientID:     credentialRecord.ClientID,
		ClientSecret: credentialRecord.ClientSecret,
		Scopes:       credentialRecord.Scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  credentialRecord.AuthURI,
			TokenURL: credentialRecord.TokenURI,
		},
		RedirectURL: credentialRecord.RedirectURIs[0], // Using the first redirect URI
	}

	return config, nil
}

// Retrieve a token, saves the token, then returns the generated client
func getClient(db *mongo.Database, config *oauth2.Config) *http.Client {
	tok, err := tokenFromDB(db)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(db, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token
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

func Run(db *mongo.Database) {
	ctx := context.Background()

	config, err := credentialsFromDB()
	if err != nil {
		log.Fatalf("Unable to retrieve client secret from DB: %v", err)
	}

	client := getClient(db, config)
	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Calendar client: %v", err)
	}

	event := &calendar.Event{
		Summary:     "Google I/O 2024",
		Location:    "800 Howard St., San Francisco, CA 94103",
		Description: "A chance to hear more about Google's developer products.",
		Start: &calendar.EventDateTime{
			DateTime: "2024-06-14T09:00:00+07:00",
			TimeZone: "Asia/Jakarta",
		},
		End: &calendar.EventDateTime{
			DateTime: "2024-06-14T17:00:00+07:00",
			TimeZone: "Asia/Jakarta",
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
