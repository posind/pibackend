package report

import "go.mongodb.org/mongo-driver/bson/primitive"

type NewLiburNasional struct {
	Tanggal    string `json:"tanggal"`
	Keterangan string `json:"keterangan"`
	IsCuti     bool   `json:"is_cuti"`
}

type RekapUser struct {
	Nama        string
	PhoneNumber string
	NamaProject string
}

type LogPoin struct {
	ID          primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	UserID      primitive.ObjectID `json:"userid,omitempty" bson:"userid,omitempty"`
	Name        string             `json:"name,omitempty" bson:"name,omitempty"`
	PhoneNumber string             `json:"phonenumber,omitempty" bson:"phonenumber,omitempty"`
	Email       string             `json:"email,omitempty" bson:"email,omitempty"`
	ProjectID   primitive.ObjectID `json:"projectid,omitempty" bson:"projectid,omitempty"`
	ProjectName string             `json:"projectname,omitempty" bson:"projectname,omitempty"`
	Poin        float64            `json:"poin,omitempty" bson:"poin,omitempty"`
	Activity    string             `json:"activity,omitempty" bson:"activity,omitempty"`
}
