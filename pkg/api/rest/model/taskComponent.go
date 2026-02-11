package model

import (
	"github.com/cloud-barista/cm-cicada/internal/domain"
)

type CreateTaskComponentReq struct {
	Name string                   `json:"name" mapstructure:"name" validate:"required"`
	Data domain.TaskComponentData `gorm:"column:data" json:"data" mapstructure:"data" validate:"required"`
}
