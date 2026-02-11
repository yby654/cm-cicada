package domain

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

// ============================================================================
// DB Models (GORM 태그 포함 - DB와 직접 매핑)
// ============================================================================

// Workflow is a DB model for workflow table.
// Data is the JSON snapshot (embedded TaskGroups/Tasks); separate task_group/task tables
// reference this via TaskGroup.WorkflowVersionID (= Workflow.ID) and Task.WorkflowID/TaskGroupID.
type Workflow struct {
	ID          string    `gorm:"primaryKey" json:"id" mapstructure:"id" validate:"required"`
	SpecVersion string    `gorm:"column:spec_version" json:"spec_version" mapstructure:"spec_version" validate:"required"`
	Name        string    `gorm:"index:,column:name,unique;type:text collate nocase" json:"name" mapstructure:"name" validate:"required"`
	Data        Data      `gorm:"column:data" json:"data" mapstructure:"data" validate:"required"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime:false" json:"created_at" mapstructure:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at;autoCreateTime:false" json:"updated_at" mapstructure:"updated_at"`
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

// Value implements driver.Valuer interface for Workflow
func (w Workflow) Value() (driver.Value, error) {
	return json.Marshal(w)
}

// Scan implements sql.Scanner interface for Workflow
func (w *Workflow) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("invalid type for Workflow")
	}
	return json.Unmarshal(bytes, w)
}

// BeforeDelete Hook for Workflow to delete WorkflowVersion on delete
func (w *Workflow) BeforeDelete(tx *gorm.DB) (err error) {
	if err := tx.Where("workflow_id = ?", w.ID).Delete(&WorkflowVersion{}).Error; err != nil {
		return err
	}
	return nil
}
