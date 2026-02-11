package domain

import (
	"time"
)

type WorkflowVersion struct {
	ID                 string    `gorm:"primaryKey" json:"id" mapstructure:"id" validate:"required"`
	WorkflowID         string    `gorm:"column:workflow_id;not null;uniqueIndex:idx_workflow_version" json:"workflow_id" mapstructure:"workflow_id" validate:"required"`
	Version            int       `gorm:"column:version;not null;uniqueIndex:idx_workflow_version" json:"version" mapstructure:"version" validate:"required"`
	SpecVersion        string    `gorm:"column:spec_version;not null" json:"spec_version" mapstructure:"spec_version" validate:"required"`
	DefinitionSnapshot []byte    `gorm:"column:definition_snapshot;type:blob" json:"definition_snapshot" mapstructure:"definition_snapshot"`
	CreatedAt          time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at" mapstructure:"created_at"`
	Workflow           Workflow  `gorm:"foreignKey:WorkflowID;references:ID"`
}
