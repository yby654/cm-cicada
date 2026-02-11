package adapter

import (
	"errors"
	"time"

	"github.com/cloud-barista/cm-cicada/internal/db"
	"github.com/cloud-barista/cm-cicada/internal/domain"

	"gorm.io/gorm"
)

func WorkflowCreate(workflow *domain.Workflow) (*domain.Workflow, error) {
	now := time.Now()

	workflow.CreatedAt = now
	workflow.UpdatedAt = now

	result := db.DB.Create(workflow)
	err := result.Error
	if err != nil {
		return nil, err
	}

	return workflow, nil
}

func WorkflowGet(id string) (*domain.Workflow, error) {
	workflow := &domain.Workflow{}

	// Ensure db.DB is not nil to avoid runtime panics
	if db.DB == nil {
		return nil, errors.New("database connection is not initialized")
	}

	result := db.DB.Where("id = ?", id).First(workflow)
	err := result.Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("workflow not found with the provided id")
		}
		return nil, err
	}

	return workflow, nil
}

func WorkflowGetByName(name string) (*domain.Workflow, error) {
	workflow := &domain.Workflow{}

	// Ensure db.DB is not nil to avoid runtime panics
	if db.DB == nil {
		return nil, errors.New("database connection is not initialized")
	}

	result := db.DB.Where("name = ?", name).First(workflow)
	err := result.Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("workflow not found with the provided name")
		}
		return nil, err
	}

	return workflow, nil
}

func WorkflowGetList(workflow *domain.Workflow, page int, row int) (*[]domain.Workflow, error) {
	WorkflowList := &[]domain.Workflow{}
	// Ensure db.DB is not nil to avoid runtime panics
	if db.DB == nil {
		return nil, errors.New("database connection is not initialized")
	}

	result := db.DB.Scopes(func(d *gorm.DB) *gorm.DB {
		var filtered = d

		if len(workflow.Name) != 0 {
			filtered = filtered.Where("name LIKE ?", "%"+workflow.Name+"%")
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
	}).Find(WorkflowList)

	err := result.Error
	if err != nil {
		return nil, err
	}

	return WorkflowList, nil
}

func WorkflowUpdate(workflow *domain.Workflow) error {
	workflow.UpdatedAt = time.Now()

	result := db.DB.Model(&domain.Workflow{}).Where("id = ?", workflow.ID).Updates(workflow)
	err := result.Error
	if err != nil {
		return err
	}

	return nil
}

func WorkflowDelete(workflow *domain.Workflow) error {
	result := db.DB.Delete(workflow)
	err := result.Error
	if err != nil {
		return err
	}

	return nil
}

func ExistsWorkflow(id string) bool {
	workflow, err := WorkflowGet(id)
	if err != nil {
		return false
	}
	return workflow != nil
}
