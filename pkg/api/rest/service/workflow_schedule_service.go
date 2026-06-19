package service

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/cloud-barista/cm-cicada/dao"
	"github.com/cloud-barista/cm-cicada/db"
	"github.com/cloud-barista/cm-cicada/lib/airflow"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/common"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/mapper"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
)

// ErrActiveScheduleExists is returned when a workflow already has an active
// schedule. It is surfaced both by the in-transaction re-check and by the
// partial unique index (translated gorm.ErrDuplicatedKey).
var ErrActiveScheduleExists = errors.New("workflow already has an active schedule; cancel it first")

type WorkflowScheduleService struct{}

func NewWorkflowScheduleService() *WorkflowScheduleService {
	return &WorkflowScheduleService{}
}

// Schedule registers a one-shot (RunAt) or recurring (Cron) schedule.
// Exactly one of the two must be set. Only one active schedule per workflow
// is allowed; existing active must be canceled first.
func (s *WorkflowScheduleService) Schedule(workflowID string, req model.CreateWorkflowScheduleReq) (*model.WorkflowSchedule, error) {
	if workflowID == "" {
		return nil, errors.New("please provide the workflow id")
	}
	s.syncOverdueOnceSchedule(workflowID)

	hasRunAt := req.RunAt != nil && !req.RunAt.IsZero()
	hasCron := req.Cron != nil && strings.TrimSpace(*req.Cron) != ""
	switch {
	case hasRunAt && hasCron:
		return nil, errors.New("provide exactly one of run_at / cron, not both")
	case !hasRunAt && !hasCron:
		return nil, errors.New("provide one of run_at / cron")
	}

	workflow, err := mapper.GetWorkflowFromDB(workflowID)
	if err != nil {
		return nil, errors.New("workflow not found: " + err.Error())
	}

	row := &model.WorkflowSchedule{
		ID:         uuid.New().String(),
		WorkflowID: workflowID,
		Status:     model.WorkflowScheduleStatusActive,
	}
	if hasRunAt {
		if !req.RunAt.After(time.Now()) {
			return nil, errors.New("run_at must be in the future")
		}
		runAtUTC := req.RunAt.UTC()
		row.Type = model.WorkflowScheduleTypeOnce
		row.RunAt = &runAtUTC
	} else {
		cron := strings.TrimSpace(*req.Cron)
		row.Type = model.WorkflowScheduleTypeCron
		row.Cron = &cron
	}

	// Re-check active uniqueness and insert on the same connection. The
	// partial unique index (idx_workflow_schedules_active) is the real guard
	// against the check-then-create race; the in-tx re-check just yields a
	// friendlier error in the common (non-racing) case.
	err = db.DB.Transaction(func(tx *gorm.DB) error {
		existing, err := dao.WorkflowScheduleGetActiveTx(tx, workflowID)
		if err != nil {
			return err
		}
		if existing != nil {
			return ErrActiveScheduleExists
		}
		return dao.WorkflowScheduleCreateTx(tx, row)
	})
	if err != nil {
		if errors.Is(err, ErrActiveScheduleExists) || errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, ErrActiveScheduleExists
		}
		return nil, err
	}

	// DAG refresh is outside the transaction. On failure, compensate by
	// canceling the just-created schedule so the active-uniqueness invariant
	// is not left occupied by a schedule the DAG never picked up.
	if err := s.refreshDAG(workflow); err != nil {
		if cerr := dao.WorkflowScheduleUpdateStatus(row.ID, model.WorkflowScheduleStatusCanceled); cerr != nil {
			return nil, errors.New(err.Error() + "; additionally failed to roll back the schedule: " + cerr.Error())
		}
		return nil, err
	}

	return row, nil
}

// Cancel marks the workflow's active schedule as canceled and drops the
// schedule line from DAG metadata.
func (s *WorkflowScheduleService) Cancel(workflowID string) (*model.WorkflowSchedule, error) {
	if workflowID == "" {
		return nil, errors.New("please provide the workflow id")
	}
	s.syncOverdueOnceSchedule(workflowID)

	row, err := dao.WorkflowScheduleGetActive(workflowID)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, errors.New("no active schedule for this workflow")
	}

	if err := dao.WorkflowScheduleUpdateStatus(row.ID, model.WorkflowScheduleStatusCanceled); err != nil {
		return nil, err
	}
	row.Status = model.WorkflowScheduleStatusCanceled

	workflow, err := mapper.GetWorkflowFromDB(workflowID)
	if err == nil {
		_ = s.refreshDAG(workflow)
	}

	return row, nil
}

// GetLatest returns the workflow's most recently created schedule row
// regardless of status, or (nil, nil) when there's no schedule history.
func (s *WorkflowScheduleService) GetLatest(workflowID string) (*model.WorkflowSchedule, error) {
	if workflowID == "" {
		return nil, errors.New("please provide the workflow id")
	}
	s.syncOverdueOnceSchedule(workflowID)
	return dao.WorkflowScheduleGetLatest(workflowID)
}

// syncOverdueOnceSchedule promotes the active once schedule to executed when
// a matching scheduler-triggered DAGRun exists. Matching is strict:
// run_type=="scheduled" and logical_date within ±1s of run_at. Cron rows are
// not touched (recurring rules stay active). Best-effort: any error silently
// returns without mutating state.
func (s *WorkflowScheduleService) syncOverdueOnceSchedule(workflowID string) {
	row, err := dao.WorkflowScheduleGetActive(workflowID)
	if err != nil || row == nil {
		return
	}
	if row.Type != model.WorkflowScheduleTypeOnce || row.RunAt == nil {
		return
	}
	if row.RunAt.After(time.Now()) {
		return // not due yet
	}

	workflow, err := mapper.GetWorkflowFromDB(workflowID)
	if err != nil {
		return
	}
	client, err := airflow.GetClient()
	if err != nil {
		return
	}
	runs, err := client.GetDAGRuns(common.WorkflowDagID(workflow))
	if err != nil || runs.DagRuns == nil {
		return
	}

	target := row.RunAt.UTC()
	const tolerance = time.Second
	for _, run := range *runs.DagRuns {
		if run.GetRunType() != "scheduled" {
			continue
		}
		ld := run.GetLogicalDate()
		if ld.IsZero() {
			continue
		}
		diff := ld.UTC().Sub(target)
		if diff < 0 {
			diff = -diff
		}
		if diff <= tolerance {
			_ = dao.WorkflowScheduleUpdateStatus(row.ID, model.WorkflowScheduleStatusExecuted)
			return
		}
	}
}

func (s *WorkflowScheduleService) refreshDAG(workflow *model.Workflow) error {
	client, err := airflow.GetClient()
	if err != nil {
		return err
	}
	if err := client.CreateDAG(workflow); err != nil {
		return errors.New("failed to refresh the DAG metadata (error: " + err.Error() + ")")
	}
	return nil
}
