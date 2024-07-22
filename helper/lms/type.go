package lms

import "time"

type LoginProfile struct {
	Username string `bson:"user,omitempty"`
	Bearer   string `bson:"bearer,omitempty"`
	Xsrf     string `bson:"xsrf,omitempty"`
	Lsession string `bson:"lsession,omitempty"`
}

type Position struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	ParentID  string    `json:"parent_id"`
	Order     *int      `json:"order"`
	IsDelete  bool      `json:"is_delete"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Province struct {
	Kode      string    `json:"kode"`
	Nama      string    `json:"nama"`
	IsDelete  bool      `json:"is_delete"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	IDs       string    `json:"ids"`
}

type Regency struct {
	Kode      string    `json:"kode"`
	Nama      string    `json:"nama"`
	IsDelete  bool      `json:"is_delete"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	IDs       *string   `json:"ids"`
}

type District struct {
	Kode      string    `json:"kode"`
	Nama      string    `json:"nama"`
	IsDelete  bool      `json:"is_delete"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	IDs       *string   `json:"ids"`
}

type Village struct {
	Kode      string    `json:"kode"`
	Nama      string    `json:"nama"`
	IsDelete  bool      `json:"is_delete"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	IDs       *string   `json:"ids"`
}

type UserProfile struct {
	TMT         string   `json:"tmt"`
	Position    Position `json:"position"`
	Province    Province `json:"province"`
	Regency     Regency  `json:"regency"`
	District    District `json:"district"`
	Village     Village  `json:"village"`
	Decree      string   `json:"decree"`
	TrainerCert *string  `json:"trainer_cert"`
}

type User struct {
	ID              string       `json:"id"`
	Fullname        string       `json:"fullname"`
	Username        string       `json:"username"`
	Email           string       `json:"email"`
	EmailVerified   int64        `json:"email_verified"`
	ProfileVerified bool         `json:"profile_verified"`
	ProfileApproved int          `json:"profile_approved"`
	LastLoginAt     int64        `json:"last_login_at"`
	UserProfile     *UserProfile `json:"user_profile"`
	CreatedAt       time.Time    `json:"created_at"`
	Roles           []string     `json:"roles"`
	ApprovedBy      *string      `json:"approved_by"`
	ApprovedAt      *time.Time   `json:"approved_at"`
	RejectedBy      *string      `json:"rejected_by"`
	RejectedAt      *time.Time   `json:"rejected_at"`
}

type Meta struct {
	CurrentPage int `json:"current_page"`
	FirstItem   int `json:"first_item"`
	LastItem    int `json:"last_item"`
	LastPage    int `json:"last_page"`
	Total       int `json:"total"`
}

type Data struct {
	Data []User `json:"data"`
	Meta Meta   `json:"meta"`
}

type Root struct {
	Data Data `json:"data"`
}
