package dao

import (
	"errors"
	"github.com/cloud-barista/cm-cicada/db"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
	"gorm.io/gorm"
)

func WorkflowTemplateGet(id string) (*model.WorkflowTemplate, error) {
	workflowTemplate := &model.WorkflowTemplate{}

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

	return workflowTemplate, nil
}

func WorkflowTemplateGetList(page int, row int) (*[]model.WorkflowTemplate, error) {
	workflowTemplateList := &[]model.WorkflowTemplate{}
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

			return nil
		} else if page != 0 && row == 0 {
			filtered.Error = errors.New("page is not 0 but row is 0")

			return nil
		}
		return filtered
	}).Find(workflowTemplateList)

	err := result.Error
	if err != nil {
		return nil, err
	}

	return workflowTemplateList, nil
}

func WorkflowTemplateCreate(workflowTemplate *model.WorkflowTemplate) (*model.WorkflowTemplate, error) {
	// UUID, err := uuid.NewRandom()
	// if err != nil {
	// 	return nil, err
	// }

	// migrationGroup.UUID = UUID.String()

	result := db.DB.Create(workflowTemplate)
	err := result.Error
	if err != nil {
		return nil, err
	}

	return workflowTemplate, nil
}