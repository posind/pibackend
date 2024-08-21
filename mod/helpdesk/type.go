package helpdesk

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

/* type User struct {
	ID           primitive.ObjectID  `json:"id,omitempty" bson:"_id,omitempty"`
	Team         string              `json:"team,omitempty" bson:"team,omitempty"`
	Scope        string              `json:"scope,omitempty" bson:"scope,omitempty"`
	Name         string              `json:"name,omitempty" bson:"name,omitempty"`
	Phonenumbers string              `json:"phonenumbers,omitempty" bson:"phonenumbers,omitempty"`
	Terlayani    bool                `json:"terlayani,omitempty" bson:"terlayani,omitempty"`
	Masalah      string              `json:"masalah,omitempty" bson:"masalah,omitempty"`
	Solusi       string              `json:"solusi,omitempty" bson:"solusi,omitempty"`
	RateLayanan  int                 `json:"ratelayanan,omitempty" bson:"ratelayanan,omitempty"`
	Operator     model.Userdomyikado `json:"operator,omitempty" bson:"operator,omitempty"`
} */

type ContactAdmin struct {
	Fullname string `json:"fullname"`
	Phone    string `json:"phone"`
}

type Data struct {
	Fullname             string         `json:"fullname"`
	Province             string         `json:"province"`
	Regency              string         `json:"regency"`
	District             string         `json:"district"`
	Village              string         `json:"village"`
	ContactAdminRegency  []ContactAdmin `json:"contact_admin_regency"`
	ContactAdminProvince []ContactAdmin `json:"contact_admin_province"`
}

type Response struct {
	Success bool `json:"success"`
	Data    Data `json:"data"`
}

type Session struct {
	ID          string     `bson:"_id,omitempty"`
	PhoneNumber string     `bson:"phonenumber"`
	Menulist    []MenuList `bson:"list"`
	CreatedAt   time.Time  `bson:"createdAt"`
}

type Menu struct {
	ID      primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"` // Field ID untuk MongoDB
	Keyword string             `bson:"keyword,omitempty" json:"keyword,omitempty"`
	Header  string             `bson:"header,omitempty" json:"header,omitempty"`
	List    []MenuList         `bson:"list,omitempty" json:"list,omitempty"`
	Footer  string             `bson:"footer,omitempty" json:"footer,omitempty"`
}

type MenuList struct {
	No      int    `bson:"no,omitempty" json:"no,omitempty"`
	Keyword string `bson:"keyword,omitempty" json:"keyword,omitempty"`
	Konten  string `bson:"konten,omitempty" json:"konten,omitempty"`
}
