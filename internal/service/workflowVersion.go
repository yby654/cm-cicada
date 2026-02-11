package service

import (
	"encoding/json"

	"github.com/cloud-barista/cm-cicada/internal/adapter"
	"github.com/cloud-barista/cm-cicada/internal/domain"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
)

func GetWorkflowVersion(id string) (*model.WorkflowVersion, error) {
	workflowVersion, err := adapter.WorkflowVersionGet(id)
	if err != nil {
		return nil, err
	}
	return &model.WorkflowVersion{
		ID:                 workflowVersion.ID,
		WorkflowID:         workflowVersion.WorkflowID,
		Version:            workflowVersion.Version,
		SpecVersion:        workflowVersion.SpecVersion,
		DefinitionSnapshot: json.RawMessage(workflowVersion.DefinitionSnapshot),
		CreatedAt:          workflowVersion.CreatedAt,
	}, nil
}

func ListWorkflowVersion(workflowID string, page int, row int) ([]model.WorkflowVersion, error) {
	workflowVersion := &domain.WorkflowVersion{}
	if workflowID != "" {
		workflowVersion.WorkflowID = workflowID
	}
	workflowVersionList, err := adapter.WorkflowVersionGetList(workflowVersion, page, row)
	if err != nil {
		return nil, err
	}
	workflowVersionDTOList := []model.WorkflowVersion{}
	for _, workflowVersion := range *workflowVersionList {
		workflowVersionDTOList = append(workflowVersionDTOList, model.WorkflowVersion{
			ID:                 workflowVersion.ID,
			WorkflowID:         workflowVersion.WorkflowID,
			Version:            workflowVersion.Version,
			SpecVersion:        workflowVersion.SpecVersion,
			DefinitionSnapshot: json.RawMessage(workflowVersion.DefinitionSnapshot),
			CreatedAt:          workflowVersion.CreatedAt,
		})
	}

	return workflowVersionDTOList, nil
}
