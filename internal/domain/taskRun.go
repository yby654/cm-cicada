package domain

import (
	"time"
)

// TaskRun belongs to WorkflowRun and Task (FK: WorkflowRunID, TaskID).
type TaskRun struct {
	ID            string    `gorm:"primaryKey" json:"id" mapstructure:"id" validate:"required"`
	WorkflowRunID string    `gorm:"column:workflow_run_id;index" json:"workflow_run_id" mapstructure:"workflow_run_id" validate:"required"`
	TaskID        string    `gorm:"column:task_id;index" json:"task_id" mapstructure:"task_id" validate:"required"`
	State         string    `gorm:"column:state" json:"state" mapstructure:"state"`
	StartDate     time.Time `gorm:"column:start_date" json:"start_date" mapstructure:"start_date"`
	EndDate       time.Time `gorm:"column:end_date" json:"end_date" mapstructure:"end_date"`
	DurationDate  float64   `gorm:"column:duration_date" json:"duration_date" mapstructure:"duration_date"`
	TryNumber     int       `gorm:"column:try_number" json:"try_number" mapstructure:"try_number"`
	CreatedAt     time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at" mapstructure:"created_at"`

	WorkflowRun *WorkflowRun `gorm:"foreignKey:WorkflowRunID;references:ID" json:"-"`
	Task        *Task        `gorm:"foreignKey:TaskID;references:ID" json:"-"`
}
