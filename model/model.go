package model

type Response struct {
	Response string `json:"response"`
	Info     string `json:"info,omitempty"`
	Status   string `json:"status,omitempty"`
	Location string `json:"location,omitempty"`
}

type Profile struct {
	Token       string `bson:"token"`
	Phonenumber string `bson:"phonenumber"`
	Secret      string `bson:"secret"`
	URL         string `bson:"url"`
	QRKeyword   string `bson:"qrkeyword"`
	PublicKey   string `bson:"publickey"`
}
