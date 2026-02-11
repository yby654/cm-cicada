package adapter

import (
	"errors"
	"time"

	"github.com/cloud-barista/cm-cicada/internal/domain"
	"gorm.io/gorm"
)

// WorkflowVersionMaxVersionTx returns the maximum version number for the given workflow_id within the transaction.
// Returns 0 if no WorkflowVersion exists for the workflow.
func WorkflowVersionMaxVersionTx(tx *gorm.DB, workflowID string) (int, error) {
	if tx == nil {
		return 0, errors.New("transaction is nil")
	}
	var maxVer *int
	err := tx.Model(&domain.WorkflowVersion{}).Where("workflow_id = ?", workflowID).Select("MAX(version)").Scan(&maxVer).Error
	if err != nil {
		return 0, err
	}
	if maxVer == nil {
		return 0, nil
	}
	return *maxVer, nil
}

// WorkflowVersionCreateTx creates a WorkflowVersion within the given transaction.
func WorkflowVersionCreateTx(tx *gorm.DB, workflowVersion *domain.WorkflowVersion) (*domain.WorkflowVersion, error) {
	if tx == nil {
		return nil, errors.New("transaction is nil")
	}
	if err := tx.Create(workflowVersion).Error; err != nil {
		return nil, err
	}
	return workflowVersion, nil
}

// WorkflowUpdateTx updates a Workflow row within the given transaction.
func WorkflowUpdateTx(tx *gorm.DB, workflow *domain.Workflow) error {
	if tx == nil {
		return errors.New("transaction is nil")
	}
	workflow.UpdatedAt = time.Now()
	if err := tx.Model(&domain.Workflow{}).Where("id = ?", workflow.ID).Updates(workflow).Error; err != nil {
		return err
	}
	return nil
}
