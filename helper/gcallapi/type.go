package gcallapi

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CredentialRecord struct {
	Token               string    `bson:"token"`
	RefreshToken        string    `bson:"refresh_token"`
	TokenURI            string    `bson:"token_uri"`
	ClientID            string    `bson:"client_id"`
	ClientSecret        string    `bson:"client_secret"`
	Expiry              time.Time `bson:"expiry"`
	AuthProviderCertURL string    `bson:"auth_provider_x509_cert_url"`
	AuthURI             string    `bson:"auth_uri"`
	ProjectID           string    `bson:"project_id"`
	RedirectURIs        []string  `bson:"redirect_uris"`
	JavascriptOrigins   []string  `bson:"javascript_origins"`
	Scopes              []string  `bson:"scopes"`
}

type SimpleEvent struct {
	ProjectID   primitive.ObjectID `json:"project_id"`
	Summary     string             `json:"summary"`
	Location    string             `json:"location"`
	Description string             `json:"description"`
	Date        string             `json:"date"`      // YYYY-MM-DD
	TimeStart   string             `json:"timestart"` // HH:MM:SS
	TimeEnd     string             `json:"timeend"`   // HH:MM:SS
	Attendees   []string
}
