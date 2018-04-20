package erudito

import "time"

type Model interface {
	ModelSingular() string
	ModelPlural() string

	AcceptCollection() bool
	AcceptGET() bool
	AcceptPOST() bool
	AcceptPUT() bool
	AcceptDELETE() bool
}

type FullModel struct {
	ID        uint       `json:"id" gorm:"primary_key"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at" sql:"index"`
}

type NoDeleteModel struct {
	ID        uint      `json:"id" gorm:"primary_key"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
