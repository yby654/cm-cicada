package adapter

import (
	"errors"
	"time"

	"github.com/cloud-barista/cm-cicada/internal/db"
	"github.com/cloud-barista/cm-cicada/internal/domain"
	"gorm.io/gorm"
)

func TaskComponentGet(id string) (*domain.TaskComponent, error) {
	taskComponent := &domain.TaskComponent{}

	// Ensure db.DB is not nil to avoid runtime panics
	if db.DB == nil {
		return nil, errors.New("database connection is not initialized")
	}

	result := db.DB.Where("id = ?", id).First(taskComponent)
	err := result.Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("task component not found with the provided id")
		}
		return nil, err
	}

	return taskComponent, nil
}

func TaskComponentGetByName(name string) (*domain.TaskComponent, error) {
	taskComponent := &domain.TaskComponent{}

	// Ensure db.DB is not nil to avoid runtime panics
	if db.DB == nil {
		return nil, errors.New("database connection is not initialized")
	}

	result := db.DB.Where("name = ?", name).First(taskComponent)
	err := result.Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("task component not found with the provided name")
		}
		return nil, err
	}

	return taskComponent, nil
}

func TaskComponentGetList(page int, row int) (*[]domain.TaskComponent, error) {
	taskComponentList := &[]domain.TaskComponent{}
	// Ensure db.DB is not nil to avoid runtime panics
	if db.DB == nil {
		return nil, errors.New("database connection is not initialized")
	}

	result := db.DB.Scopes(func(d *gorm.DB) *gorm.DB {
		var filtered = d

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
	}).Find(taskComponentList)

	err := result.Error
	if err != nil {
		return nil, err
	}

	return taskComponentList, nil
}

func TaskComponentCreate(taskComponent *domain.TaskComponent) (*domain.TaskComponent, error) {
	now := time.Now()

	taskComponent.CreatedAt = now
	taskComponent.UpdatedAt = now
	taskComponent.IsExample = false

	result := db.DB.Session(&gorm.Session{SkipHooks: true}).Create(taskComponent)
	err := result.Error
	if err != nil {
		return nil, err
	}

	return taskComponent, nil
}

func TaskComponentUpdate(taskComponent *domain.TaskComponent) error {
	if taskComponent.IsExample {
		return errors.New("example task component can't be updated")
	}

	taskComponent.UpdatedAt = time.Now()

	result := db.DB.Model(&domain.TaskComponent{}).Where("id = ?", taskComponent.ID).Updates(taskComponent)
	err := result.Error
	if err != nil {
		return err
	}

	return nil
}

func TaskComponentDelete(taskComponent *domain.TaskComponent) error {
	if taskComponent.IsExample {
		return errors.New("example task component can't be deleted")
	}

	result := db.DB.Delete(taskComponent)
	err := result.Error
	if err != nil {
		return err
	}

	return nil
}
