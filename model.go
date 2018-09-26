package erudito

import (
	"encoding/json"
	"time"

	"github.com/go-sql-driver/mysql"
)

type Model interface {
	CRUDOptions() CRUDOptions
}

type FullModel struct {
	ID        uint       `json:"id" gorm:"primary_key" erudito:"excludePOST;excludePUT"`
	CreatedAt time.Time  `json:"created_at" erudito:"excludePOST;excludePUT"`
	UpdatedAt time.Time  `json:"updated_at" erudito:"excludePOST;excludePUT"`
	DeletedAt *time.Time `json:"deleted_at" sql:"index" erudito:"excludePOST;excludePUT"`
}

type HardDeleteModel struct {
	ID        uint      `json:"id" gorm:"primary_key" erudito:"excludePOST;excludePUT"`
	CreatedAt time.Time `json:"created_at" erudito:"excludePOST;excludePUT"`
	UpdatedAt time.Time `json:"updated_at" erudito:"excludePOST;excludePUT"`
}

type SimpleModel struct {
	ID uint `json:"id" gorm:"primary_key" erudito:"excludePOST;excludePUT"`
}

/*
 *
 */
type CRUDOptions struct {
	ModelSingular    string
	ModelPlural      string
	AcceptCollection bool
	AcceptGET        bool
	AcceptPOST       bool
	AcceptPUT        bool
	AcceptPATCH      bool
	AcceptDELETE     bool
}

/*
 * NULL TYPES
 */
type NullTime struct {
	mysql.NullTime
}

func (v NullTime) MarshalJSON() ([]byte, error) {
	if v.Valid {
		return json.Marshal(v.Time)
	} else {
		return json.Marshal(nil)
	}
}

func (v *NullTime) UnmarshalJSON(data []byte) error {
	var x *time.Time

	if err := json.Unmarshal(data, &x); err != nil {
		return err
	}

	if x != nil {
		v.Valid = true
		v.Time = *x
	} else {
		v.Valid = false
	}

	return nil
}
