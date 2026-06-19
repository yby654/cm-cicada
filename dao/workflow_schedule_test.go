package dao

import (
	"errors"
	"testing"

	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
	"gorm.io/gorm"
)

func strptr(s string) *string { return &s }

func activeCronSchedule(id, workflowID string) *model.WorkflowSchedule {
	return &model.WorkflowSchedule{
		ID:         id,
		WorkflowID: workflowID,
		Type:       model.WorkflowScheduleTypeCron,
		Cron:       strptr("0 0 * * *"),
		Status:     model.WorkflowScheduleStatusActive,
	}
}

// TestActiveScheduleUniqueSecondActiveFails is the core C2 guarantee: the
// partial unique index rejects a second active schedule for the same workflow,
// surfaced as gorm.ErrDuplicatedKey (TranslateError on).
func TestActiveScheduleUniqueSecondActiveFails(t *testing.T) {
	setupTxTestDB(t)

	if err := WorkflowScheduleCreate(activeCronSchedule("s1", "wf-c2")); err != nil {
		t.Fatalf("first active schedule should succeed: %v", err)
	}

	err := WorkflowScheduleCreate(activeCronSchedule("s2", "wf-c2"))
	if !errors.Is(err, gorm.ErrDuplicatedKey) {
		t.Fatalf("second active schedule error = %v, want gorm.ErrDuplicatedKey", err)
	}

	if n := countRows(t, &model.WorkflowSchedule{}, "workflow_id = ? AND status = ?", "wf-c2", model.WorkflowScheduleStatusActive); n != 1 {
		t.Fatalf("active schedule rows = %d, want 1", n)
	}
}

// TestActiveScheduleUniqueCancelThenReschedule confirms that canceling the
// active schedule frees the slot so a new active schedule can be created.
func TestActiveScheduleUniqueCancelThenReschedule(t *testing.T) {
	setupTxTestDB(t)

	if err := WorkflowScheduleCreate(activeCronSchedule("s1", "wf-recycle")); err != nil {
		t.Fatalf("first active schedule: %v", err)
	}
	if err := WorkflowScheduleUpdateStatus("s1", model.WorkflowScheduleStatusCanceled); err != nil {
		t.Fatalf("cancel: %v", err)
	}

	if err := WorkflowScheduleCreate(activeCronSchedule("s2", "wf-recycle")); err != nil {
		t.Fatalf("reschedule after cancel should succeed: %v", err)
	}
}

// TestActiveScheduleUniqueDifferentWorkflows confirms the index is scoped per
// workflow, not global.
func TestActiveScheduleUniqueDifferentWorkflows(t *testing.T) {
	setupTxTestDB(t)

	if err := WorkflowScheduleCreate(activeCronSchedule("a1", "wf-A")); err != nil {
		t.Fatalf("workflow A active schedule: %v", err)
	}
	if err := WorkflowScheduleCreate(activeCronSchedule("b1", "wf-B")); err != nil {
		t.Fatalf("workflow B active schedule should succeed independently: %v", err)
	}
}

// TestActiveScheduleUniqueNonActiveNotConstrained confirms the partial index
// only constrains active rows — multiple canceled/executed rows coexist.
func TestActiveScheduleUniqueNonActiveNotConstrained(t *testing.T) {
	setupTxTestDB(t)

	c1 := activeCronSchedule("c1", "wf-history")
	c1.Status = model.WorkflowScheduleStatusCanceled
	c2 := activeCronSchedule("c2", "wf-history")
	c2.Status = model.WorkflowScheduleStatusCanceled
	x1 := activeCronSchedule("x1", "wf-history")
	x1.Status = model.WorkflowScheduleStatusExecuted

	for _, s := range []*model.WorkflowSchedule{c1, c2, x1} {
		if err := WorkflowScheduleCreate(s); err != nil {
			t.Fatalf("non-active schedule %s should be allowed: %v", s.ID, err)
		}
	}

	// And one active alongside the history rows is still permitted.
	if err := WorkflowScheduleCreate(activeCronSchedule("act", "wf-history")); err != nil {
		t.Fatalf("active schedule alongside history rows: %v", err)
	}
}
