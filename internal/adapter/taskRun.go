package adapter

import (
	"errors"

	"github.com/cloud-barista/cm-cicada/internal/db"
	"github.com/cloud-barista/cm-cicada/internal/domain"

	"gorm.io/gorm"
)

func TaskRunCreate(taskRun *domain.TaskRun) (*domain.TaskRun, error) {
	// Ensure db.DB is not nil to avoid runtime panics
	if db.DB == nil {
		return nil, errors.New("database connection is not initialized")
	}

	result := db.DB.Create(taskRun)
	err := result.Error
	if err != nil {
		return nil, err
	}

	return taskRun, nil
}

func TaskRunGet(id string) (*domain.TaskRun, error) {
	taskRun := &domain.TaskRun{}

	// Ensure db.DB is not nil to avoid runtime panics
	if db.DB == nil {
		return nil, errors.New("database connection is not initialized")
	}

	result := db.DB.Where("id = ?", id).First(taskRun)
	err := result.Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("task run not found with the provided id")
		}
		return nil, err
	}

	return taskRun, nil
}

func TaskRunGetByWorkflowRunID(workflowRunID string) (*[]domain.TaskRun, error) {
	taskRunList := &[]domain.TaskRun{}

	// Ensure db.DB is not nil to avoid runtime panics
	if db.DB == nil {
		return nil, errors.New("database connection is not initialized")
	}

	result := db.DB.Where("workflow_run_id = ?", workflowRunID).Find(taskRunList)
	err := result.Error
	if err != nil {
		return nil, err
	}

	return taskRunList, nil
}

func TaskRunGetByTaskID(taskID string) (*[]domain.TaskRun, error) {
	taskRunList := &[]domain.TaskRun{}

	// Ensure db.DB is not nil to avoid runtime panics
	if db.DB == nil {
		return nil, errors.New("database connection is not initialized")
	}

	result := db.DB.Where("task_id = ?", taskID).Find(taskRunList)
	err := result.Error
	if err != nil {
		return nil, err
	}

	return taskRunList, nil
}

func TaskRunGetList(taskRun *domain.TaskRun, page int, row int) (*[]domain.TaskRun, error) {
	taskRunList := &[]domain.TaskRun{}
	// Ensure db.DB is not nil to avoid runtime panics
	if db.DB == nil {
		return nil, errors.New("database connection is not initialized")
	}

	result := db.DB.Scopes(func(d *gorm.DB) *gorm.DB {
		var filtered = d

		if len(taskRun.WorkflowRunID) != 0 {
			filtered = filtered.Where("workflow_run_id = ?", taskRun.WorkflowRunID)
		}

		if len(taskRun.TaskID) != 0 {
			filtered = filtered.Where("task_id = ?", taskRun.TaskID)
		}

		if len(taskRun.State) != 0 {
			filtered = filtered.Where("state = ?", taskRun.State)
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
	}).Find(taskRunList)

	err := result.Error
	if err != nil {
		return nil, err
	}

	return taskRunList, nil
}

func TaskRunUpdate(taskRun *domain.TaskRun) error {
	// Ensure db.DB is not nil to avoid runtime panics
	if db.DB == nil {
		return errors.New("database connection is not initialized")
	}

	result := db.DB.Model(&domain.TaskRun{}).Where("id = ?", taskRun.ID).Updates(taskRun)
	err := result.Error
	if err != nil {
		return err
	}

	return nil
}

func TaskRunDelete(taskRun *domain.TaskRun) error {
	// Ensure db.DB is not nil to avoid runtime panics
	if db.DB == nil {
		return errors.New("database connection is not initialized")
	}

	result := db.DB.Delete(taskRun)
	err := result.Error
	if err != nil {
		return err
	}

	return nil
}

func TaskRunDeleteByWorkflowRunID(workflowRunID string) error {
	// Ensure db.DB is not nil to avoid runtime panics
	if db.DB == nil {
		return errors.New("database connection is not initialized")
	}

	result := db.DB.Where("workflow_run_id = ?", workflowRunID).Delete(&domain.TaskRun{})
	err := result.Error
	if err != nil {
		return err
	}

	return nil
}
