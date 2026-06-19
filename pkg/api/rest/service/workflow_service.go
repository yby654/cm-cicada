package service

import (
	"errors"
	"strconv"

	"github.com/cloud-barista/cm-cicada/dao"
	"github.com/cloud-barista/cm-cicada/db"
	"github.com/cloud-barista/cm-cicada/lib/airflow"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/common"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/mapper"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
	"github.com/google/uuid"
	"github.com/jollaman999/utils/logger"
	"gorm.io/gorm"
)

type WorkflowService struct{}

func NewWorkflowService() *WorkflowService {
	return &WorkflowService{}
}

func (s *WorkflowService) CreateWorkflow(req model.CreateWorkflowReq) (*model.Workflow, error) {
	if req.Name == "" {
		return nil, errors.New("please provide the name")
	}

	specVersion := model.WorkflowSpecVersion_LATEST
	if req.SpecVersion != "" {
		specVersion = req.SpecVersion
	}

	sourceType, sourceTemplateID, err := mapper.ResolveCreateSourceType(specVersion, req.Data)
	if err != nil {
		return nil, err
	}

	return s.createWorkflowInternal(req, specVersion, sourceType, sourceTemplateID)
}

// CloneWorkflow deep-copies an existing workflow. Name is auto-generated as
// "<source>_copy" (with _copy_N on collision); task/task_group IDs are
// re-issued; runs/snapshots are not copied. The new workflow's first
// snapshot records source_type="clone" + source_template_id=<source id>.
func (s *WorkflowService) CloneWorkflow(srcWfID string) (*model.Workflow, error) {
	if srcWfID == "" {
		return nil, errors.New("please provide the source workflow id")
	}

	src, err := mapper.GetWorkflowFromDB(srcWfID)
	if err != nil {
		return nil, errors.New("source workflow not found: " + err.Error())
	}

	createReq := model.CreateWorkflowReq{
		SpecVersion: src.SpecVersion,
		Name:        nextCloneName(src.Name),
		Data:        mapper.DataToCreateDataReq(src.Data),
	}

	return s.createWorkflowInternal(createReq, src.SpecVersion, "clone", srcWfID)
}

// nextCloneName probes for the first non-colliding "<base>_copy" or
// "<base>_copy_N" name.
func nextCloneName(baseName string) string {
	candidate := baseName + "_copy"
	for i := 2; ; i++ {
		if existing, _ := dao.WorkflowGetByName(candidate); existing == nil {
			return candidate
		}
		candidate = baseName + "_copy_" + strconv.Itoa(i)
	}
}

