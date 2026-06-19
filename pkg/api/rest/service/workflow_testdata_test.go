package service

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/cloud-barista/cm-cicada/dao"
	"github.com/cloud-barista/cm-cicada/db"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/mapper"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupWorkflowServiceTestDB swaps db.DB for a temp SQLite database migrated
// with every table the create/update transactions touch.
func setupWorkflowServiceTestDB(t *testing.T) {
	t.Helper()

	old := db.DB
	path := filepath.Join(t.TempDir(), "wf-svc-test.db")
	testDB, err := gorm.Open(sqlite.Open(path), &gorm.Config{TranslateError: true})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	if err := testDB.AutoMigrate(
		&model.Workflow{},
		&model.TaskGroupDBModel{},
		&model.TaskDBModel{},
		&model.TaskSnapshot{},
		&model.WorkflowVersion{},
	); err != nil {
		t.Fatalf("failed to migrate test tables: %v", err)
	}

	db.DB = testDB
	t.Cleanup(func() {
		if db.DB != nil {
			if sqlDB, derr := db.DB.DB(); derr == nil {
				_ = sqlDB.Close()
			}
		}
		db.DB = old
	})
}

// loadCreateWorkflowReq reads a CreateWorkflowReq payload from the repo's
// docs/test testdata fixtures (the same JSON the API regression suite POSTs).
func loadCreateWorkflowReq(t *testing.T, name string) model.CreateWorkflowReq {
	t.Helper()
	path := filepath.Join("..", "..", "..", "..", "docs", "test", "tests", "testdata", "workflow", name)
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read testdata %s: %v", path, err)
	}
	var req model.CreateWorkflowReq
	if err := json.Unmarshal(raw, &req); err != nil {
		t.Fatalf("unmarshal %s: %v", name, err)
	}
	return req
}

func countWorkflowRows(t *testing.T, m any, query string, args ...any) int64 {
	t.Helper()
	var n int64
	if err := db.DB.Model(m).Where(query, args...).Count(&n).Error; err != nil {
		t.Fatalf("count: %v", err)
	}
	return n
}

// createGraphTx mirrors createWorkflowInternal's transaction body so the real
// fixture data is exercised through the exact Tx-variant sequence the service
// commits. forceFail injects an error after all inserts to assert rollback.
func createGraphTx(wf *model.Workflow, sourceType, sourceTemplateID string, forceFail error) error {
	return db.DB.Transaction(func(tx *gorm.DB) error {
		if _, err := dao.WorkflowCreateTx(tx, wf); err != nil {
			return err
		}
		for _, tg := range wf.Data.TaskGroups {
			if _, err := dao.TaskGroupCreateTx(tx, &model.TaskGroupDBModel{
				ID: tg.ID, Name: tg.Name, WorkflowID: wf.ID, WorkflowKey: wf.WorkflowKey, TaskGroupKey: tg.ID,
			}); err != nil {
				return err
			}
			for _, task := range tg.Tasks {
				if _, err := dao.TaskCreateTx(tx, &model.TaskDBModel{
					ID: task.ID, Name: task.Name, WorkflowID: wf.ID, WorkflowKey: wf.WorkflowKey,
					TaskGroupID: tg.ID, TaskGroupKey: tg.ID, TaskKey: task.ID,
				}); err != nil {
					return err
				}
			}
		}
		if _, err := dao.WorkflowCreateSnapshotTx(tx, wf, "create", sourceType, sourceTemplateID); err != nil {
			return err
		}
		return forceFail
	})
}

func workflowFromCreateReq(t *testing.T, req model.CreateWorkflowReq) *model.Workflow {
	t.Helper()
	data, err := mapper.CreateDataReqToData(req.SpecVersion, req.Data)
	if err != nil {
		t.Fatalf("CreateDataReqToData: %v", err)
	}
	return &model.Workflow{
		ID:          uuid.New().String(),
		WorkflowKey: uuid.New().String(),
		SpecVersion: req.SpecVersion,
		Name:        req.Name,
		Data:        data,
	}
}

