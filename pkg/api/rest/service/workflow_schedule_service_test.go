package service

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/cloud-barista/cm-cicada/dao"
	"github.com/cloud-barista/cm-cicada/db"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupScheduleServiceTestDB(t *testing.T) {
	t.Helper()

	old := db.DB
	path := filepath.Join(t.TempDir(), "sched-svc-test.db")
	testDB, err := gorm.Open(sqlite.Open(path), &gorm.Config{TranslateError: true})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	if err := testDB.AutoMigrate(
		&model.Workflow{},
		&model.TaskGroupDBModel{},
		&model.TaskDBModel{},
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

// TestScheduleConflictWhenActiveExists verifies the service maps an existing
// active schedule to ErrActiveScheduleExists (C2) without reaching the DAG
// refresh. The pre-existing schedule is a cron row so syncOverdueOnceSchedule
// returns early and no Airflow client is needed.
func TestScheduleConflictWhenActiveExists(t *testing.T) {
	setupScheduleServiceTestDB(t)

	const wfID = "wf-svc-conflict"
	if _, err := dao.WorkflowCreate(&model.Workflow{
		ID:          wfID,
		WorkflowKey: wfID + "-key",
		SpecVersion: "v0.1",
		Name:        "conflict-wf",
		Data:        model.Data{TaskGroups: []model.TaskGroup{}},
	}); err != nil {
		t.Fatalf("seed workflow: %v", err)
	}

	existingCron := "0 0 * * *"
	if err := dao.WorkflowScheduleCreate(&model.WorkflowSchedule{
		ID:         "existing",
		WorkflowID: wfID,
		Type:       model.WorkflowScheduleTypeCron,
		Cron:       &existingCron,
		Status:     model.WorkflowScheduleStatusActive,
	}); err != nil {
		t.Fatalf("seed active schedule: %v", err)
	}

	svc := NewWorkflowScheduleService()
	newCron := "30 1 * * *"
	_, err := svc.Schedule(wfID, model.CreateWorkflowScheduleReq{Cron: &newCron})
	if !errors.Is(err, ErrActiveScheduleExists) {
		t.Fatalf("Schedule() error = %v, want ErrActiveScheduleExists", err)
	}

	// No second active row should have been inserted.
	var n int64
	if err := db.DB.Model(&model.WorkflowSchedule{}).
		Where("workflow_id = ? AND status = ?", wfID, model.WorkflowScheduleStatusActive).
		Count(&n).Error; err != nil {
		t.Fatalf("count active schedules: %v", err)
	}
	if n != 1 {
		t.Fatalf("active schedule rows = %d, want 1", n)
	}
}
