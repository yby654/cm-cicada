package model

import "time"

// WorkflowRun is the response model for workflow run.
type WorkflowRun struct {
	ID            string    `json:"id" mapstructure:"id"`
	WorkflowID    string    `json:"workflow_id" mapstructure:"workflow_id"`
	ExecutionDate time.Time `json:"execution_date" mapstructure:"execution_date"`
	StartDate     time.Time `json:"start_date" mapstructure:"start_date"`
	EndDate       time.Time `json:"end_date" mapstructure:"end_date"`
	DurationDate  float64   `json:"duration_date" mapstructure:"duration_date"`
	RunType       string    `json:"run_type" mapstructure:"run_type"`
	State         string    `json:"state" mapstructure:"state"`
}

// CreateWorkflowRunReq is the request body for Airflow DAG 종료 시 POST callback.
type CreateWorkflowRunReq struct {
	WorkflowRunID string    `json:"workflow_run_id" mapstructure:"workflow_run_id" validate:"required"`
	WorkflowID    string    `json:"workflow_id" mapstructure:"workflow_id" validate:"required"`
	ExecutionDate time.Time `json:"execution_date" mapstructure:"execution_date"`
	StartDate     time.Time `json:"start_date" mapstructure:"start_date"`
	EndDate       time.Time `json:"end_date" mapstructure:"end_date"`
	RunType       string    `json:"run_type" mapstructure:"run_type"`
	State         string    `json:"state" mapstructure:"state"`
}
