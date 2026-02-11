package service

import (
	"errors"

	"github.com/cloud-barista/cm-cicada/internal/adapter"
)

type WorkflowValidationInput struct {
	WorkflowID    *string
	TaskGroupID   *string
	WorkflowRunID *string
	TaskID        *string
}

type WorkflowService struct{}

func (s *WorkflowService) ValidateWorkflowInfo(input WorkflowValidationInput) error {
	if input.WorkflowID != nil {
		if *input.WorkflowID == "" {
			return errors.New("workflowId is empty")
		}
		if !adapter.ExistsWorkflow(*input.WorkflowID) {
			return errors.New("workflow not found")
		}
	}

	if input.TaskGroupID != nil {
		if *input.TaskGroupID == "" {
			return errors.New("taskGroupId is empty")
		}
		if !adapter.ExistsTaskGroup(*input.TaskGroupID) {
			return errors.New("taskGroup not found")
		}
	}

	if input.WorkflowRunID != nil {
		if *input.WorkflowRunID == "" {
			return errors.New("workflowRunId is empty")
		}
		if !adapter.ExistsWorkflowRun(*input.WorkflowRunID) {
			return errors.New("workflowRun not found")
		}
	}

	if input.TaskID != nil {
		if *input.TaskID == "" {
			return errors.New("taskId is empty")
		}
		if !adapter.ExistsTask(*input.TaskID) {
			return errors.New("task not found")
		}
	}

	return nil
}
