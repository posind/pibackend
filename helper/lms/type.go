package lms

import (
	"encoding/json"
	"time"
)

type UnixTime struct {
	time.Time
}

func (ut *UnixTime) UnmarshalJSON(b []byte) error {
	var ts int64
	if err := json.Unmarshal(b, &ts); err != nil {
		return err
	}
	ut.Time = time.Unix(ts, 0).UTC()
	return nil
}

type LoginProfile struct {
	Username string `bson:"user,omitempty"`
	Bearer   string `bson:"bearer,omitempty"`
	Xsrf     string `bson:"xsrf,omitempty"`
	Lsession string `bson:"lsession,omitempty"`
}

type Position struct {
	ID        string    `json:"id,omitempty"`
	Name      string    `json:"name,omitempty"`
	ParentID  string    `json:"parent_id,omitempty"`
	Order     *int      `json:"order,omitempty"`
	IsDelete  bool      `json:"is_delete,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

type Province struct {
	Kode      string    `json:"kode,omitempty"`
	Nama      string    `json:"nama,omitempty"`
	IsDelete  bool      `json:"is_delete,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
	IDs       string    `json:"ids,omitempty"`
}

type Regency struct {
	Kode      string    `json:"kode,omitempty"`
	Nama      string    `json:"nama,omitempty"`
	IsDelete  bool      `json:"is_delete,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
	IDs       *string   `json:"ids,omitempty"`
}

type District struct {
	Kode      string    `json:"kode,omitempty"`
	Nama      string    `json:"nama,omitempty"`
	IsDelete  bool      `json:"is_delete,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
	IDs       *string   `json:"ids,omitempty"`
}

type Village struct {
	Kode      string    `json:"kode,omitempty"`
	Nama      string    `json:"nama,omitempty"`
	IsDelete  bool      `json:"is_delete,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
	IDs       *string   `json:"ids,omitempty"`
}

type UserProfile struct {
	TMT         string   `json:"tmt,omitempty"`
	Position    Position `json:"position,omitempty"`
	Province    Province `json:"province,omitempty"`
	Regency     Regency  `json:"regency,omitempty"`
	District    District `json:"district,omitempty"`
	Village     Village  `json:"village,omitempty"`
	Decree      string   `json:"decree,omitempty"`
	TrainerCert *string  `json:"trainer_cert,omitempty"`
}

type User struct {
	ID              string       `json:"id,omitempty"`
	Fullname        string       `json:"fullname,omitempty"`
	Username        string       `json:"username,omitempty"`
	Email           string       `json:"email,omitempty"`
	EmailVerified   UnixTime     `json:"email_verified,omitempty"`
	ProfileVerified bool         `json:"profile_verified,omitempty"`
	ProfileApproved int          `json:"profile_approved,omitempty"`
	LastLoginAt     UnixTime     `json:"last_login_at,omitempty"`
	UserProfile     *UserProfile `json:"user_profile,omitempty"`
	CreatedAt       time.Time    `json:"created_at,omitempty"`
	Roles           []string     `json:"roles,omitempty"`
	ApprovedBy      *string      `json:"approved_by,omitempty"`
	ApprovedAt      *time.Time   `json:"approved_at,omitempty"`
	RejectedBy      *string      `json:"rejected_by,omitempty"`
	RejectedAt      *time.Time   `json:"rejected_at,omitempty"`
}

type Meta struct {
	CurrentPage int `json:"current_page,omitempty"`
	FirstItem   int `json:"first_item,omitempty"`
	LastItem    int `json:"last_item,omitempty"`
	LastPage    int `json:"last_page,omitempty"`
	Total       int `json:"total,omitempty"`
}

type Data struct {
	Data []User `json:"data,omitempty"`
	Meta Meta   `json:"meta,omitempty"`
}

type Root struct {
	Data Data `json:"data,omitempty"`
}