// TestCreateWorkflowGraphFromTestdataPersists runs the real create.json fixture
// (1 task group, 11 tasks with specs + dependencies) through the create
// transaction and verifies the whole graph + initial snapshot commits, and
// that the graph round-trips through the JSON data column intact.
func TestCreateWorkflowGraphFromTestdataPersists(t *testing.T) {
	setupWorkflowServiceTestDB(t)

	req := loadCreateWorkflowReq(t, "create.json")
	wf := workflowFromCreateReq(t, req)

	// Sanity: the fixture really is the multi-task booker workflow.
	if len(wf.Data.TaskGroups) != 1 || len(wf.Data.TaskGroups[0].Tasks) != 10 {
		t.Fatalf("fixture shape = %d groups / %d tasks, want 1/10",
			len(wf.Data.TaskGroups), len(wf.Data.TaskGroups[0].Tasks))
	}

	if err := createGraphTx(wf, "custom", "", nil); err != nil {
		t.Fatalf("create transaction: %v", err)
	}

	if n := countWorkflowRows(t, &model.Workflow{}, "id = ?", wf.ID); n != 1 {
		t.Fatalf("workflow rows = %d, want 1", n)
	}
	if n := countWorkflowRows(t, &model.TaskGroupDBModel{}, "workflow_id = ?", wf.ID); n != 1 {
		t.Fatalf("task group rows = %d, want 1", n)
	}
	if n := countWorkflowRows(t, &model.TaskDBModel{}, "workflow_id = ?", wf.ID); n != 10 {
		t.Fatalf("task rows = %d, want 10", n)
	}
	if n := countWorkflowRows(t, &model.WorkflowVersion{}, "workflowId = ? AND version_no = ?", wf.ID, 1); n != 1 {
		t.Fatalf("snapshot v1 rows = %d, want 1", n)
	}

	// The graph survives the round-trip through the JSON data column,
	// dependencies included.
	reloaded, err := dao.WorkflowGet(wf.ID)
	if err != nil {
		t.Fatalf("reload workflow: %v", err)
	}
	tasks := reloaded.Data.TaskGroups[0].Tasks
	if len(tasks) != 10 {
		t.Fatalf("reloaded task count = %d, want 10", len(tasks))
	}
	depsByName := map[string][]string{}
	for _, task := range tasks {
		depsByName[task.Name] = task.Dependencies
	}
	if got := depsByName["update_booking_1_json"]; len(got) != 2 {
		t.Fatalf("update_booking_1_json dependencies = %v, want 2 (create_token, get_booking_1)", got)
	}
	if reloaded.CurrentVersionID == "" {
		t.Fatalf("current_version_id not set after create")
	}
}

// TestCreateWorkflowGraphFromTestdataRollback proves that a failure anywhere in
// the create transaction (here injected after the full 11-task graph is
// staged) rolls every row back — no orphan workflow, group, task, or snapshot.
func TestCreateWorkflowGraphFromTestdataRollback(t *testing.T) {
	setupWorkflowServiceTestDB(t)

	req := loadCreateWorkflowReq(t, "create.json")
	wf := workflowFromCreateReq(t, req)

	boom := errors.New("simulated DAG failure before commit")
	err := createGraphTx(wf, "custom", "", boom)
	if !errors.Is(err, boom) {
		t.Fatalf("transaction error = %v, want %v", err, boom)
	}

	if n := countWorkflowRows(t, &model.Workflow{}, "id = ?", wf.ID); n != 0 {
		t.Fatalf("workflow rows = %d, want 0 (rolled back)", n)
	}
	if n := countWorkflowRows(t, &model.TaskGroupDBModel{}, "workflow_id = ?", wf.ID); n != 0 {
		t.Fatalf("task group rows = %d, want 0 (rolled back)", n)
	}
	if n := countWorkflowRows(t, &model.TaskDBModel{}, "workflow_id = ?", wf.ID); n != 0 {
		t.Fatalf("task rows = %d, want 0 (rolled back)", n)
	}
	if n := countWorkflowRows(t, &model.WorkflowVersion{}, "workflowId = ?", wf.ID); n != 0 {
		t.Fatalf("snapshot rows = %d, want 0 (rolled back)", n)
	}
}

