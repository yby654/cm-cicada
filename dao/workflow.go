package dao

import (
	"errors"
	"time"

	"github.com/cloud-barista/cm-cicada/db"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func ensureDB() error {
	if db.DB == nil {
		return errors.New("database connection is not initialized")
	}
	return nil
}

func ensureWorkflowKey(workflow *model.Workflow) error {
	if workflow == nil {
		return errors.New("workflow is nil")
	}

	if workflow.WorkflowKey != "" {
		return nil
	}

	workflow.WorkflowKey = workflow.ID
	return db.DB.Model(&model.Workflow{}).
		Where("id = ?", workflow.ID).
		Update("workflow_key", workflow.WorkflowKey).Error
}

func WorkflowCreate(workflow *model.Workflow) (*model.Workflow, error) {
	return WorkflowCreateTx(db.DB, workflow)
}

// WorkflowCreateTx is the transaction-aware variant of WorkflowCreate. Pass a
// *gorm.DB obtained from db.DB.Transaction to participate in that transaction;
// WorkflowCreate delegates here with the global handle for non-tx callers.
func WorkflowCreateTx(tx *gorm.DB, workflow *model.Workflow) (*model.Workflow, error) {
	if tx == nil {
		return nil, errors.New("database connection is not initialized")
	}

	now := time.Now()
	if workflow.WorkflowKey == "" {
		workflow.WorkflowKey = workflow.ID
	}

	workflow.IsDeleted = false
	workflow.DeletedAt = nil
	workflow.CreatedAt = now
	workflow.UpdatedAt = now

	result := tx.Create(workflow)
	if result.Error != nil {
		return nil, result.Error
	}

	return workflow, nil
}

func WorkflowGet(id string) (*model.Workflow, error) {
	if err := ensureDB(); err != nil {
		return nil, err
	}

	workflow := &model.Workflow{}
	result := db.DB.Where("id = ? AND is_deleted = ?", id, false).First(workflow)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("workflow not found with the provided id")
		}
		return nil, result.Error
	}

	if err := ensureWorkflowKey(workflow); err != nil {
		return nil, err
	}

	return workflow, nil
}

func WorkflowGetIncludeDeleted(id string) (*model.Workflow, error) {
	if err := ensureDB(); err != nil {
		return nil, err
	}

	workflow := &model.Workflow{}
	result := db.DB.Where("id = ?", id).First(workflow)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("workflow not found with the provided id")
		}
		return nil, result.Error
	}

	if err := ensureWorkflowKey(workflow); err != nil {
		return nil, err
	}

	return workflow, nil
}

func WorkflowGetByName(name string) (*model.Workflow, error) {
	if err := ensureDB(); err != nil {
		return nil, err
	}

	workflow := &model.Workflow{}
	result := db.DB.Where("name = ? AND is_deleted = ?", name, false).First(workflow)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("workflow not found with the provided name")
		}
		return nil, result.Error
	}

	if err := ensureWorkflowKey(workflow); err != nil {
		return nil, err
	}

	return workflow, nil
}

func WorkflowGetByNameIncludeDeleted(name string) (*model.Workflow, error) {
	if err := ensureDB(); err != nil {
		return nil, err
	}

	workflow := &model.Workflow{}
	result := db.DB.Where("name = ?", name).First(workflow)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("workflow not found with the provided name")
		}
		return nil, result.Error
	}

	if err := ensureWorkflowKey(workflow); err != nil {
		return nil, err
	}

	return workflow, nil
}

