package adapter

import (
	"errors"

	"github.com/cloud-barista/cm-cicada/internal/db"
	"github.com/cloud-barista/cm-cicada/internal/domain"
	"gorm.io/gorm"
)

func TaskGroupCreate(taskGroup *domain.TaskGroup) (*domain.TaskGroup, error) {
	// Ensure db.DB is not nil to avoid runtime panics
	if db.DB == nil {
		return nil, errors.New("database connection is not initialized")
	}

	result := db.DB.Create(taskGroup)
	err := result.Error
	if err != nil {
		return nil, err
	}

	return taskGroup, nil
}

func TaskGroupGet(id string) (*domain.TaskGroup, error) {
	taskGroup := &domain.TaskGroup{}

	// Ensure db.DB is not nil to avoid runtime panics
	if db.DB == nil {
		return nil, errors.New("database connection is not initialized")
	}

	result := db.DB.Where("id = ?", id).First(taskGroup)
	err := result.Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("task_group not found with the provided id")
		}
		return nil, err
	}

	return taskGroup, nil
}

func TaskGroupGetByName(name string) (*domain.TaskGroup, error) {
	taskGroup := &domain.TaskGroup{}

	// Ensure db.DB is not nil to avoid runtime panics
	if db.DB == nil {
		return nil, errors.New("database connection is not initialized")
	}

	result := db.DB.Where("name = ?", name).First(taskGroup)
	err := result.Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("task_group not found with the provided name")
		}
		return nil, err
	}

	return taskGroup, nil
}

func TaskGroupGetList(taskGroup *domain.TaskGroup, page int, row int) (*[]domain.TaskGroup, error) {
	taskGroupList := &[]domain.TaskGroup{}
	// Ensure db.DB is not nil to avoid runtime panics
	if db.DB == nil {
		return nil, errors.New("database connection is not initialized")
	}

	result := db.DB.Scopes(func(d *gorm.DB) *gorm.DB {
		var filtered = d

		if len(taskGroup.Name) != 0 {
			filtered = filtered.Where("name LIKE ?", "%"+taskGroup.Name+"%")
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
	}).Find(taskGroupList)

	err := result.Error
	if err != nil {
		return nil, err
	}

	return taskGroupList, nil
}

func TaskGroupUpdate(taskGroup *domain.TaskGroup) error {
	// Ensure db.DB is not nil to avoid runtime panics
	if db.DB == nil {
		return errors.New("database connection is not initialized")
	}

	result := db.DB.Model(&domain.TaskGroup{}).Where("id = ?", taskGroup.ID).Updates(taskGroup)
	err := result.Error
	if err != nil {
		return err
	}

	return nil
}

func TaskGroupDelete(taskGroup *domain.TaskGroup) error {
	// Ensure db.DB is not nil to avoid runtime panics
	if db.DB == nil {
		return errors.New("database connection is not initialized")
	}

	result := db.DB.Delete(taskGroup)
	err := result.Error
	if err != nil {
		return err
	}

	return nil
}

// TaskGroupDBModel 관련 함수들 (하위 호환성을 위해 유지)
func TaskGroupCreateDBModel(taskGroup *domain.TaskGroup) (*domain.TaskGroup, error) {
	result := db.DB.Create(taskGroup)
	err := result.Error
	if err != nil {
		return nil, err
	}

	return taskGroup, nil
}

func TaskGroupGetDBModel(id string) (*domain.TaskGroup, error) {
	taskGroup := &domain.TaskGroup{}

	// Ensure db.DB is not nil to avoid runtime panics
	if db.DB == nil {
		return nil, errors.New("database connection is not initialized")
	}

	result := db.DB.Where("id = ?", id).First(taskGroup)
	err := result.Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("task_group not found with the provided id")
		}
		return nil, err
	}

	return taskGroup, nil
}

func TaskGroupGetByWorkflowIDAndName(workflowID string, name string) (*domain.TaskGroup, error) {
	taskGroup := &domain.TaskGroup{}

	// Ensure db.DB is not nil to avoid runtime panics
	if db.DB == nil {
		return nil, errors.New("database connection is not initialized")
	}

	result := db.DB.Where("workflow_version_id = ? AND name = ?", workflowID, name).First(taskGroup)
	err := result.Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("task_group not found with the provided name")
		}
		return nil, err
	}

	return taskGroup, nil
}

func TaskGroupDeleteDBModel(taskGroup *domain.TaskGroup) error {
	result := db.DB.Delete(taskGroup)
	err := result.Error
	if err != nil {
		return err
	}

	return nil
}

func ExistsTaskGroup(id string) bool {
	taskGroup, err := TaskGroupGet(id)
	if err != nil {
		return false
	}
	return taskGroup != nil
}
