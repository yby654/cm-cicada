package dao

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/cloud-barista/cm-cicada/db"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTxTestDB swaps db.DB for a temp SQLite database migrated with every
// table the transaction flows touch, plus the partial unique index that
// guards active-schedule uniqueness (mirrors db.ensureWorkflowScheduleActiveUniqueIndex).
// TranslateError is enabled so UNIQUE violations surface as gorm.ErrDuplicatedKey,
// matching the production configuration in db.Open.
func setupTxTestDB(t *testing.T) {
	t.Helper()

	old := db.DB
	path := filepath.Join(t.TempDir(), "tx-test.db")
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
		&model.WorkflowSchedule{},
	); err != nil {
		t.Fatalf("failed to migrate test tables: %v", err)
	}

	if err := testDB.Exec(
		"CREATE UNIQUE INDEX IF NOT EXISTS idx_workflow_schedules_active " +
			"ON workflow_schedules(workflow_id) WHERE status = 'active'",
	).Error; err != nil {
		t.Fatalf("failed to create active schedule unique index: %v", err)
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

func newTestWorkflow(id string) *model.Workflow {
	return &model.Workflow{
		ID:          id,
		WorkflowKey: id + "-key",
		SpecVersion: "v0.1",
		Name:        "wf-" + id,
		Data:        model.Data{TaskGroups: []model.TaskGroup{}},
	}
}

func countRows(t *testing.T, m any, query string, args ...any) int64 {
	t.Helper()
	var n int64
	if err := db.DB.Model(m).Where(query, args...).Count(&n).Error; err != nil {
		t.Fatalf("count: %v", err)
	}
	return n
}

// TestTransactionRollbackLeavesNoState proves the Tx variants actually
// participate in db.DB.Transaction: when the callback returns an error after
// inserting a workflow + task group + task, every row is rolled back. This is
// the atomicity guarantee the create/update/rollback flows rely on (C1/C3).
func TestTransactionRollbackLeavesNoState(t *testing.T) {
	setupTxTestDB(t)

	wf := newTestWorkflow("rb1")
	sentinel := errors.New("forced failure")

	err := db.DB.Transaction(func(tx *gorm.DB) error {
		if _, e := WorkflowCreateTx(tx, wf); e != nil {
			return e
		}
		if _, e := TaskGroupCreateTx(tx, &model.TaskGroupDBModel{
			ID: "tg1", Name: "g", WorkflowID: wf.ID, WorkflowKey: wf.WorkflowKey, TaskGroupKey: "tg1",
		}); e != nil {
			return e
		}
		if _, e := TaskCreateTx(tx, &model.TaskDBModel{
			ID: "t1", Name: "t", WorkflowID: wf.ID, WorkflowKey: wf.WorkflowKey,
			TaskGroupID: "tg1", TaskGroupKey: "tg1", TaskKey: "t1",
		}); e != nil {
			return e
		}
		return sentinel
	})

	if !errors.Is(err, sentinel) {
		t.Fatalf("transaction error = %v, want %v", err, sentinel)
	}

	if n := countRows(t, &model.Workflow{}, "id = ?", wf.ID); n != 0 {
		t.Fatalf("workflow rows = %d, want 0 (rolled back)", n)
	}
	if n := countRows(t, &model.TaskGroupDBModel{}, "workflow_id = ?", wf.ID); n != 0 {
		t.Fatalf("task group rows = %d, want 0 (rolled back)", n)
	}
	if n := countRows(t, &model.TaskDBModel{}, "workflow_id = ?", wf.ID); n != 0 {
		t.Fatalf("task rows = %d, want 0 (rolled back)", n)
	}
}

// TestTransactionCommitPersistsAll is the positive counterpart: a callback
// that returns nil commits every insert.
func TestTransactionCommitPersistsAll(t *testing.T) {
	setupTxTestDB(t)

	wf := newTestWorkflow("ok1")

	err := db.DB.Transaction(func(tx *gorm.DB) error {
		if _, e := WorkflowCreateTx(tx, wf); e != nil {
			return e
		}
		if _, e := TaskGroupCreateTx(tx, &model.TaskGroupDBModel{
			ID: "tg2", Name: "g", WorkflowID: wf.ID, WorkflowKey: wf.WorkflowKey, TaskGroupKey: "tg2",
		}); e != nil {
			return e
		}
		if _, e := TaskCreateTx(tx, &model.TaskDBModel{
			ID: "t2", Name: "t", WorkflowID: wf.ID, WorkflowKey: wf.WorkflowKey,
			TaskGroupID: "tg2", TaskGroupKey: "tg2", TaskKey: "t2",
		}); e != nil {
			return e
		}
		return nil
	})
	if err != nil {
		t.Fatalf("transaction returned unexpected error: %v", err)
	}

	if n := countRows(t, &model.Workflow{}, "id = ?", wf.ID); n != 1 {
		t.Fatalf("workflow rows = %d, want 1", n)
	}
	if n := countRows(t, &model.TaskGroupDBModel{}, "workflow_id = ?", wf.ID); n != 1 {
		t.Fatalf("task group rows = %d, want 1", n)
	}
	if n := countRows(t, &model.TaskDBModel{}, "workflow_id = ?", wf.ID); n != 1 {
		t.Fatalf("task rows = %d, want 1", n)
	}
}

// TestWorkflowCreateSnapshotTxMonotonicVersion verifies that two snapshots
// created within the same transaction get monotonic version numbers (the
// second sees the first's uncommitted row) and that the workflow's
// current_version_id ends pointing at the latest snapshot.
func TestWorkflowCreateSnapshotTxMonotonicVersion(t *testing.T) {
	setupTxTestDB(t)

	wf := newTestWorkflow("snap1")
	if _, err := WorkflowCreate(wf); err != nil {
		t.Fatalf("seed workflow: %v", err)
	}

	err := db.DB.Transaction(func(tx *gorm.DB) error {
		if _, e := WorkflowCreateSnapshotTx(tx, wf, "create", "custom", ""); e != nil {
			return e
		}
		if _, e := WorkflowCreateSnapshotTx(tx, wf, "update", "modified", ""); e != nil {
			return e
		}
		return nil
	})
	if err != nil {
		t.Fatalf("snapshot transaction: %v", err)
	}

	var versions []model.WorkflowVersion
	if err := db.DB.Where("workflowId = ?", wf.ID).Order("version_no ASC").Find(&versions).Error; err != nil {
		t.Fatalf("load versions: %v", err)
	}
	if len(versions) != 2 {
		t.Fatalf("version count = %d, want 2", len(versions))
	}
	if versions[0].VersionNo != 1 || versions[1].VersionNo != 2 {
		t.Fatalf("version numbers = [%d, %d], want [1, 2]", versions[0].VersionNo, versions[1].VersionNo)
	}

	reloaded, err := WorkflowGet(wf.ID)
	if err != nil {
		t.Fatalf("reload workflow: %v", err)
	}
	if reloaded.CurrentVersionID != versions[1].ID {
		t.Fatalf("current_version_id = %q, want %q (latest snapshot)", reloaded.CurrentVersionID, versions[1].ID)
	}
}

// TestTaskSaveTxInsertThenUpdate exercises both branches of TaskSaveTx (create
// when absent, update when present) and confirms the non-tx TaskSave delegates
// to the same logic on the global handle.
func TestTaskSaveTxInsertThenUpdate(t *testing.T) {
	setupTxTestDB(t)

	wf := newTestWorkflow("save1")
	if _, err := WorkflowCreate(wf); err != nil {
		t.Fatalf("seed workflow: %v", err)
	}

	// Insert via the non-tx delegating wrapper.
	if err := TaskSave(&model.TaskDBModel{
		ID: "ts1", Name: "orig", WorkflowID: wf.ID, WorkflowKey: wf.WorkflowKey,
		TaskGroupID: "g1", TaskGroupKey: "g1", TaskKey: "ts1",
	}); err != nil {
		t.Fatalf("initial TaskSave: %v", err)
	}

	// Update via the tx variant.
	err := db.DB.Transaction(func(tx *gorm.DB) error {
		return TaskSaveTx(tx, &model.TaskDBModel{
			ID: "ts1", Name: "updated", WorkflowID: wf.ID, WorkflowKey: wf.WorkflowKey,
			TaskGroupID: "g1", TaskGroupKey: "g1", TaskKey: "ts1",
		})
	})
	if err != nil {
		t.Fatalf("TaskSaveTx update: %v", err)
	}

	got, err := TaskGet("ts1")
	if err != nil {
		t.Fatalf("reload task: %v", err)
	}
	if got.Name != "updated" {
		t.Fatalf("task name = %q, want %q", got.Name, "updated")
	}
	if n := countRows(t, &model.TaskDBModel{}, "id = ?", "ts1"); n != 1 {
		t.Fatalf("task rows = %d, want 1 (update, not insert)", n)
	}
}