func WorkflowGetList(workflow *model.Workflow, page int, row int) (*[]model.Workflow, error) {
	if err := ensureDB(); err != nil {
		return nil, err
	}

	workflowList := &[]model.Workflow{}
	result := db.DB.Scopes(func(d *gorm.DB) *gorm.DB {
		filtered := d.Where("is_deleted = ?", false)

		if len(workflow.Name) != 0 {
			filtered = filtered.Where("name LIKE ?", "%"+workflow.Name+"%")
		}

		if page != 0 && row != 0 {
			offset := (page - 1) * row
			return filtered.Offset(offset).Limit(row)
		}
		if row != 0 && page == 0 {
			filtered.Error = errors.New("row is not 0 but page is 0")
			return filtered
		}
		if page != 0 && row == 0 {
			filtered.Error = errors.New("page is not 0 but row is 0")
			return filtered
		}

		return filtered
	}).Find(workflowList)
	if result.Error != nil {
		return nil, result.Error
	}

	for i := range *workflowList {
		if err := ensureWorkflowKey(&(*workflowList)[i]); err != nil {
			return nil, err
		}
	}

	return workflowList, nil
}

func WorkflowGetListIncludeDeleted(workflow *model.Workflow, page int, row int) (*[]model.Workflow, error) {
	if err := ensureDB(); err != nil {
		return nil, err
	}

	workflowList := &[]model.Workflow{}
	result := db.DB.Scopes(func(d *gorm.DB) *gorm.DB {
		filtered := d

		if len(workflow.Name) != 0 {
			filtered = filtered.Where("name LIKE ?", "%"+workflow.Name+"%")
		}

		if page != 0 && row != 0 {
			offset := (page - 1) * row
			return filtered.Offset(offset).Limit(row)
		}
		if row != 0 && page == 0 {
			filtered.Error = errors.New("row is not 0 but page is 0")
			return filtered
		}
		if page != 0 && row == 0 {
			filtered.Error = errors.New("page is not 0 but row is 0")
			return filtered
		}

		return filtered
	}).Find(workflowList)
	if result.Error != nil {
		return nil, result.Error
	}

	for i := range *workflowList {
		if err := ensureWorkflowKey(&(*workflowList)[i]); err != nil {
			return nil, err
		}
	}

	return workflowList, nil
}

func WorkflowUpdate(workflow *model.Workflow) error {
	return WorkflowUpdateTx(db.DB, workflow)
}

// WorkflowUpdateTx is the transaction-aware variant of WorkflowUpdate.
func WorkflowUpdateTx(tx *gorm.DB, workflow *model.Workflow) error {
	if tx == nil {
		return errors.New("database connection is not initialized")
	}

	if workflow.WorkflowKey == "" {
		workflow.WorkflowKey = workflow.ID
	}

	workflow.UpdatedAt = time.Now()
	result := tx.Model(&model.Workflow{}).
		Where("id = ? AND is_deleted = ?", workflow.ID, false).
		Updates(workflow)
	return result.Error
}

func WorkflowDelete(workflow *model.Workflow) error {
	if err := ensureDB(); err != nil {
		return err
	}

	now := time.Now()
	result := db.DB.Model(&model.Workflow{}).
		Where("id = ? AND is_deleted = ?", workflow.ID, false).
		Updates(map[string]interface{}{
			"is_deleted": true,
			"deleted_at": &now,
			"updated_at": now,
		})

	return result.Error
}

func WorkflowSetCurrentVersion(workflowID string, versionID string) error {
	return WorkflowSetCurrentVersionTx(db.DB, workflowID, versionID)
}

// WorkflowSetCurrentVersionTx is the transaction-aware variant of
// WorkflowSetCurrentVersion.
func WorkflowSetCurrentVersionTx(tx *gorm.DB, workflowID string, versionID string) error {
	if tx == nil {
		return errors.New("database connection is not initialized")
	}

	return tx.Model(&model.Workflow{}).
		Where("id = ?", workflowID).
		Update("current_version_id", versionID).Error
}

func WorkflowVersionGetLatestVersionNo(workflowID string) (int, error) {
	return WorkflowVersionGetLatestVersionNoTx(db.DB, workflowID)
}

