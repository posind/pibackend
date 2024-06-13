package gcalendar

import (
	"context"
	"errors"
	"net/http"

	"github.com/gocroot/helper/atdb"
	"golang.org/x/oauth2"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
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

// Retrieve credentials.json from MongoDB
func credentialsFromDB(db *mongo.Database) (*oauth2.Config, error) {
	credentialRecord, err := atdb.GetOneDoc[CredentialRecord](db, "credentials", bson.M{})
	if err != nil {
		return nil, err
	}

	if len(credentialRecord.RedirectURIs) == 0 {
		return nil, errors.New("no redirect URIs found in credentials")
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
func getClient(db *mongo.Database, config *oauth2.Config) (*http.Client, error) {
	tok, err := tokenFromDB(db)
	if err != nil {
		return nil, err
		// jika token habis buka ini dan jalankan di lokal
		// tok, err = GetTokenFromWeb(config)
		// if err != nil {
		// 	return nil, err
		// }
		// err = SaveToken(db, tok)
		// if err != nil {
		// 	return nil, err
		// }
	}
	return config.Client(context.Background(), tok), nil
}

func HandlerCalendar(db *mongo.Database) (event *calendar.Event, err error) {
	ctx := context.Background()

	config, err := credentialsFromDB(db)
	if err != nil {
		return
	}

	client, err := getClient(db, config)
	if err != nil {
		return
	}
	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return
	}

	event = &calendar.Event{
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
		return
	}
	return
}
