package model

import "time"

// WorkflowScheduleStatus transitions: active -> canceled or
// active -> executed (executed applies to once only; cron stays active).
type WorkflowScheduleStatus string

const (
	WorkflowScheduleStatusActive   WorkflowScheduleStatus = "active"
	WorkflowScheduleStatusCanceled WorkflowScheduleStatus = "canceled"
	WorkflowScheduleStatusExecuted WorkflowScheduleStatus = "executed"
)

type WorkflowScheduleType string

const (
	WorkflowScheduleTypeOnce WorkflowScheduleType = "once"
	WorkflowScheduleTypeCron WorkflowScheduleType = "cron"
)

// WorkflowSchedule persists the schedule intent; Airflow does the actual
// triggering via DAG metadata. Exactly one of RunAt / Cron is set per Type.
// At most one row per workflow_id is in active state at any time.
type WorkflowSchedule struct {
	ID         string                 `gorm:"primaryKey;column:id" json:"id"`
	WorkflowID string                 `gorm:"column:workflow_id;index;not null" json:"workflow_id"`
	Type       WorkflowScheduleType   `gorm:"column:type;not null;default:once" json:"type"`
	RunAt      *time.Time             `gorm:"column:run_at" json:"run_at,omitempty"`
	Cron       *string                `gorm:"column:cron" json:"cron,omitempty"`
	Status     WorkflowScheduleStatus `gorm:"column:status;not null;default:active" json:"status"`
	CreatedAt  time.Time              `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time              `gorm:"autoUpdateTime" json:"updated_at"`
}

func (WorkflowSchedule) TableName() string {
	return "workflow_schedules"
}

// CreateWorkflowScheduleReq is the POST /schedule body. Exactly one of
// RunAt (one-shot) or Cron (recurring) must be set.
type CreateWorkflowScheduleReq struct {
	RunAt *time.Time `json:"run_at,omitempty" mapstructure:"run_at"`
	Cron  *string    `json:"cron,omitempty" mapstructure:"cron"`
}
