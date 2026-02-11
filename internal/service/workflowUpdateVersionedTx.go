package service

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/cloud-barista/cm-cicada/internal/adapter"
	"github.com/cloud-barista/cm-cicada/internal/db"
	"github.com/cloud-barista/cm-cicada/internal/domain"
	"github.com/google/uuid"
)

// UpdateWorkflowGraphVersionedTx implements "버전 생성" Update in a single DB transaction.
// Workflow = 현재 최신 정의 포인터, WorkflowVersion = 불변 스냅샷(append-only).
// 1) Tx 시작
// 2) 기존 Workflow 조회(호출자가 전달, 현재 버전 번호는 Tx 내에서 조회)
// 3) 새 WorkflowVersion 생성(version+1, definition_snapshot)
// 4) 삭제된 TG/Task는 soft delete, 나머지는 upsert
// 5) Workflow.data(캐시) 갱신, Workflow.updated_at 갱신
// 6) Tx 커밋
func UpdateWorkflowGraphVersionedTx(workflow *domain.Workflow, workflowData domain.Data, specVersion string) error {
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

	// 2) 현재 버전 번호 조회
	maxVer, err := adapter.WorkflowVersionMaxVersionTx(tx, workflow.ID)
	if err != nil {
		tx.Rollback()
		return err
	}

	// 3) 새 WorkflowVersion 생성 (불변 스냅샷, append-only)
	snapshot, err := json.Marshal(workflowData)
	if err != nil {
		tx.Rollback()
		return err
	}
	newVer := &domain.WorkflowVersion{
		ID:                 uuid.New().String(),
		WorkflowID:         workflow.ID,
		Version:            maxVer + 1,
		SpecVersion:        specVersion,
		DefinitionSnapshot: snapshot,
	}
	if _, err = adapter.WorkflowVersionCreateTx(tx, newVer); err != nil {
		tx.Rollback()
		return err
	}

	// 4) 삭제분: old에 있으나 new에 없는 TG/Task → soft delete
	newTGIds := make(map[string]struct{})
	for _, tg := range workflowData.TaskGroups {
		newTGIds[tg.ID] = struct{}{}
	}
	newTaskIdsByTG := make(map[string]map[string]struct{})
	for _, tg := range workflowData.TaskGroups {
		m := make(map[string]struct{})
		for _, t := range tg.Tasks {
			m[t.ID] = struct{}{}
		}
		newTaskIdsByTG[tg.ID] = m
	}

	for _, oldTG := range workflow.Data.TaskGroups {
		if _, inNew := newTGIds[oldTG.ID]; !inNew {
			// TG가 새 정의에 없음 → TG 및 하위 Task 전부 soft delete
			for _, t := range oldTG.Tasks {
				if err := adapter.TaskMarkDeletedTx(tx, t.ID); err != nil {
					tx.Rollback()
					return err
				}
			}
			if err := adapter.TaskGroupMarkDeletedTx(tx, oldTG.ID); err != nil {
				tx.Rollback()
				return err
			}
			continue
		}
		// TG는 유지, Task만 삭제된 것 soft delete
		newTaskIds := newTaskIdsByTG[oldTG.ID]
		for _, oldTask := range oldTG.Tasks {
			if _, inNew := newTaskIds[oldTask.ID]; !inNew {
				if err := adapter.TaskMarkDeletedTx(tx, oldTask.ID); err != nil {
					tx.Rollback()
					return err
				}
			}
		}
	}

	// 4) 새 정의 기준 TG/Task upsert (WorkflowVersionID = Workflow.ID 유지)
	for _, tg := range workflowData.TaskGroups {
		tgRow := &domain.TaskGroup{
			ID:                tg.ID,
			Name:              tg.Name,
			Description:       tg.Description,
			WorkflowVersionID: workflow.ID,
		}
		if _, err := adapter.TaskGroupUpsertTx(tx, tgRow); err != nil {
			tx.Rollback()
			return err
		}
		for _, t := range tg.Tasks {
			taskRow := &domain.Task{
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
			if _, err := adapter.TaskUpsertTx(tx, taskRow); err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	// 5) Workflow.data(캐시) 갱신, updated_at 갱신
	workflow.Data = workflowData
	workflow.UpdatedAt = time.Now()
	if err := adapter.WorkflowUpdateTx(tx, workflow); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}
