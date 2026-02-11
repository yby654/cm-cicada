package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

const (
	WorkflowSpecVersion_1_0 = "1.0"
)

const (
	WorkflowSpecVersion_LATEST = WorkflowSpecVersion_1_0
)

// Data is a common type used in both DB models and DTOs

// CreateDataReq is a common type for creating workflow data
type CreateDataReq struct {
	Description string               `json:"description" mapstructure:"description"`
	TaskGroups  []CreateTaskGroupReq `json:"task_groups" mapstructure:"task_groups" validate:"required"`
}

func (d CreateDataReq) Value() (driver.Value, error) {
	return json.Marshal(d)
}

func (d *CreateDataReq) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("invalid type for CreateDataReq")
	}
	return json.Unmarshal(bytes, d)
}

type CreateWorkflowReq struct {
	SpecVersion string        `json:"spec_version" mapstructure:"spec_version"`
	Name        string        `json:"name" mapstructure:"name" validate:"required"`
	Data        CreateDataReq `json:"data" mapstructure:"data" validate:"required"`
}

// UpdateWorkflowReq is a DTO for updating workflow requests
type UpdateWorkflowReq struct {
	SpecVersion string        `json:"spec_version,omitempty" mapstructure:"spec_version"`
	Name        string        `json:"name,omitempty" mapstructure:"name"`
	Data        CreateDataReq `json:"data,omitempty" mapstructure:"data"`
}

type WorkflowStatus struct {
	State string `json:"state"`
	Count int    `json:"count"`
}

type Workflow struct {
	ID          string    ` json:"id" mapstructure:"id" validate:"required"`
	SpecVersion string    ` json:"spec_version" mapstructure:"spec_version" validate:"required"`
	Name        string    `json:"name" mapstructure:"name" validate:"required"`
	Data        Data      ` json:"data" mapstructure:"data" validate:"required"`
	CreatedAt   time.Time ` json:"created_at" mapstructure:"created_at"`
	UpdatedAt   time.Time `  json:"updated_at" mapstructure:"updated_at"`
}
