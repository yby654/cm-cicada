package adapter

import (
	"errors"

	"github.com/cloud-barista/cm-cicada/internal/domain"
	"gorm.io/gorm"
)

func WorkflowCreateTx(tx *gorm.DB, workflow *domain.Workflow) (*domain.Workflow, error) {
	if tx == nil {
		return nil, errors.New("transaction is nil")
	}
	if err := tx.Create(workflow).Error; err != nil {
		return nil, err
	}
	return workflow, nil
}

func TaskGroupCreateTx(tx *gorm.DB, taskGroup *domain.TaskGroup) (*domain.TaskGroup, error) {
	if tx == nil {
		return nil, errors.New("transaction is nil")
	}
	if err := tx.Create(taskGroup).Error; err != nil {
		return nil, err
	}
	return taskGroup, nil
}

func TaskCreateTx(tx *gorm.DB, task *domain.Task) (*domain.Task, error) {
	if tx == nil {
		return nil, errors.New("transaction is nil")
	}
	if err := tx.Create(task).Error; err != nil {
		return nil, err
	}
	return task, nil
}
