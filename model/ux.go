package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type Laporan struct {
	ID        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty" query:"id" url:"_id,omitempty" reqHeader:"_id"`
	Project   Project            `json:"project,omitempty" bson:"project,omitempty"`
	User      Userdomyikado      `json:"user,omitempty" bson:"user,omitempty"`
	Petugas   string             `json:"petugas,omitempty" bson:"petugas,omitempty"`
	NoPetugas string             `json:"nopetugas,omitempty" bson:"nopetugas,omitempty"`
	Kode      string             `json:"kode,omitempty" bson:"kode,omitempty"`
	Nama      string             `json:"nama,omitempty" bson:"nama,omitempty"`
	Phone     string             `json:"phone,omitempty" bson:"phone,omitempty"`
	Solusi    string             `json:"solusi,omitempty" bson:"solusi,omitempty"`
	Komentar  string             `json:"komentar,omitempty" bson:"komentar,omitempty"`
	Rating    float64            `json:"rating,omitempty" bson:"rating,omitempty"`
}

type Rating struct {
	ID       string `json:"id,omitempty" bson:"id,omitempty" query:"id" url:"id,omitempty" reqHeader:"id"`
	Komentar string `json:"komentar,omitempty" bson:"komentar,omitempty"`
	Rating   int    `json:"rating,omitempty" bson:"rating,omitempty"`
}
