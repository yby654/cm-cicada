package adapter

import (
	"errors"

	"github.com/cloud-barista/cm-cicada/internal/db"
	"github.com/cloud-barista/cm-cicada/internal/domain"

	"gorm.io/gorm"
)

func WorkflowTemplateGet(id string) (*domain.GetWorkflowTemplate, error) {
	workflowTemplate := &domain.WorkflowTemplate{}

	// Ensure db.DB is not nil to avoid runtime panics
	if db.DB == nil {
		return nil, errors.New("database connection is not initialized")
	}

	result := db.DB.Where("id = ?", id).First(workflowTemplate)
	err := result.Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("workflow template not found with the provided id")
		}
		return nil, err
	}

	return &domain.GetWorkflowTemplate{
		SpecVersion: workflowTemplate.SpecVersion,
		Name:        workflowTemplate.Name,
		Data:        workflowTemplate.Data,
	}, nil
}

func WorkflowTemplateGetByName(name string) (*domain.WorkflowTemplate, error) {
	workflowTemplate := &domain.WorkflowTemplate{}

	// Ensure db.DB is not nil to avoid runtime panics
	if db.DB == nil {
		return nil, errors.New("database connection is not initialized")
	}

	result := db.DB.Where("name = ?", name).First(workflowTemplate)
	err := result.Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("workflow template not found with the provided name")
		}
		return nil, err
	}

	return workflowTemplate, nil
}

func WorkflowTemplateGetList(workflowTemplate *domain.WorkflowTemplate, page int, row int) (*[]domain.WorkflowTemplate, error) {
	workflowTemplateList := &[]domain.WorkflowTemplate{}
	// Ensure db.DB is not nil to avoid runtime panics
	if db.DB == nil {
		return nil, errors.New("database connection is not initialized")
	}

	result := db.DB.Scopes(func(d *gorm.DB) *gorm.DB {
		var filtered = d

		if len(workflowTemplate.Name) != 0 {
			filtered = filtered.Where("name LIKE ?", "%"+workflowTemplate.Name+"%")
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
	}).Find(workflowTemplateList)

	err := result.Error
	if err != nil {
		return nil, err
	}

	return workflowTemplateList, nil
}