// TestUpdateWorkflowDiffFromTestdata seeds the create.json graph, then computes
// the diff against update.json (same task names, changed specs) and applies it
// in a transaction. Tasks matched by name keep their IDs, nothing is dropped,
// and the task count stays 11 — verifying realistic update data flows through
// BuildWorkflowGraphDiff + the Tx-variant upserts.
func TestUpdateWorkflowDiffFromTestdata(t *testing.T) {
	setupWorkflowServiceTestDB(t)

	createReq := loadCreateWorkflowReq(t, "create.json")
	wf := workflowFromCreateReq(t, createReq)
	if err := createGraphTx(wf, "custom", "", nil); err != nil {
		t.Fatalf("seed create transaction: %v", err)
	}

	original, err := dao.WorkflowGet(wf.ID)
	if err != nil {
		t.Fatalf("reload seeded workflow: %v", err)
	}
	idByName := map[string]string{}
	for _, task := range original.Data.TaskGroups[0].Tasks {
		idByName[task.Name] = task.ID
	}

	updateReq := loadCreateWorkflowReq(t, "update.json")
	updateData, err := mapper.CreateDataReqToData(updateReq.SpecVersion, updateReq.Data)
	if err != nil {
		t.Fatalf("CreateDataReqToData(update): %v", err)
	}

	diff, err := mapper.BuildWorkflowGraphDiff(original, updateData)
	if err != nil {
		t.Fatalf("BuildWorkflowGraphDiff: %v", err)
	}
	if len(diff.TasksToSoftDrop) != 0 || len(diff.TaskGroupsToSoftDrop) != 0 {
		t.Fatalf("expected no drops (same task names), got %d task / %d group drops",
			len(diff.TasksToSoftDrop), len(diff.TaskGroupsToSoftDrop))
	}

	err = db.DB.Transaction(func(tx *gorm.DB) error {
		for _, tg := range diff.TaskGroupsToUpsert {
			taskGroup := tg
			if e := dao.TaskGroupSaveTx(tx, &taskGroup); e != nil {
				return e
			}
		}
		for _, task := range diff.TasksToUpsert {
			tt := task
			if e := dao.TaskSaveTx(tx, &tt); e != nil {
				return e
			}
		}
		original.Name = updateReq.Name
		original.Data = diff.WorkflowData
		if e := dao.WorkflowUpdateTx(tx, original); e != nil {
			return e
		}
		_, e := dao.WorkflowCreateSnapshotTx(tx, original, "update", "modified", "")
		return e
	})
	if err != nil {
		t.Fatalf("update transaction: %v", err)
	}

	// Still exactly 10 active tasks, and IDs preserved (matched by name).
	if n := countWorkflowRows(t, &model.TaskDBModel{}, "workflow_id = ? AND is_deleted = ?", wf.ID, false); n != 10 {
		t.Fatalf("active task rows after update = %d, want 10", n)
	}
	updated, err := dao.WorkflowGet(wf.ID)
	if err != nil {
		t.Fatalf("reload updated workflow: %v", err)
	}
	for _, task := range updated.Data.TaskGroups[0].Tasks {
		if idByName[task.Name] != task.ID {
			t.Fatalf("task %q id changed on update: %q -> %q (name match should keep id)",
				task.Name, idByName[task.Name], task.ID)
		}
	}
	if updated.Name != "booker_custom_workflow_updated" {
		t.Fatalf("workflow name = %q, want booker_custom_workflow_updated", updated.Name)
	}
}
