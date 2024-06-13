package gcalendar

import "time"

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
