package adapter

import (
	"errors"

	"github.com/cloud-barista/cm-cicada/internal/db"
	"github.com/cloud-barista/cm-cicada/internal/domain"
	"gorm.io/gorm"
)

func TaskCreate(task *domain.Task) (*domain.Task, error) {
	// Ensure db.DB is not nil to avoid runtime panics
	if db.DB == nil {
		return nil, errors.New("database connection is not initialized")
	}

	result := db.DB.Create(task)
	err := result.Error
	if err != nil {
		return nil, err
	}

	return task, nil
}

func TaskGet(id string) (*domain.Task, error) {
	task := &domain.Task{}

	// Ensure db.DB is not nil to avoid runtime panics
	if db.DB == nil {
		return nil, errors.New("database connection is not initialized")
	}

	result := db.DB.Where("id = ?", id).First(task)
	err := result.Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("task not found with the provided id")
		}
		return nil, err
	}

	return task, nil
}

func TaskGetByName(name string) (*domain.Task, error) {
	task := &domain.Task{}

	// Ensure db.DB is not nil to avoid runtime panics
	if db.DB == nil {
		return nil, errors.New("database connection is not initialized")
	}

	result := db.DB.Where("name = ?", name).First(task)
	err := result.Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("task not found with the provided name")
		}
		return nil, err
	}

	return task, nil
}

func TaskGetList(task *domain.Task, page int, row int) (*[]domain.Task, error) {
	taskList := &[]domain.Task{}
	// Ensure db.DB is not nil to avoid runtime panics
	if db.DB == nil {
		return nil, errors.New("database connection is not initialized")
	}

	result := db.DB.Scopes(func(d *gorm.DB) *gorm.DB {
		var filtered = d

		if len(task.Name) != 0 {
			filtered = filtered.Where("name LIKE ?", "%"+task.Name+"%")
		}

		if len(task.TaskComponent) != 0 {
			filtered = filtered.Where("task_component = ?", task.TaskComponent)
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
	}).Find(taskList)

	err := result.Error
	if err != nil {
		return nil, err
	}

	return taskList, nil
}

func TaskUpdate(task *domain.Task) error {
	// Ensure db.DB is not nil to avoid runtime panics
	if db.DB == nil {
		return errors.New("database connection is not initialized")
	}

	result := db.DB.Model(&domain.Task{}).Where("id = ?", task.ID).Updates(task)
	err := result.Error
	if err != nil {
		return err
	}

	return nil
}

func TaskDelete(task *domain.Task) error {
	// Ensure db.DB is not nil to avoid runtime panics
	if db.DB == nil {
		return errors.New("database connection is not initialized")
	}

	result := db.DB.Delete(task)
	err := result.Error
	if err != nil {
		return err
	}

	return nil
}

// TaskDBdomain 관련 함수들 (하위 호환성을 위해 유지)
func TaskCreateDBdomain(task *domain.Task) (*domain.Task, error) {
	result := db.DB.Create(task)
	err := result.Error
	if err != nil {
		return nil, err
	}

	return task, nil
}

func TaskGetDBdomain(id string) (*domain.Task, error) {
	task := &domain.Task{}

	// Ensure db.DB is not nil to avoid runtime panics
	if db.DB == nil {
		return nil, errors.New("database connection is not initialized")
	}

	result := db.DB.Where("id = ?", id).First(task)
	err := result.Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("task not found with the provided id")
		}
		return nil, err
	}

	return task, nil
}

func TaskGetByWorkflowIDAndName(workflowID string, name string) (*domain.Task, error) {
	task := &domain.Task{}

	// Ensure db.DB is not nil to avoid runtime panics
	if db.DB == nil {
		return nil, errors.New("database connection is not initialized")
	}

	result := db.DB.Where("workflow_id = ? and name = ?", workflowID, name).First(task)
	err := result.Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("task not found with the provided name")
		}
		return nil, err
	}

	return task, nil
}

func ExistsTask(id string) bool {
	task, err := TaskGet(id)
	if err != nil {
		return false
	}
	return task != nil
}
