package dao

import (
	"errors"
	"time"

	"github.com/cloud-barista/cm-cicada/db"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
	"gorm.io/gorm"
)

func TaskGroupCreate(taskGroup *model.TaskGroupDBModel) (*model.TaskGroupDBModel, error) {
	return TaskGroupCreateTx(db.DB, taskGroup)
}

// TaskGroupCreateTx is the transaction-aware variant of TaskGroupCreate.
func TaskGroupCreateTx(tx *gorm.DB, taskGroup *model.TaskGroupDBModel) (*model.TaskGroupDBModel, error) {
	if tx == nil {
		return nil, errors.New("database connection is not initialized")
	}

	taskGroup.IsDeleted = false
	taskGroup.DeletedAt = nil
	if taskGroup.TaskGroupKey == "" {
		taskGroup.TaskGroupKey = taskGroup.ID
	}

	result := tx.Create(taskGroup)
	if result.Error != nil {
		return nil, result.Error
	}

	return taskGroup, nil
}

func TaskGroupSave(taskGroup *model.TaskGroupDBModel) error {
	return TaskGroupSaveTx(db.DB, taskGroup)
}

// TaskGroupSaveTx is the transaction-aware variant of TaskGroupSave.
func TaskGroupSaveTx(tx *gorm.DB, taskGroup *model.TaskGroupDBModel) error {
	if tx == nil {
		return errors.New("database connection is not initialized")
	}

	var existing model.TaskGroupDBModel
	result := tx.Unscoped().Where("id = ?", taskGroup.ID).First(&existing)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			_, err := TaskGroupCreateTx(tx, taskGroup)
			return err
		}
		return result.Error
	}

	taskGroup.IsDeleted = false
	taskGroup.DeletedAt = nil
	if taskGroup.TaskGroupKey == "" {
		taskGroup.TaskGroupKey = existing.TaskGroupKey
		if taskGroup.TaskGroupKey == "" {
			taskGroup.TaskGroupKey = taskGroup.ID
		}
	}

	return tx.Model(&model.TaskGroupDBModel{}).
		Where("id = ?", taskGroup.ID).
		Updates(taskGroup).Error
}

func TaskGroupGet(id string) (*model.TaskGroupDBModel, error) {
	if err := ensureDB(); err != nil {
		return nil, err
	}

	taskGroup := &model.TaskGroupDBModel{}
	result := db.DB.Where("id = ? AND is_deleted = ?", id, false).First(taskGroup)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("task_group not found with the provided id")
		}
		return nil, result.Error
	}

	if taskGroup.TaskGroupKey == "" {
		taskGroup.TaskGroupKey = taskGroup.ID
		_ = db.DB.Model(&model.TaskGroupDBModel{}).
			Where("id = ?", taskGroup.ID).
			Update("task_group_key", taskGroup.TaskGroupKey).Error
	}

	return taskGroup, nil
}

func TaskGroupGetIncludeDeleted(id string) (*model.TaskGroupDBModel, error) {
	if err := ensureDB(); err != nil {
		return nil, err
	}

	taskGroup := &model.TaskGroupDBModel{}
	result := db.DB.Where("id = ?", id).First(taskGroup)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("task_group not found with the provided id")
		}
		return nil, result.Error
	}

	if taskGroup.TaskGroupKey == "" {
		taskGroup.TaskGroupKey = taskGroup.ID
		_ = db.DB.Model(&model.TaskGroupDBModel{}).
			Where("id = ?", taskGroup.ID).
			Update("task_group_key", taskGroup.TaskGroupKey).Error
	}

	return taskGroup, nil
}

func TaskGroupGetByWorkflowIDAndName(workflowID string, name string) (*model.TaskGroupDBModel, error) {
	if err := ensureDB(); err != nil {
		return nil, err
	}

	taskGroup := &model.TaskGroupDBModel{}
	result := db.DB.Where("workflow_id = ? AND name = ? AND is_deleted = ?", workflowID, name, false).First(taskGroup)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("task_group not found with the provided name")
		}
		return nil, result.Error
	}

	if taskGroup.TaskGroupKey == "" {
		taskGroup.TaskGroupKey = taskGroup.ID
		_ = db.DB.Model(&model.TaskGroupDBModel{}).
			Where("id = ?", taskGroup.ID).
			Update("task_group_key", taskGroup.TaskGroupKey).Error
	}

	return taskGroup, nil
}

func TaskGroupGetByWorkflowIDAndNameIncludeDeleted(workflowID string, name string) (*model.TaskGroupDBModel, error) {
	if err := ensureDB(); err != nil {
		return nil, err
	}

	taskGroup := &model.TaskGroupDBModel{}
	result := db.DB.Where("workflow_id = ? AND name = ?", workflowID, name).First(taskGroup)
	if result.Error != nil {
		return nil, result.Error
	}

	if taskGroup.TaskGroupKey == "" {
		taskGroup.TaskGroupKey = taskGroup.ID
		_ = db.DB.Model(&model.TaskGroupDBModel{}).
			Where("id = ?", taskGroup.ID).
			Update("task_group_key", taskGroup.TaskGroupKey).Error
	}

	return taskGroup, nil
}

func TaskGroupGetListByWorkflowID(workflowID string, includeDeleted bool) ([]model.TaskGroupDBModel, error) {
	if err := ensureDB(); err != nil {
		return nil, err
	}

	taskGroups := []model.TaskGroupDBModel{}
	query := db.DB.Where("workflow_id = ?", workflowID)
	if !includeDeleted {
		query = query.Where("is_deleted = ?", false)
	}

	if err := query.Find(&taskGroups).Error; err != nil {
		return nil, err
	}

	for i := range taskGroups {
		if taskGroups[i].TaskGroupKey == "" {
			taskGroups[i].TaskGroupKey = taskGroups[i].ID
			_ = db.DB.Model(&model.TaskGroupDBModel{}).
				Where("id = ?", taskGroups[i].ID).
				Update("task_group_key", taskGroups[i].TaskGroupKey).Error
		}
	}

	return taskGroups, nil
}

func TaskGroupDelete(taskGroup *model.TaskGroupDBModel) error {
	return TaskGroupDeleteTx(db.DB, taskGroup)
}

// TaskGroupDeleteTx is the transaction-aware variant of TaskGroupDelete.
func TaskGroupDeleteTx(tx *gorm.DB, taskGroup *model.TaskGroupDBModel) error {
	if tx == nil {
		return errors.New("database connection is not initialized")
	}

	now := time.Now()
	return tx.Model(&model.TaskGroupDBModel{}).
		Where("id = ? AND is_deleted = ?", taskGroup.ID, false).
		Updates(map[string]interface{}{
			"is_deleted": true,
			"deleted_at": &now,
		}).Error
}

func TaskGroupSoftDeleteByWorkflowID(workflowID string) error {
	if err := ensureDB(); err != nil {
		return err
	}

	now := time.Now()
	return db.DB.Model(&model.TaskGroupDBModel{}).
		Where("workflow_id = ? AND is_deleted = ?", workflowID, false).
		Updates(map[string]interface{}{
			"is_deleted": true,
			"deleted_at": &now,
		}).Error
}
