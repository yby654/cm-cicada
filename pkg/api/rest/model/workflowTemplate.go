package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

type WorkflowTemplate struct {
	SpecVersion string `json:"spec_version" mapstructure:"spec_version" validate:"required"`
	Name        string `json:"name" mapstructure:"name" validate:"required"`
	Data        Data   `json:"data" mapstructure:"data" validate:"required"`
}

type Data struct {
	Description string      `json:"description" mapstructure:"description"`
	TaskGroups  []TaskGroup `json:"task_groups" mapstructure:"task_groups" validate:"required"`
}

func (d Data) Value() (driver.Value, error) {
	return json.Marshal(d)
}

func (d *Data) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("invalid type for Data")
	}
	return json.Unmarshal(bytes, d)
}
