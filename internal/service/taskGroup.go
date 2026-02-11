package service

import (
	"github.com/cloud-barista/cm-cicada/internal/adapter"
	"github.com/cloud-barista/cm-cicada/internal/domain"
)

func ListTaskGroup(wfId string) ([]domain.TaskGroup, error) {

	workflow, err := adapter.WorkflowGet(wfId)
	if err != nil {
		return nil, err
	}

	return workflow.Data.TaskGroups, nil
}

func GetTaskGroup(wfId string, tgId string) (*domain.TaskGroup, error) {

	workflow, err := adapter.WorkflowGet(wfId)
	if err != nil {
		return nil, err
	}

	for _, tg := range workflow.Data.TaskGroups {
		if tg.ID == tgId {
			return &tg, nil
		}
	}

	return nil, err
}
