package model

type QRStatus struct {
	PhoneNumber string `json:"phonenumber"`
	Status      bool   `json:"status"`
	Code        string `json:"code"`
	Message     string `json:"message"`
}

type SendText struct {
	To       string `json:"to"`
	IsGroup  bool   `json:"isgroup"`
	Messages string `json:"messages"`
}
