package domain

import (
	"time"
)

// WorkflowRun belongs to Workflow (FK: WorkflowID).
type WorkflowRun struct {
	ID              string    `gorm:"primaryKey" json:"id" mapstructure:"id" validate:"required"`
	WorkflowID      string    `gorm:"column:workflow_id;index" json:"workflow_id" mapstructure:"workflow_id"`
	ExecutionDate   time.Time `gorm:"column:execution_date" json:"execution_date" mapstructure:"execution_date"`
	StartDate       time.Time `gorm:"column:start_date" json:"start_date" mapstructure:"start_date"`
	EndDate         time.Time `gorm:"column:end_date" json:"end_date" mapstructure:"end_date"`
	DurationDate    float64   `gorm:"column:duration_date" json:"duration_date" mapstructure:"duration_date"`
	RunType         string    `gorm:"column:run_type" json:"run_type" mapstructure:"run_type"`
	State           string    `gorm:"column:state" json:"state" mapstructure:"state"`
	ExternalTrigger *bool     `gorm:"column:external_trigger" json:"external_trigger" mapstructure:"external_trigger"`

	Workflow *Workflow `gorm:"foreignKey:WorkflowID;references:ID" json:"-"`
}
