package erudito

import (
	"time"
)

const (
	MODEL_TYPE_FULL       = 1
	MODEL_TYPE_HARDDELETE = 2
	MODEL_TYPE_SIMPLE     = 3
)

const (
	FIELD_TYPE_INTERNAL                    = 1
	FIELD_TYPE_COMMON                      = 2
	FIELD_TYPE_COMMON_TIME                 = 3
	FIELD_TYPE_COMMON_JSON                 = 4
	FIELD_TYPE_RELATION_ID                 = 5
	FIELD_TYPE_RELATION_MODEL              = 6
	FIELD_TYPE_RELATION_COL_IDS            = 7
	FIELD_TYPE_RELATION_COL_IDS_UPDATE     = 8
	FIELD_TYPE_RELATION_COL_IDS_REPLACE    = 9
	FIELD_TYPE_RELATION_COL_IDS_REMOVE     = 10
	FIELD_TYPE_RELATION_COL_MODELS         = 11
	FIELD_TYPE_RELATION_COL_MODELS_UPDATE  = 12
	FIELD_TYPE_RELATION_COL_MODELS_REPLACE = 13
	FIELD_TYPE_RELATION_COL_MODELS_REMOVE  = 14
)

type (
	Model interface {
		CRUDOptions() CRUDOptions
		MiddlewaresPRE() []MiddlewarePRE
	}

	FullModel struct {
		ID        uint       `json:"id" gorm:"primary_key"`
		CreatedAt *time.Time `json:"created_at"`
		UpdatedAt *time.Time `json:"updated_at"`
		DeletedAt *time.Time `json:"deleted_at" sql:"index"`
	}

	HardDeleteModel struct {
		ID        uint       `json:"id" gorm:"primary_key"`
		CreatedAt *time.Time `json:"created_at"`
		UpdatedAt *time.Time `json:"updated_at"`
	}

	SimpleModel struct {
		ID uint `json:"id" gorm:"primary_key"`
	}

	/*
	 * Options
	 */
	CRUDOptions struct {
		ModelSingular string
		ModelPlural   string
	}
)

/*
 * STRUCTURES
 */
type fieldStructure struct {
	Name            string
	JsonName        string
	Type            int
	Nullable        bool
	RelationalModel string
}

type modelStructure struct {
	Name     string
	Type     int
	Singular string
	Plural   string
	Model    Model
	Fields   []fieldStructure
}

func (m modelStructure) getFieldByJson(json string) *fieldStructure {
	for _, field := range m.Fields {
		if field.JsonName == json {
			return &field
		}
	}

	return nil
}
