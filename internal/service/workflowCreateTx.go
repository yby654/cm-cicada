package service

import (
	"errors"

	"github.com/cloud-barista/cm-cicada/internal/adapter"
	"github.com/cloud-barista/cm-cicada/internal/db"
	"github.com/cloud-barista/cm-cicada/internal/domain"
)

// CreateWorkflowGraphTx creates Workflow + TaskGroups + Tasks in a single DB transaction.
// Note: This does NOT include external side-effects (e.g. Airflow DAG creation).
func CreateWorkflowGraphTx(workflow *domain.Workflow) error {
	if workflow == nil {
		return errors.New("workflow is nil")
	}
	if workflow.ID == "" {
		return errors.New("workflow id is empty")
	}

	tx, err := db.BeginTransaction()
	if err != nil {
		return err
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if _, err := adapter.WorkflowCreateTx(tx, workflow); err != nil {
		tx.Rollback()
		return err
	}

	for _, tg := range workflow.Data.TaskGroups {
		tgRow := domain.TaskGroup{
			ID:                tg.ID,
			Name:              tg.Name,
			Description:       tg.Description,
			WorkflowVersionID: workflow.ID, // current design: stores Workflow.ID
		}

		if _, err := adapter.TaskGroupCreateTx(tx, &tgRow); err != nil {
			tx.Rollback()
			return err
		}

		for _, t := range tg.Tasks {
			taskRow := domain.Task{
				ID:            t.ID,
				Name:          t.Name,
				WorkflowID:    workflow.ID,
				TaskGroupID:   tg.ID,
				TaskComponent: t.TaskComponent,
				RequestBody:   t.RequestBody,
				PathParams:    t.PathParams,
				QueryParams:   t.QueryParams,
				Extra:         t.Extra,
				Dependencies:  t.Dependencies,
			}

			if _, err := adapter.TaskCreateTx(tx, &taskRow); err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	return tx.Commit().Error
}

// DeleteWorkflowGraph deletes Tasks + TaskGroups + Workflow in a single transaction.
// This is intended for compensating cleanup when an external side-effect fails after commit.
func DeleteWorkflowGraph(workflowID string) error {
	if workflowID == "" {
		return errors.New("workflow id is empty")
	}

	tx, err := db.BeginTransaction()
	if err != nil {
		return err
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// tasks (FK: workflow_id)
	if err := tx.Where("workflow_id = ?", workflowID).Delete(&domain.Task{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// task_groups (current design: workflow_version_id stores Workflow.ID)
	if err := tx.Where("workflow_version_id = ?", workflowID).Delete(&domain.TaskGroup{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// workflow
	if err := tx.Where("id = ?", workflowID).Delete(&domain.Workflow{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}
