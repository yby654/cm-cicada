package service

import (
	"github.com/cloud-barista/cm-cicada/internal/adapter"
	"github.com/cloud-barista/cm-cicada/internal/domain"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
	"github.com/google/uuid"
)

func CreateTaskComponent(createTaskComponentReq model.CreateTaskComponentReq) (*domain.TaskComponent, error) {
	taskComponent := &domain.TaskComponent{
		ID:   uuid.New().String(),
		Name: createTaskComponentReq.Name,
		Data: createTaskComponentReq.Data,
	}

	savedTaskComponent, err := adapter.TaskComponentCreate(taskComponent)
	if err != nil {
		return nil, err
	}
	return savedTaskComponent, nil
}

func GetTaskComponent(id string) (*domain.TaskComponent, error) {

	taskComponent, err := adapter.TaskComponentGet(id)
	if err != nil {
		return nil, err
	}
	return taskComponent, nil
}

func GetTaskComponentByName(name string) (*domain.TaskComponent, error) {

	taskComponent, err := adapter.TaskComponentGetByName(name)
	if err != nil {
		return nil, err
	}
	return taskComponent, nil
}

func ListTaskComponent(page int, row int) (*[]domain.TaskComponent, error) {
	taskComponentList, err := adapter.TaskComponentGetList(page, row)
	if err != nil {
		return nil, err
	}
	return taskComponentList, nil
}

func UpdateTaskComponent(id string, updateTaskComponentReq model.CreateTaskComponentReq) (*domain.TaskComponent, error) {

	oldTaskComponent, err := adapter.TaskComponentGet(id)
	if err != nil {
		return nil, err
	}

	if updateTaskComponentReq.Name != "" {
		oldTaskComponent.Name = updateTaskComponentReq.Name
	}

	oldTaskComponent.Data = updateTaskComponentReq.Data

	err = adapter.TaskComponentUpdate(oldTaskComponent)
	if err != nil {
		return nil, err
	}

	return oldTaskComponent, nil
}

func DeleteTaskComponent(id string) error {

	taskComponent, err := adapter.TaskComponentGet(id)
	if err != nil {
		return err
	}

	err = adapter.TaskComponentDelete(taskComponent)
	if err != nil {
		return err
	}

	return nil
}
