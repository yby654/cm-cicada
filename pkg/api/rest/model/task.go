package model

type CreateTaskReq struct {
	Name          string                 `json:"name" mapstructure:"name" validate:"required"`
	TaskComponent string                 `json:"task_component" mapstructure:"task_component" validate:"required"`
	RequestBody   string                 `json:"request_body" mapstructure:"request_body" validate:"required"`
	PathParams    map[string]string      `json:"path_params" mapstructure:"path_params"`
	QueryParams   map[string]string      `json:"query_params" mapstructure:"query_params"`
	Extra         map[string]interface{} `json:"extra,omitempty" mapstructure:"extra"`
	Dependencies  []string               `json:"dependencies" mapstructure:"dependencies"`
}
type Task struct {
	ID            string                 `gorm:"primaryKey" json:"id" mapstructure:"id" validate:"required"`
	WorkflowID    string                 `json:"workflow_id" mapstructure:"workflow_id" validate:"required"`
	TaskGroupID   string                 `json:"task_group_id" mapstructure:"task_group_id" validate:"required"`
	Name          string                 `json:"name" mapstructure:"name" validate:"required"`
	TaskComponent string                 `json:"task_component" mapstructure:"task_component" validate:"required"`
	RequestBody   string                 `json:"request_body" mapstructure:"request_body" validate:"required"`
	PathParams    map[string]string      `json:"path_params" mapstructure:"path_params"`
	QueryParams   map[string]string      `json:"query_params" mapstructure:"query_params"`
	Extra         map[string]interface{} `json:"extra,omitempty" mapstructure:"extra"`
	Dependencies  []string               `json:"dependencies" mapstructure:"dependencies"`
}
