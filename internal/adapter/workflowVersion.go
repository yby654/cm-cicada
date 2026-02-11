package adapter

import (
	"errors"

	"github.com/cloud-barista/cm-cicada/internal/db"
	"github.com/cloud-barista/cm-cicada/internal/domain"

	"gorm.io/gorm"
)

func WorkflowVersionCreate(workflowVersion *domain.WorkflowVersion) (*domain.WorkflowVersion, error) {
	// Ensure db.DB is not nil to avoid runtime panics
	if db.DB == nil {
		return nil, errors.New("database connection is not initialized")
	}

	result := db.DB.Create(workflowVersion)
	err := result.Error
	if err != nil {
		return nil, err
	}

	return workflowVersion, nil
}

func WorkflowVersionGet(id string) (*domain.WorkflowVersion, error) {
	workflowVersion := &domain.WorkflowVersion{}

	// Ensure db.DB is not nil to avoid runtime panics
	if db.DB == nil {
		return nil, errors.New("database connection is not initialized")
	}

	result := db.DB.Where("id = ?", id).First(workflowVersion)
	err := result.Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("workflow version not found with the provided id")
		}
		return nil, err
	}

	return workflowVersion, nil
}

func WorkflowVersionGetByWorkflowIDAndVersion(workflowID string, version int) (*domain.WorkflowVersion, error) {
	workflowVersion := &domain.WorkflowVersion{}

	// Ensure db.DB is not nil to avoid runtime panics
	if db.DB == nil {
		return nil, errors.New("database connection is not initialized")
	}

	result := db.DB.Where("workflow_id = ? AND version = ?", workflowID, version).First(workflowVersion)
	err := result.Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("workflow version not found with the provided workflow_id and version")
		}
		return nil, err
	}

	return workflowVersion, nil
}

func WorkflowVersionGetList(workflowVersion *domain.WorkflowVersion, page int, row int) (*[]domain.WorkflowVersion, error) {
	WorkflowVersionList := &[]domain.WorkflowVersion{}
	// Ensure db.DB is not nil to avoid runtime panics
	if db.DB == nil {
		return nil, errors.New("database connection is not initialized")
	}

	result := db.DB.Scopes(func(d *gorm.DB) *gorm.DB {
		var filtered = d

		if len(workflowVersion.WorkflowID) != 0 {
			filtered = filtered.Where("workflow_id LIKE ?", "%"+workflowVersion.WorkflowID+"%")
		}

		if workflowVersion.Version != 0 {
			filtered = filtered.Where("version = ?", workflowVersion.Version)
		}

		if len(workflowVersion.SpecVersion) != 0 {
			filtered = filtered.Where("spec_version = ?", workflowVersion.SpecVersion)
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

func WorkflowVersionUpdate(workflowVersion *domain.WorkflowVersion) error {
	// Ensure db.DB is not nil to avoid runtime panics
	if db.DB == nil {
		return errors.New("database connection is not initialized")
	}

	result := db.DB.Model(&domain.WorkflowVersion{}).Where("id = ?", workflowVersion.ID).Updates(workflowVersion)
	err := result.Error
	if err != nil {
		return err
	}

	return nil
}

func WorkflowVersionDelete(workflowVersion *domain.WorkflowVersion) error {
	// Ensure db.DB is not nil to avoid runtime panics
	if db.DB == nil {
		return errors.New("database connection is not initialized")
	}

	result := db.DB.Delete(workflowVersion)
	err := result.Error
	if err != nil {
		return err
	}

	return nil
}

func WorkflowVersionDeleteByWorkflowID(workflowID string) error {
	// Ensure db.DB is not nil to avoid runtime panics
	if db.DB == nil {
		return errors.New("database connection is not initialized")
	}

	result := db.DB.Where("workflow_id = ?", workflowID).Delete(&domain.WorkflowVersion{})
	err := result.Error
	if err != nil {
		return err
	}

	return nil
}
