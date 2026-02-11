package domain

import "time"

// TaskGroup belongs to Workflow (WorkflowVersionID stores Workflow.ID in current design).
// Tasks here is for JSON embedding in Workflow.Data; DB relation to task table uses Task.TaskGroupID.
type TaskGroup struct {
	ID                string    `gorm:"primaryKey" json:"id" mapstructure:"id" validate:"required"`
	Name              string    `gorm:"column:name" json:"name" mapstructure:"name" validate:"required"`
	WorkflowVersionID string    `gorm:"column:workflow_version_id;index" json:"workflow_version_id" mapstructure:"workflow_version_id" validate:"required"`
	Description       string     `gorm:"column:description" json:"description" mapstructure:"description"`
	Tasks             []Task     `json:"tasks,omitempty" mapstructure:"tasks"` // JSON용 필드 (GORM 태그 없음)
	CreatedAt         time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at" mapstructure:"created_at"`
	DeletedAt         *time.Time `gorm:"column:deleted_at;index" json:"-" mapstructure:"-"` // soft delete (버전 업데이트 시 삭제분)

	// 참조 관계 (Preload 시 사용, JSON 직렬화 제외)
	Workflow *Workflow `gorm:"foreignKey:WorkflowVersionID;references:ID" json:"-"`
}
