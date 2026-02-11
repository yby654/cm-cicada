package adapter

import (
	"errors"

	"github.com/cloud-barista/cm-cicada/internal/db"
	"github.com/cloud-barista/cm-cicada/internal/domain"

	"gorm.io/gorm"
)

func WorkflowRunCreate(workflowRun *domain.WorkflowRun) (*domain.WorkflowRun, error) {
	// Ensure db.DB is not nil to avoid runtime panics
	if db.DB == nil {
		return nil, errors.New("database connection is not initialized")
	}

	result := db.DB.Create(workflowRun)
	err := result.Error
	if err != nil {
		return nil, err
	}

	return workflowRun, nil
}

func WorkflowRunGet(id string) (*domain.WorkflowRun, error) {
	workflowRun := &domain.WorkflowRun{}

	// Ensure db.DB is not nil to avoid runtime panics
	if db.DB == nil {
		return nil, errors.New("database connection is not initialized")
	}

	result := db.DB.Where("id = ?", id).First(workflowRun)
	err := result.Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("workflow run not found with the provided id")
		}
		return nil, err
	}

	return workflowRun, nil
}

func WorkflowRunGetByWorkflowID(workflowID string) (*[]domain.WorkflowRun, error) {
	workflowRunList := &[]domain.WorkflowRun{}

	// Ensure db.DB is not nil to avoid runtime panics
	if db.DB == nil {
		return nil, errors.New("database connection is not initialized")
	}

	result := db.DB.Where("workflow_id = ?", workflowID).Find(workflowRunList)
	err := result.Error
	if err != nil {
		return nil, err
	}

	return workflowRunList, nil
}

func WorkflowRunGetList(workflowRun *domain.WorkflowRun, page int, row int) (*[]domain.WorkflowRun, error) {
	workflowRunList := &[]domain.WorkflowRun{}
	// Ensure db.DB is not nil to avoid runtime panics
	if db.DB == nil {
		return nil, errors.New("database connection is not initialized")
	}

	result := db.DB.Scopes(func(d *gorm.DB) *gorm.DB {
		var filtered = d

		if len(workflowRun.WorkflowID) != 0 {
			filtered = filtered.Where("workflow_id = ?", workflowRun.WorkflowID)
		}

		if len(workflowRun.State) != 0 {
			filtered = filtered.Where("state = ?", workflowRun.State)
		}

		if len(workflowRun.RunType) != 0 {
			filtered = filtered.Where("run_type = ?", workflowRun.RunType)
		}

		if page != 0 && row != 0 {
			offset := (page - 1) * row
			return filtered.Offset(offset).Limit(row)
		} else if row != 0 && page == 0 {
			filtered.Error = errors.New("row is not 0 but page is 0")
			return filtered
		} else if page != 0 && row == 0 {
			filtered.Error = errors.New("page is not 0 but row is 0")
			return filtered
		}
		return filtered
	}).Find(workflowRunList)

	err := result.Error
	if err != nil {
		return nil, err
	}

	return workflowRunList, nil
}

func WorkflowRunUpdate(workflowRun *domain.WorkflowRun) error {
	// Ensure db.DB is not nil to avoid runtime panics
	if db.DB == nil {
		return errors.New("database connection is not initialized")
	}

	result := db.DB.Model(&domain.WorkflowRun{}).Where("id = ?", workflowRun.ID).Updates(workflowRun)
	err := result.Error
	if err != nil {
		return err
	}

	return nil
}

func WorkflowRunDelete(workflowRun *domain.WorkflowRun) error {
	// Ensure db.DB is not nil to avoid runtime panics
	if db.DB == nil {
		return errors.New("database connection is not initialized")
	}

	result := db.DB.Delete(workflowRun)
	err := result.Error
	if err != nil {
		return err
	}

	return nil
}

func WorkflowRunDeleteByWorkflowID(workflowID string) error {
	// Ensure db.DB is not nil to avoid runtime panics
	if db.DB == nil {
		return errors.New("database connection is not initialized")
	}

	result := db.DB.Where("workflow_id = ?", workflowID).Delete(&domain.WorkflowRun{})
	err := result.Error
	if err != nil {
		return err
	}

	return nil
}

func ExistsWorkflowRun(id string) bool {
	workflowRun, err := WorkflowRunGet(id)
	if err != nil {
		return false
	}
	return workflowRun != nil
}