// WorkflowVersionGetLatestVersionNoTx is the transaction-aware variant; reading
// the latest version within the same tx keeps version_no monotonic under
// concurrent snapshot creation.
func WorkflowVersionGetLatestVersionNoTx(tx *gorm.DB, workflowID string) (int, error) {
	if tx == nil {
		return 0, errors.New("database connection is not initialized")
	}

	latest := &model.WorkflowVersion{}
	result := tx.Where("workflowId = ?", workflowID).
		Order("version_no DESC").
		First(latest)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return 0, nil
		}
		return 0, result.Error
	}

	return latest.VersionNo, nil
}

func WorkflowCreateSnapshot(workflow *model.Workflow, action string, sourceType string, sourceTemplateID string) (*model.WorkflowVersion, error) {
	return WorkflowCreateSnapshotTx(db.DB, workflow, action, sourceType, sourceTemplateID)
}

// WorkflowCreateSnapshotTx is the transaction-aware variant of
// WorkflowCreateSnapshot. The version-number read and the current-version
// update run on the same tx so the snapshot + pointer move atomically.
func WorkflowCreateSnapshotTx(tx *gorm.DB, workflow *model.Workflow, action string, sourceType string, sourceTemplateID string) (*model.WorkflowVersion, error) {
	if tx == nil {
		return nil, errors.New("database connection is not initialized")
	}

	latestVersionNo, err := WorkflowVersionGetLatestVersionNoTx(tx, workflow.ID)
	if err != nil {
		return nil, err
	}

	workflowVersion := &model.WorkflowVersion{
		ID:               uuid.New().String(),
		SpecVersion:      workflow.SpecVersion,
		WorkflowID:       workflow.ID,
		VersionNo:        latestVersionNo + 1,
		RawData:          *workflow,
		Action:           action,
		SourceType:       sourceType,
		SourceTemplateID: sourceTemplateID,
		CreatedAt:        time.Now(),
	}

	if err := tx.Create(workflowVersion).Error; err != nil {
		return nil, err
	}

	if err := WorkflowSetCurrentVersionTx(tx, workflow.ID, workflowVersion.ID); err != nil {
		return nil, err
	}
	workflow.CurrentVersionID = workflowVersion.ID

	return workflowVersion, nil
}

func WorkflowVersionGetList(workflowVersion *model.WorkflowVersion, page int, row int) (*[]model.WorkflowVersion, error) {
	if err := ensureDB(); err != nil {
		return nil, err
	}

	workflowVersionList := &[]model.WorkflowVersion{}
	result := db.DB.Scopes(func(d *gorm.DB) *gorm.DB {
		filtered := d

		if len(workflowVersion.WorkflowID) != 0 {
			filtered = filtered.Where("workflowId = ?", workflowVersion.WorkflowID)
		}

		if page != 0 && row != 0 {
			offset := (page - 1) * row
			return filtered.Offset(offset).Limit(row)
		}
		if row != 0 && page == 0 {
			filtered.Error = errors.New("row is not 0 but page is 0")
			return filtered
		}
		if page != 0 && row == 0 {
			filtered.Error = errors.New("page is not 0 but row is 0")
			return filtered
		}
		return filtered
	}).Order("version_no DESC, created_at DESC").Find(workflowVersionList)
	if result.Error != nil {
		return nil, result.Error
	}

	return workflowVersionList, nil
}

func WorkflowVersionGet(id string, wkID string) (*model.WorkflowVersion, error) {
	if err := ensureDB(); err != nil {
		return nil, err
	}

	workflowVersion := &model.WorkflowVersion{}
	result := db.DB.Where("id = ? AND workflowId = ?", id, wkID).First(workflowVersion)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("workflow not found with the provided id")
		}
		return nil, result.Error
	}

	return workflowVersion, nil
}

// WorkflowVersionGetByNo returns (nil, nil) when no row matches.
func WorkflowVersionGetByNo(workflowID string, versionNo int) (*model.WorkflowVersion, error) {
	if err := ensureDB(); err != nil {
		return nil, err
	}

	workflowVersion := &model.WorkflowVersion{}
	result := db.DB.Where("workflowId = ? AND version_no = ?", workflowID, versionNo).First(workflowVersion)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}

	return workflowVersion, nil
}
