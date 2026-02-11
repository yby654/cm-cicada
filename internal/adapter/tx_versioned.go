package adapter

import (
	"errors"
	"time"

	"github.com/cloud-barista/cm-cicada/internal/domain"
	"gorm.io/gorm"
)

// TaskGroupMarkDeletedTx sets deleted_at for the given TaskGroup (soft delete) within the transaction.
func TaskGroupMarkDeletedTx(tx *gorm.DB, id string) error {
	if tx == nil {
		return errors.New("transaction is nil")
	}
	now := time.Now()
	return tx.Model(&domain.TaskGroup{}).Where("id = ?", id).Update("deleted_at", now).Error
}

// TaskMarkDeletedTx sets deleted_at for the given Task (soft delete) within the transaction.
func TaskMarkDeletedTx(tx *gorm.DB, id string) error {
	if tx == nil {
		return errors.New("transaction is nil")
	}
	now := time.Now()
	return tx.Model(&domain.Task{}).Where("id = ?", id).Update("deleted_at", now).Error
}

// TaskGroupUpsertTx creates or updates a TaskGroup within the transaction.
// If a row exists (including soft-deleted), it is updated and DeletedAt is cleared.
func TaskGroupUpsertTx(tx *gorm.DB, tg *domain.TaskGroup) (*domain.TaskGroup, error) {
	if tx == nil {
		return nil, errors.New("transaction is nil")
	}
	var existing domain.TaskGroup
	err := tx.Unscoped().Where("id = ?", tg.ID).First(&existing).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			tg.DeletedAt = nil
			if err := tx.Create(tg).Error; err != nil {
				return nil, err
			}
			return tg, nil
		}
		return nil, err
	}
	// 존재하면 갱신, soft delete 복구
	up := map[string]interface{}{
		"name":                tg.Name,
		"description":         tg.Description,
		"workflow_version_id": tg.WorkflowVersionID,
		"deleted_at":          nil,
	}
	if err := tx.Model(&domain.TaskGroup{}).Where("id = ?", tg.ID).Updates(up).Error; err != nil {
		return nil, err
	}
	tg.DeletedAt = nil
	tg.CreatedAt = existing.CreatedAt
	return tg, nil
}

// TaskUpsertTx creates or updates a Task within the transaction.
// If a row exists (including soft-deleted), it is updated and DeletedAt is cleared.
func TaskUpsertTx(tx *gorm.DB, t *domain.Task) (*domain.Task, error) {
	if tx == nil {
		return nil, errors.New("transaction is nil")
	}
	var existing domain.Task
	err := tx.Unscoped().Where("id = ?", t.ID).First(&existing).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			t.DeletedAt = nil
			if err := tx.Create(t).Error; err != nil {
				return nil, err
			}
			return t, nil
		}
		return nil, err
	}
	up := map[string]interface{}{
		"name":            t.Name,
		"workflow_id":     t.WorkflowID,
		"task_group_id":   t.TaskGroupID,
		"task_component": t.TaskComponent,
		"request_body":    t.RequestBody,
		"path_params":     t.PathParams,
		"query_params":    t.QueryParams,
		"extra":           t.Extra,
		"dependencies":    t.Dependencies,
		"deleted_at":      nil,
	}
	if err := tx.Model(&domain.Task{}).Where("id = ?", t.ID).Updates(up).Error; err != nil {
		return nil, err
	}
	t.DeletedAt = nil
	t.CreatedAt = existing.CreatedAt
	return t, nil
}
