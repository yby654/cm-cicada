package dao

import (
	"errors"

	"gorm.io/gorm"

	"github.com/cloud-barista/cm-cicada/db"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
)

func WorkflowScheduleCreate(s *model.WorkflowSchedule) error {
	return WorkflowScheduleCreateTx(db.DB, s)
}

// WorkflowScheduleCreateTx is the transaction-aware variant of
// WorkflowScheduleCreate.
func WorkflowScheduleCreateTx(tx *gorm.DB, s *model.WorkflowSchedule) error {
	if tx == nil {
		return errors.New("database connection is not initialized")
	}
	return tx.Create(s).Error
}

// WorkflowScheduleGetByID returns (nil, nil) when the row does not exist.
func WorkflowScheduleGetByID(id string) (*model.WorkflowSchedule, error) {
	var s model.WorkflowSchedule
	if err := db.DB.Where("id = ?", id).First(&s).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

// WorkflowScheduleListByWorkflowID returns all rows ordered by run_at ASC.
func WorkflowScheduleListByWorkflowID(workflowID string) ([]model.WorkflowSchedule, error) {
	var out []model.WorkflowSchedule
	if err := db.DB.Where("workflow_id = ?", workflowID).
		Order("run_at ASC").
		Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

// WorkflowScheduleGetLatest returns the most recently created row regardless
// of status, or (nil, nil) when there's no schedule history.
func WorkflowScheduleGetLatest(workflowID string) (*model.WorkflowSchedule, error) {
	var s model.WorkflowSchedule
	err := db.DB.Where("workflow_id = ?", workflowID).
		Order("created_at DESC").
		First(&s).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

// WorkflowScheduleGetActive returns the row that drives DAG metadata, or
// (nil, nil) when none exists. Picks earliest run_at if multiple actives
// somehow exist (service guards against that).
func WorkflowScheduleGetActive(workflowID string) (*model.WorkflowSchedule, error) {
	return WorkflowScheduleGetActiveTx(db.DB, workflowID)
}

// WorkflowScheduleGetActiveTx is the transaction-aware variant of
// WorkflowScheduleGetActive. Used inside the schedule-create transaction to
// re-check active uniqueness on the same connection as the insert.
func WorkflowScheduleGetActiveTx(tx *gorm.DB, workflowID string) (*model.WorkflowSchedule, error) {
	if tx == nil {
		return nil, errors.New("database connection is not initialized")
	}
	var s model.WorkflowSchedule
	err := tx.Where("workflow_id = ? AND status = ?", workflowID, model.WorkflowScheduleStatusActive).
		Order("run_at ASC").
		First(&s).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

func WorkflowScheduleUpdateStatus(id string, status model.WorkflowScheduleStatus) error {
	return db.DB.Model(&model.WorkflowSchedule{}).
		Where("id = ?", id).
		Update("status", status).Error
}
