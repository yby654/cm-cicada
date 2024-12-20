package dao

import (
	"errors"
	"fmt"
	"time"

	"github.com/cloud-barista/cm-cicada/db"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
	"gorm.io/gorm"
)

func WorkflowCreate(workflow *model.Workflow) (*model.Workflow, error) {
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

func WorkflowGet(id string) (*model.Workflow, error) {
	workflow := &model.Workflow{}

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

func WorkflowGetByName(name string) (*model.Workflow, error) {
	workflow := &model.Workflow{}

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

func WorkflowGetList(workflow *model.Workflow, page int, row int) (*[]model.Workflow, error) {
	WorkflowList := &[]model.Workflow{}
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

func WorkflowUpdate(workflow *model.Workflow) error {
	workflow.UpdatedAt = time.Now()

	result := db.DB.Model(&model.Workflow{}).Where("id = ?", workflow.ID).Updates(workflow)
	err := result.Error
	if err != nil {
		return err
	}

	return nil
}

func WorkflowDelete(workflow *model.Workflow) error {
	result := db.DB.Delete(workflow)
	err := result.Error
	if err != nil {
		return err
	}

	return nil
}

func WorkflowVersionGetList(workflowVersion *model.WorkflowVersion, page int, row int) (*[]model.WorkflowVersion, error) {
	WorkflowVersionList := &[]model.WorkflowVersion{}
	// Ensure db.DB is not nil to avoid runtime panics
	if db.DB == nil {
		return nil, errors.New("database connection is not initialized")
	}

	result := db.DB.Scopes(func(d *gorm.DB) *gorm.DB {
		var filtered = d

		if len(workflowVersion.WorkflowID) != 0 {
			filtered = filtered.Where("workflowId LIKE ?", "%"+workflowVersion.WorkflowID+"%")
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
	}).Find(WorkflowVersionList)

	err := result.Error
	if err != nil {
		return nil, err
	}

	return WorkflowVersionList, nil
}

func WorkflowVersionGet(id string, wkId string) (*model.WorkflowVersion, error) {
	workflowVersion := &model.WorkflowVersion{}

	// Ensure db.DB is not nil to avoid runtime panics
	if db.DB == nil {
		return nil, errors.New("database connection is not initialized")
	}

	result := db.DB.Where("id = ? and workflowId = ?", id, wkId).First(workflowVersion)
	fmt.Println("result : ", result)
	err := result.Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("workflow not found with the provided id")
		}
		return nil, err
	}

	return workflowVersion, nil
}
