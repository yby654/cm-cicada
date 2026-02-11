package model

import (
	"encoding/json"
	"time"
)

type WorkflowVersion struct {
	ID                 string          `json:"id" mapstructure:"id" validate:"required"`
	WorkflowID         string          ` json:"workflow_id" mapstructure:"workflow_id" validate:"required"`
	Version            int             ` json:"version" mapstructure:"version" validate:"required"`
	SpecVersion        string          ` json:"spec_version" mapstructure:"spec_version" validate:"required"`
	DefinitionSnapshot json.RawMessage `json:"definition_snapshot" mapstructure:"definition_snapshot"`
	CreatedAt          time.Time       ` json:"created_at" mapstructure:"created_at"`
}

type WorkflowVersionSwg struct {
	ID                 string                 `json:"id" mapstructure:"id" validate:"required"`
	WorkflowID         string                 ` json:"workflow_id" mapstructure:"workflow_id" validate:"required"`
	Version            int                    ` json:"version" mapstructure:"version" validate:"required"`
	SpecVersion        string                 ` json:"spec_version" mapstructure:"spec_version" validate:"required"`
	DefinitionSnapshot map[string]interface{} `json:"definition_snapshot" mapstructure:"definition_snapshot"`
	CreatedAt          time.Time              ` json:"created_at" mapstructure:"created_at"`
}
