package domain

import "time"

type GetWorkflowTemplate struct {
	SpecVersion string `gorm:"column:spec_version" json:"spec_version" mapstructure:"spec_version" validate:"required"`
	Name        string `gorm:"index:,column:name,unique;type:text collate nocase" json:"name" mapstructure:"name" validate:"required"`
	Data        Data   `gorm:"column:data" json:"data" mapstructure:"data" validate:"required"`
}

type WorkflowTemplate struct {
	ID          string    `gorm:"primaryKey" json:"id" mapstructure:"id" validate:"required"`
	Name        string    `gorm:"index:,column:name,unique;type:text collate nocase" json:"name" mapstructure:"name" validate:"required"`
	SpecVersion string    `gorm:"column:spec_version;not null" json:"spec_version" mapstructure:"spec_version" validate:"required"`
	Data        Data      `gorm:"column:data" json:"data" mapstructure:"data" validate:"required"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at" mapstructure:"created_at"`
}