// createWorkflowInternal is the shared persist+DAG path used by
// CreateWorkflow and CloneWorkflow.
func (s *WorkflowService) createWorkflowInternal(req model.CreateWorkflowReq, specVersion, sourceType, sourceTemplateID string) (*model.Workflow, error) {
	workflowData, err := mapper.CreateDataReqToData(specVersion, req.Data)
	if err != nil {
		return nil, err
	}

	workflow := &model.Workflow{}
	workflow.ID = uuid.New().String()
	workflow.WorkflowKey = uuid.New().String()
	workflow.SpecVersion = specVersion
	workflow.Name = req.Name
	workflow.Data = workflowData

	// Persist the workflow, its task graph, and the initial snapshot atomically.
	// A failure anywhere rolls the whole thing back — no orphan workflow rows,
	// no dangling snapshot/current-version pointer.
	err = db.DB.Transaction(func(tx *gorm.DB) error {
		if _, err := dao.WorkflowCreateTx(tx, workflow); err != nil {
			return err
		}

		for _, tg := range workflow.Data.TaskGroups {
			if _, err := dao.TaskGroupCreateTx(tx, &model.TaskGroupDBModel{
				ID:           tg.ID,
				Name:         tg.Name,
				WorkflowID:   workflow.ID,
				WorkflowKey:  workflow.WorkflowKey,
				TaskGroupKey: tg.ID,
			}); err != nil {
				return err
			}

			for _, t := range tg.Tasks {
				if _, err := dao.TaskCreateTx(tx, &model.TaskDBModel{
					ID:           t.ID,
					Name:         t.Name,
					WorkflowID:   workflow.ID,
					WorkflowKey:  workflow.WorkflowKey,
					TaskGroupID:  tg.ID,
					TaskGroupKey: tg.ID,
					TaskKey:      t.ID,
				}); err != nil {
					return err
				}
			}
		}

		if _, err := dao.WorkflowCreateSnapshotTx(tx, workflow, "create", sourceType, sourceTemplateID); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// DAG generation is a filesystem side-effect outside the transaction. On
	// failure, compensate: this is a brand-new workflow, so remove the
	// freshly-created DAG dir and soft-delete the committed rows, then surface
	// the error.
	client, err := airflow.GetClient()
	if err != nil {
		s.compensateCreate(workflow)
		return nil, err
	}

	if err := client.CreateDAG(workflow); err != nil {
		s.compensateCreate(workflow)
		return nil, errors.New("failed to create the workflow (error: " + err.Error() + ")")
	}

	return workflow, nil
}

// compensateCreate rolls back a brand-new workflow after a post-commit DAG
// failure. Best-effort: each step is logged but not fatal, since the workflow
// is already being torn down.
func (s *WorkflowService) compensateCreate(workflow *model.Workflow) {
	if client, err := airflow.GetClient(); err == nil {
		if derr := client.DeleteDAG(common.WorkflowDagID(workflow), true); derr != nil {
			logger.Println(logger.ERROR, true, "compensateCreate: delete DAG dir: "+derr.Error())
		}
	}
	if err := dao.TaskSoftDeleteByWorkflowID(workflow.ID); err != nil {
		logger.Println(logger.ERROR, true, "compensateCreate: soft-delete tasks: "+err.Error())
	}
	if err := dao.TaskGroupSoftDeleteByWorkflowID(workflow.ID); err != nil {
		logger.Println(logger.ERROR, true, "compensateCreate: soft-delete task groups: "+err.Error())
	}
	if err := dao.WorkflowDelete(workflow); err != nil {
		logger.Println(logger.ERROR, true, "compensateCreate: soft-delete workflow: "+err.Error())
	}
}

func (s *WorkflowService) GetWorkflow(wfId string, includeDeleted bool) (*model.Workflow, error) {
	var (
		workflow *model.Workflow
		err      error
	)
	if includeDeleted {
		workflow, err = mapper.GetWorkflowFromDBIncludeDeleted(wfId)
	} else {
		workflow, err = mapper.GetWorkflowFromDB(wfId)
	}
	if err != nil {
		return nil, err
	}

	client, err := airflow.GetClient()
	if err != nil {
		return nil, err
	}

	_, err = client.GetDAG(common.WorkflowDagID(workflow))
	if err != nil {
		return nil, errors.New("failed to get the workflow from the airflow server")
	}

	return workflow, nil
}

func (s *WorkflowService) GetWorkflowByName(wfName string, includeDeleted bool) (*model.Workflow, error) {
	var (
		workflowByName *model.Workflow
		err            error
	)
	if includeDeleted {
		workflowByName, err = dao.WorkflowGetByNameIncludeDeleted(wfName)
	} else {
		workflowByName, err = dao.WorkflowGetByName(wfName)
	}
	if err != nil {
		return nil, err
	}

	return s.GetWorkflow(workflowByName.ID, includeDeleted)
}

func (s *WorkflowService) ListWorkflow(name string, includeDeleted bool, page int, row int) (*[]model.Workflow, error) {
	workflow := &model.Workflow{Name: name}
	if includeDeleted {
		return dao.WorkflowGetListIncludeDeleted(workflow, page, row)
	}
	return dao.WorkflowGetList(workflow, page, row)
}

func (s *WorkflowService) UpdateWorkflow(wfId string, req model.CreateWorkflowReq) (*model.Workflow, error) {
	oldWorkflow, err := dao.WorkflowGet(wfId)
	if err != nil {
		return nil, err
	}

	// Snapshot of the pre-update definition, captured before any mutation, so
	// the DAG can be restored to its last-known-good state if the post-commit
	// DAG refresh fails.
	preUpdate := *oldWorkflow

	if req.Name != "" {
		oldWorkflow.Name = req.Name
	}

	specVersion := model.WorkflowSpecVersion_LATEST
	if req.SpecVersion != "" {
		specVersion = req.SpecVersion
	}

	workflowData, err := mapper.CreateDataReqToData(specVersion, req.Data)
	if err != nil {
		return nil, err
	}

	diff, err := mapper.BuildWorkflowGraphDiff(oldWorkflow, workflowData)
	if err != nil {
		return nil, err
	}

	oldWorkflow.SpecVersion = specVersion
	oldWorkflow.Data = diff.WorkflowData

	// Apply the whole graph diff + workflow update + snapshot atomically.
	err = db.DB.Transaction(func(tx *gorm.DB) error {
		for _, tg := range diff.TaskGroupsToUpsert {
			taskGroup := tg
			if err := dao.TaskGroupSaveTx(tx, &taskGroup); err != nil {
				return err
			}
		}
		for _, t := range diff.TasksToUpsert {
			task := t
			if err := dao.TaskSaveTx(tx, &task); err != nil {
				return err
			}
		}
		if err := captureSoftDroppedTaskSnapshotsTx(tx, &preUpdate, diff.TasksToSoftDrop, "update_delete"); err != nil {
			return err
		}
		for _, t := range diff.TasksToSoftDrop {
			task := t
			if err := dao.TaskDeleteTx(tx, &task); err != nil {
				return err
			}
		}
		for _, tg := range diff.TaskGroupsToSoftDrop {
			taskGroup := tg
			if err := dao.TaskGroupDeleteTx(tx, &taskGroup); err != nil {
				return err
			}
		}

		if err := dao.WorkflowUpdateTx(tx, oldWorkflow); err != nil {
			return err
		}
		if _, err := dao.WorkflowCreateSnapshotTx(tx, oldWorkflow, "update", "modified", ""); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	client, err := airflow.GetClient()
	if err != nil {
		s.restoreDAG(&preUpdate)
		return nil, err
	}

	if err := client.CreateDAG(oldWorkflow); err != nil {
		s.restoreDAG(&preUpdate)
		return nil, errors.New("failed to update the workflow (error: " + err.Error() + ")")
	}

	return oldWorkflow, nil
}

// restoreDAG best-effort regenerates the DAG from a previous workflow
// definition after a post-commit DAG refresh fails. The DB has already
// advanced to the new definition, so this leaves the DAG one step behind the
// DB until the next successful refresh — logged loudly so the gap is visible.
func (s *WorkflowService) restoreDAG(previous *model.Workflow) {
	client, err := airflow.GetClient()
	if err != nil {
		logger.Println(logger.ERROR, true, "restoreDAG: get airflow client: "+err.Error())
		return
	}
	if err := client.CreateDAG(previous); err != nil {
		logger.Println(logger.ERROR, true,
			"restoreDAG: failed to restore previous DAG; DB and DAG may be out of sync for workflow "+previous.ID+": "+err.Error())
	}
}

func (s *WorkflowService) DeleteWorkflow(wfId string) error {
	workflow, err := dao.WorkflowGet(wfId)
	if err != nil {
		return err
	}

	client, err := airflow.GetClient()
	if err != nil {
		return err
	}

	err = client.DeleteDAG(common.WorkflowDagID(workflow), true)
	if err != nil {
		logger.Println(logger.ERROR, true, "AIRFLOW: "+err.Error())
	}

	activeTasks, err := dao.TaskGetListByWorkflowID(workflow.ID, false)
	if err != nil {
		return err
	}
	if err := captureSoftDroppedTaskSnapshots(workflow, activeTasks, "workflow_delete"); err != nil {
		return err
	}

	if err := dao.TaskSoftDeleteByWorkflowID(workflow.ID); err != nil {
		return err
	}
	if err := dao.TaskGroupSoftDeleteByWorkflowID(workflow.ID); err != nil {
		return err
	}

	err = dao.WorkflowDelete(workflow)
	if err != nil {
		return err
	}

	workflow.IsDeleted = true
	_, err = dao.WorkflowCreateSnapshot(workflow, "delete", "custom", "")
	if err != nil {
		return err
	}

	return nil
}

func (s *WorkflowService) RunWorkflow(wfId string) error {
	workflow, err := dao.WorkflowGet(wfId)
	if err != nil {
		return err
	}

	client, err := airflow.GetClient()
	if err != nil {
		return err
	}

	conf := map[string]interface{}{
		"workflow_id":   workflow.ID,
		"workflow_key":  common.WorkflowDagID(workflow),
		"workflow_name": workflow.Name,
	}
	_, err = client.RunDAG(common.WorkflowDagID(workflow), conf)
	if err != nil {
		return err
	}

	return nil
}

func captureSoftDroppedTaskSnapshots(workflow *model.Workflow, droppedTasks []model.TaskDBModel, snapshotType string) error {
	return captureSoftDroppedTaskSnapshotsTx(db.DB, workflow, droppedTasks, snapshotType)
}

// captureSoftDroppedTaskSnapshotsTx is the transaction-aware variant, used by
// the update/rollback flows so the snapshots commit atomically with the
// soft-deletes they describe.
func captureSoftDroppedTaskSnapshotsTx(tx *gorm.DB, workflow *model.Workflow, droppedTasks []model.TaskDBModel, snapshotType string) error {
	taskMap := workflowTaskByID(workflow)
	for _, taskDB := range droppedTasks {
		rawTask, ok := taskMap[taskDB.ID]
		if !ok {
			rawTask = model.Task{
				ID:            taskDB.ID,
				Name:          taskDB.Name,
				TaskComponent: "",
				Spec:          nil,
				Dependencies:  []string{},
			}
		}
		if rawTask.Dependencies == nil {
			rawTask.Dependencies = []string{}
		}
		if err := dao.TaskSnapshotCreateFromTaskTx(tx, &taskDB, rawTask, snapshotType); err != nil {
			return err
		}
	}
	return nil
}

func (s *WorkflowService) ListWorkflowVersions(wfID string, page, row int) (*[]model.WorkflowVersion, error) {
	filter := &model.WorkflowVersion{WorkflowID: wfID}
	return dao.WorkflowVersionGetList(filter, page, row)
}

func (s *WorkflowService) GetWorkflowVersion(wfID, versionID string) (*model.WorkflowVersion, error) {
	return dao.WorkflowVersionGet(versionID, wfID)
}

// RollbackWorkflow restores the workflow's task graph from the snapshot at
// versionNo (1-based). Diff is computed via mapper.BuildWorkflowGraphDiff —
// the same path UpdateWorkflow uses — so tasks matched by name keep their
// IDs (preserving Airflow per-task history), new tasks get fresh UUIDs, and
// dropped tasks are soft-deleted with a "rollback_drop" snapshot. The
// rollback itself is recorded as a new WorkflowVersion with action="rollback"
// and source_template_id=<source version id>. workflow_schedules untouched.
// Refuses delete-action snapshots and IsDeleted workflows.
func (s *WorkflowService) RollbackWorkflow(wfID string, versionNo int) (*model.Workflow, error) {
	if wfID == "" {
		return nil, errors.New("please provide the workflow id")
	}
	if versionNo <= 0 {
		return nil, errors.New("version_no must be a positive integer")
	}

	workflow, err := dao.WorkflowGet(wfID)
	if err != nil {
		return nil, err
	}
	if workflow.IsDeleted {
		return nil, errors.New("cannot rollback a deleted workflow")
	}

	version, err := dao.WorkflowVersionGetByNo(wfID, versionNo)
	if err != nil {
		return nil, err
	}
	if version == nil {
		return nil, errors.New("workflow version not found")
	}
	if version.Action == "delete" {
		return nil, errors.New("cannot rollback to a delete-action version")
	}

	target := version.RawData
	specVersion := target.SpecVersion
	if specVersion == "" {
		specVersion = model.WorkflowSpecVersion_LATEST
	}

	createDataReq := mapper.DataToCreateDataReq(target.Data)
	newData, err := mapper.CreateDataReqToData(specVersion, createDataReq)
	if err != nil {
		return nil, err
	}

	diff, err := mapper.BuildWorkflowGraphDiff(workflow, newData)
	if err != nil {
		return nil, err
	}

	// Pre-rollback definition, for DAG restore if the post-commit refresh fails.
	preRollback := *workflow

	workflow.Name = target.Name
	workflow.SpecVersion = specVersion
	workflow.Data = diff.WorkflowData

	// Restore the task graph + workflow + rollback snapshot atomically.
	err = db.DB.Transaction(func(tx *gorm.DB) error {
		for _, tg := range diff.TaskGroupsToUpsert {
			taskGroup := tg
			if err := dao.TaskGroupSaveTx(tx, &taskGroup); err != nil {
				return err
			}
		}
		for _, t := range diff.TasksToUpsert {
			task := t
			if err := dao.TaskSaveTx(tx, &task); err != nil {
				return err
			}
		}
		if err := captureSoftDroppedTaskSnapshotsTx(tx, &preRollback, diff.TasksToSoftDrop, "rollback_drop"); err != nil {
			return err
		}
		for _, t := range diff.TasksToSoftDrop {
			task := t
			if err := dao.TaskDeleteTx(tx, &task); err != nil {
				return err
			}
		}
		for _, tg := range diff.TaskGroupsToSoftDrop {
			taskGroup := tg
			if err := dao.TaskGroupDeleteTx(tx, &taskGroup); err != nil {
				return err
			}
		}

		if err := dao.WorkflowUpdateTx(tx, workflow); err != nil {
			return err
		}
		if _, err := dao.WorkflowCreateSnapshotTx(tx, workflow, "rollback", "rollback", version.ID); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	client, err := airflow.GetClient()
	if err != nil {
		s.restoreDAG(&preRollback)
		return nil, err
	}
	if err := client.CreateDAG(workflow); err != nil {
		s.restoreDAG(&preRollback)
		return nil, errors.New("failed to refresh the workflow DAG (error: " + err.Error() + ")")
	}

	return workflow, nil
}

func workflowTaskByID(workflow *model.Workflow) map[string]model.Task {
	tasks := make(map[string]model.Task)
	if workflow == nil {
		return tasks
	}
	for _, tg := range workflow.Data.TaskGroups {
		for _, task := range tg.Tasks {
			tasks[task.ID] = task
		}
	}
	return tasks
}
