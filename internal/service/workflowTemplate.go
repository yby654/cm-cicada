package service

import (
	"errors"

	"github.com/cloud-barista/cm-cicada/internal/adapter"
	"github.com/cloud-barista/cm-cicada/internal/domain"
)

func GetWorkflowTemplate(id string) (*domain.GetWorkflowTemplate, error) {
	if id == "" {
		return nil, errors.New("wftId is empty")
	}
	return adapter.WorkflowTemplateGet(id)
}

func ListWorkflowTemplate(name string, page int, row int) (*[]domain.WorkflowTemplate, error) {
	filter := &domain.WorkflowTemplate{
		Name: name,
	}
	return adapter.WorkflowTemplateGetList(filter, page, row)
}

func GetWorkflowTemplateByName(name string) (*domain.GetWorkflowTemplate, error) {
	if name == "" {
		return nil, errors.New("wfName is empty")
	}

	dbModel, err := adapter.WorkflowTemplateGetByName(name)
	if err != nil {
		return nil, err
	}

	return &domain.GetWorkflowTemplate{
		SpecVersion: dbModel.SpecVersion,
		Name:        dbModel.Name,
		Data:        dbModel.Data,
	}, nil
}

