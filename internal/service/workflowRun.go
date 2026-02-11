package service

import (
	"github.com/cloud-barista/cm-cicada/internal/adapter"
	"github.com/cloud-barista/cm-cicada/internal/domain"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
)

// CreateWorkflowRun creates a WorkflowRun in the DB and returns the response DTO.
// Airflow에서 DAG 종료 시 POST로 보낸 데이터를 저장하는 용도. 실행내역 전부 보존.

func CreateWorkflowRun(createWorkflowRunReq model.CreateWorkflowRunReq) (*domain.WorkflowRun, error) {
	workflowRun := &domain.WorkflowRun{
		ID:            createWorkflowRunReq.WorkflowRunID,
		WorkflowID:    createWorkflowRunReq.WorkflowID,
		ExecutionDate: createWorkflowRunReq.ExecutionDate,
		StartDate:     createWorkflowRunReq.StartDate,
		EndDate:       createWorkflowRunReq.EndDate,
		DurationDate:  createWorkflowRunReq.EndDate.Sub(createWorkflowRunReq.StartDate).Seconds(),
		RunType:       createWorkflowRunReq.RunType,
		State:         createWorkflowRunReq.State,
	}

	savedWorkflowRun, err := adapter.WorkflowRunCreate(workflowRun)
	if err != nil {
		return nil, err
	}
	return savedWorkflowRun, nil
}

func GetWorkflowRun(id string) (*domain.WorkflowRun, error) {
	return adapter.WorkflowRunGet(id)
}

func GetWorkflowRunByWorkflowID(workflowID string) (*[]domain.WorkflowRun, error) {
	return adapter.WorkflowRunGetByWorkflowID(workflowID)
}

func UpdateWorkflowRun(workflowRun *domain.WorkflowRun) error {
	return adapter.WorkflowRunUpdate(workflowRun)
}

func DeleteWorkflowRun(workflowRun *domain.WorkflowRun) error {
	return adapter.WorkflowRunDelete(workflowRun)
}
