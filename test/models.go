package main

import (
	"time"

	"github.com/cristianob/erudito"
	"github.com/cristianob/erudito/nulls"
)

/*
 ******** A ********
 */
type A struct {
	erudito.FullModel

	Field1  string        `json:"field1" gorm:"NOT NULL"`
	Field2  int           `json:"field2" gorm:"NOT NULL"`
	Field3  float32       `json:"field3" gorm:"NOT NULL"`
	Field4  string        `json:"field4" gorm:"type:JSON;NOT NULL" eruditoType:"JSON"`
	Field5  nulls.String  `json:"field5"`
	Field6  nulls.Int32   `json:"field6"`
	Field7  nulls.Float32 `json:"field7"`
	Field8  nulls.String  `json:"field8" gorm:"type:JSON" eruditoType:"JSON"`
	Field9  time.Time     `json:"field9"`
	Field10 nulls.Time    `json:"field10"`

	Bs []B `json:"bs,omitempty" gorm:"foreignkey:AID"`

	ignore string
}

func (A) CRUDOptions() erudito.CRUDOptions {
	return erudito.CRUDOptions{
		ModelSingular: "a",
		ModelPlural:   "as",
	}
}

func (A) MiddlewareBefore() []erudito.MiddlewareBefore {
	return []erudito.MiddlewareBefore{
		globalMiddleware1,
		middlewareA1,
		//blockMiddleware,
	}
}

func (A) MiddlewareAfter() []erudito.MiddlewareAfter {
	return []erudito.MiddlewareAfter{
		middlewareA2,
	}
}

func (A) TableName() string {
	return "a"
}

/*
 ******** B ********
 */
type B struct {
	erudito.FullModel

	Field1 string        `json:"field1"`
	Field2 int           `json:"field2"`
	Field3 float32       `json:"field3"`
	Field4 string        `json:"field4" gorm:"type:JSON" eruditoType:"JSON"`
	Field5 nulls.String  `json:"field5"`
	Field6 nulls.Int32   `json:"field6"`
	Field7 nulls.Float32 `json:"field7"`
	Field8 nulls.String  `json:"field8" gorm:"type:JSON" eruditoType:"JSON"`

	A   *A   `json:"a,omitempty" gorm:"foreignkey:AID"`
	AID uint `json:"a_id" eruditoType:"REL_ID"`

	Cs []C `json:"cs,omitempty" gorm:"many2many:b_c;" erudito:"PUTautoremove"`

	E   *E           `json:"e,omitempty" gorm:"foreignkey:EID"`
	EID nulls.UInt32 `json:"e_id" eruditoType:"REL_ID"`
}

func (B) CRUDOptions() erudito.CRUDOptions {
	return erudito.CRUDOptions{
		ModelSingular: "b",
		ModelPlural:   "bs",
	}
}

func (B) MiddlewareBefore() []erudito.MiddlewareBefore {
	return []erudito.MiddlewareBefore{
		globalMiddleware1,
		middlewareB1,
	}
}

func (B) MiddlewareAfter() []erudito.MiddlewareAfter {
	return []erudito.MiddlewareAfter{}
}

func (B) TableName() string {
	return "b"
}

/*
 ******** C ********
 */
type C struct {
	erudito.FullModel

	Field1 string        `json:"field1"`
	Field2 int           `json:"field2"`
	Field3 float32       `json:"field3"`
	Field4 string        `json:"field4" gorm:"type:JSON" eruditoType:"JSON"`
	Field5 nulls.String  `json:"field5"`
	Field6 nulls.Int32   `json:"field6"`
	Field7 nulls.Float32 `json:"field7"`
	Field8 nulls.String  `json:"field8" gorm:"type:JSON" eruditoType:"JSON"`

	Bs []B `json:"bs,omitempty" gorm:"many2many:b_c;"`
	Ds []D `json:"ds,omitempty" gorm:"many2many:c_d;" erudito:"PUTautodelete"`
}

func (C) CRUDOptions() erudito.CRUDOptions {
	return erudito.CRUDOptions{
		ModelSingular: "c",
		ModelPlural:   "cs",
	}
}

func (C) MiddlewareBefore() []erudito.MiddlewareBefore {
	return []erudito.MiddlewareBefore{
		globalMiddleware1,
	}
}

func (C) MiddlewareAfter() []erudito.MiddlewareAfter {
	return []erudito.MiddlewareAfter{}
}

func (C) TableName() string {
	return "c"
}

/*
 ******** D ********
 */
type D struct {
	erudito.FullModel

	Field1 string        `json:"field1"`
	Field2 int           `json:"field2"`
	Field3 float32       `json:"field3"`
	Field4 string        `json:"field4" gorm:"type:JSON" eruditoType:"JSON"`
	Field5 nulls.String  `json:"field5"`
	Field6 nulls.Int32   `json:"field6"`
	Field7 nulls.Float32 `json:"field7"`
	Field8 nulls.String  `json:"field8" gorm:"type:JSON" eruditoType:"JSON"`

	Cs []C `json:"cs,omitempty" gorm:"many2many:c_d;"`
}

func (D) CRUDOptions() erudito.CRUDOptions {
	return erudito.CRUDOptions{
		ModelSingular: "d",
		ModelPlural:   "ds",
	}
}

func (D) MiddlewareBefore() []erudito.MiddlewareBefore {
	return []erudito.MiddlewareBefore{
		globalMiddleware1,
	}
}

func (D) MiddlewareAfter() []erudito.MiddlewareAfter {
	return []erudito.MiddlewareAfter{}
}

func (D) TableName() string {
	return "d"
}

/*
 ******** E ********
 */
type E struct {
	erudito.FullModel

	Field1 string        `json:"field1"`
	Field2 int           `json:"field2"`
	Field3 float32       `json:"field3"`
	Field4 string        `json:"field4" gorm:"type:JSON" eruditoType:"JSON"`
	Field5 string        `json:"field5" gorm:"type:JSON" eruditoType:"JSON"`
	Field6 nulls.String  `json:"field6"`
	Field7 nulls.Int32   `json:"field7"`
	Field8 nulls.Float32 `json:"field8"`
	Field9 nulls.String  `json:"field9" gorm:"type:JSON" eruditoType:"JSON"`
}

func (E) CRUDOptions() erudito.CRUDOptions {
	return erudito.CRUDOptions{
		ModelSingular: "e",
		ModelPlural:   "es",
	}
}

func (E) MiddlewareBefore() []erudito.MiddlewareBefore {
	return []erudito.MiddlewareBefore{
		globalMiddleware1,
	}
}

func (E) MiddlewareAfter() []erudito.MiddlewareAfter {
	return []erudito.MiddlewareAfter{}
}

func (E) TableName() string {
	return "e"
}
