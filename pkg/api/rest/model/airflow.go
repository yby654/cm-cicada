package model

import "time"

// ============================================================================
// API Request/Response Types
// 이들은 DB 모델이 아닌 API 응답/요청용 구조체입니다.
// ============================================================================

// TaskInstance represents a task instance in workflow runs
type TaskInstance struct {
	WorkflowRunID                string    `json:"workflow_run_id,omitempty"`
	WorkflowID                   *string   `json:"workflow_id,omitempty"`
	TaskID                       string    `json:"task_id,omitempty"`
	TaskName                     string    `json:"task_name,omitempty"`
	State                        string    `json:"state,omitempty"`
	StartDate                    time.Time `json:"start_date,omitempty"`
	EndDate                      time.Time `json:"end_date,omitempty"`
	DurationDate                 float64   `json:"duration_date"`
	ExecutionDate                time.Time `json:"execution_date,omitempty"`
	TryNumber                    int       `json:"try_number"`
	IsSoftwareMigrationTask      bool      `json:"is_software_migration_task"`
	SoftwareMigrationExecutionID string    `json:"software_migration_execution_id,omitempty"`
}

// TaskInstanceReference represents a reference to a task instance
type TaskInstanceReference struct {
	// The task ID.
	TaskId   *string `json:"task_id,omitempty"`
	TaskName string  `json:"task_name,omitempty"`
	// The DAG ID.
	WorkflowID *string `json:"workflow_id,omitempty"`
	// The DAG run ID.
	WorkflowRunID *string `json:"workflow_run_id,omitempty"`
	ExecutionDate *string `json:"execution_date,omitempty"`
}

// TaskLog represents task log content
type TaskLog struct {
	Content string `json:"content,omitempty"`
}

// EventLogs represents a collection of event logs
type EventLogs struct {
	EventLogs    []EventLog `json:"event_logs"`
	TotalEntries int        `json:"total_entries"`
}

// EventLog represents a single event log entry
type EventLog struct {
	WorkflowRunID string    `json:"workflow_run_id"`
	RunID         string    `json:"run_id,omitempty"`
	WorkflowID    string    `json:"workflow_id"`
	TaskID        string    `json:"task_id"`
	TaskName      string    `json:"task_name"`
	Event         string    `json:"event,omitempty"`
	When          time.Time `json:"when,omitempty"`
	Extra         string    `json:"extra,omitempty"`
}

// TaskClearOption represents options for clearing tasks
type TaskClearOption struct {
	DryRun            bool     `json:"dryRun"`
	TaskIds           []string `json:"taskIds"`
	IncludeDownstream bool     `json:"includeDownstream"`
	//IncludeFuture     bool     `json:"includeFuture"`
	//IncludeParentdag  bool     `json:"includeParentdag"`
	//IncludePast       bool     `json:"includePast"`
	//IncludeSubdags    bool     `json:"includeSubdags"`
	IncludeUpstream bool `json:"includeUpstream"`
	OnlyFailed      bool `json:"onlyFailed"`
	OnlyRunning     bool `json:"onlyRunning"`
	ResetDagRuns    bool `json:"resetDagRuns"`
}

type ImportErrorCollection struct {
	ImportErrors []ImportError `json:"import_errors"`
	TotalEntries int           `json:"total_entries"`
}

type ImportError struct {
	Error string `json:"error"`
}
