package domain

import (
	"encoding/json"
	"time"
)

// Task belongs to Workflow and TaskGroup (FK: WorkflowID, TaskGroupID).
type Task struct {
	ID            string          `gorm:"primaryKey" json:"id" mapstructure:"id" validate:"required"`
	Name          string          `gorm:"column:name" json:"name" mapstructure:"name" validate:"required"`
	WorkflowID    string          `gorm:"column:workflow_id;index" json:"workflow_id" mapstructure:"workflow_id" validate:"required"`
	TaskGroupID   string          `gorm:"column:task_group_id;index" json:"task_group_id" mapstructure:"task_group_id" validate:"required"`
	TaskComponent string          `gorm:"column:task_component" json:"task_component" mapstructure:"task_component" validate:"required"`
	RequestBody   string          `gorm:"column:request_body" json:"request_body" mapstructure:"request_body" validate:"required"`
	PathParams    json.RawMessage `gorm:"column:path_params;type:blob" json:"path_params" mapstructure:"path_params"`
	QueryParams   json.RawMessage `gorm:"column:query_params;type:blob" json:"query_params" mapstructure:"query_params"`
	Extra         json.RawMessage `gorm:"column:extra;type:blob" json:"extra" mapstructure:"extra"`
	Dependencies  json.RawMessage `gorm:"column:dependencies;type:blob" json:"dependencies" mapstructure:"dependencies"`
	CreatedAt     time.Time       `gorm:"column:created_at;autoCreateTime" json:"created_at" mapstructure:"created_at"`
	DeletedAt     *time.Time      `gorm:"column:deleted_at;index" json:"-" mapstructure:"-"` // soft delete (버전 업데이트 시 삭제분)

	// 참조 관계 (Preload 시 사용, JSON 직렬화 제외)
	Workflow  *Workflow  `gorm:"foreignKey:WorkflowID;references:ID" json:"-"`
	TaskGroup *TaskGroup `gorm:"foreignKey:TaskGroupID;references:ID" json:"-"`
}

// type TaskDirectly struct {
// 	ID            string                 `gorm:"primaryKey" json:"id" mapstructure:"id" validate:"required"`
// 	WorkflowID    string                 `json:"workflow_id" mapstructure:"workflow_id" validate:"required"`
// 	TaskGroupID   string                 `json:"task_group_id" mapstructure:"task_group_id" validate:"required"`
// 	Name          string                 `json:"name" mapstructure:"name" validate:"required"`
// 	TaskComponent string                 `json:"task_component" mapstructure:"task_component" validate:"required"`
// 	RequestBody   string                 `json:"request_body" mapstructure:"request_body" validate:"required"`
// 	PathParams    map[string]string      `json:"path_params" mapstructure:"path_params"`
// 	QueryParams   map[string]string      `json:"query_params" mapstructure:"query_params"`
// 	Extra         map[string]interface{} `json:"extra,omitempty" mapstructure:"extra"`
// 	Dependencies  []string               `json:"dependencies" mapstructure:"dependencies"`
// }

// type TaskDBModel struct {
// 	ID          string `gorm:"primaryKey" json:"id" mapstructure:"id" validate:"required"`
// 	Name        string `json:"name" mapstructure:"name" validate:"required"`
// 	WorkflowID  string `gorm:"column:workflow_id" json:"workflow_id" mapstructure:"workflow_id" validate:"required"`
// 	TaskGroupID string `gorm:"column:task_group_id" json:"task_group_id" mapstructure:"task_group_id" validate:"required"`
// }

// type CreateTaskReq struct {
// 	Name          string                 `json:"name" mapstructure:"name" validate:"required"`
// 	TaskComponent string                 `json:"task_component" mapstructure:"task_component" validate:"required"`
// 	RequestBody   string                 `json:"request_body" mapstructure:"request_body" validate:"required"`
// 	PathParams    map[string]string      `json:"path_params" mapstructure:"path_params"`
// 	QueryParams   map[string]string      `json:"query_params" mapstructure:"query_params"`
// 	Extra         map[string]interface{} `json:"extra,omitempty" mapstructure:"extra"`
// 	Dependencies  []string               `json:"dependencies" mapstructure:"dependencies"`
// }
