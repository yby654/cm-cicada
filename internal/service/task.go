package service

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/cloud-barista/cm-cicada/internal/adapter"
	"github.com/cloud-barista/cm-cicada/internal/domain"
	"github.com/cloud-barista/cm-cicada/internal/lib/airflow"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/common"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
)

func ListTaskFromTaskGroup(wfId string, tgId string) ([]domain.Task, error) {
	workflow, err := adapter.WorkflowGet(wfId)
	if err != nil {
		return nil, err
	}

	var tasks []domain.Task
	for _, tg := range workflow.Data.TaskGroups {
		if tg.ID == tgId {
			tasks = append(tasks, tg.Tasks...)
			break
		}
	}

	return tasks, nil
}

func GetTaskFromTaskGroup(wfId string, tgId string, taskId string) (*domain.Task, error) {

	workflow, err := adapter.WorkflowGet(wfId)
	if err != nil {
		return nil, err
	}

	for _, tg := range workflow.Data.TaskGroups {
		if tg.ID != tgId {
			continue
		}

		for _, task := range tg.Tasks {
			if task.ID == taskId {
				return &task, nil
			}
		}
		break
	}

	return nil, err
}

func ListTask(wfId string) ([]domain.Task, error) {

	workflow, err := adapter.WorkflowGet(wfId)
	if err != nil {
		return nil, err
	}

	var tasks []domain.Task
	for _, tg := range workflow.Data.TaskGroups {
		tasks = append(tasks, tg.Tasks...)
	}

	return tasks, nil
}

func GetTask(wfId string, taskId string) (*domain.Task, error) {

	workflow, err := adapter.WorkflowGet(wfId)
	if err != nil {
		return nil, err
	}

	for _, tg := range workflow.Data.TaskGroups {
		for _, task := range tg.Tasks {
			if task.ID == taskId {
				return &task, nil
			}
		}
	}

	return nil, err
}

func GetTaskDirectly(taskId string) (*domain.Task, error) {

	tDB, err := adapter.TaskGet(taskId)
	if err != nil {
		return nil, err
	}

	return tDB, nil
}

func GetTaskLogs(wfId string, wfRunId string, taskId string, taskTryNum string) (*model.TaskLog, error) {

	taskInfo, err := adapter.TaskGet(taskId)
	if err != nil {
		return nil, err
	}

	taskTryNumToInt, err := strconv.Atoi(taskTryNum)
	if err != nil {
		return nil, err
	}
	client, err := airflow.GetClient()
	if err != nil {
		return nil, err
	}

	logs, err := client.GetTaskLogs(wfId, common.UrlDecode(wfRunId), taskInfo.Name, taskTryNumToInt)
	if err != nil {
		return nil, errors.New("Failed to get the workflow logs: " + err.Error())
	}

	taskLog := model.TaskLog{
		Content: *logs.Content,
	}
	return &taskLog, nil
}

func GetTaskLogDownload(wfId string, wfRunId string, taskId string, taskTryNum string) (string, []byte, error) {

	taskInfo, err := adapter.TaskGet(taskId)
	if err != nil {
		return "", nil, err
	}

	taskTryNumToInt, err := strconv.Atoi(taskTryNum)
	if err != nil {
		return "", nil, err
	}

	client, err := airflow.GetClient()
	if err != nil {
		return "", nil, err
	}

	logs, err := client.GetTaskLogs(wfId, common.UrlDecode(wfRunId), taskInfo.Name, taskTryNumToInt)
	if err != nil {
		return "", nil, err
	}

	filename := fmt.Sprintf("%s_%s_%s.log", wfId, wfRunId, taskInfo.Name)
	return filename, []byte(*logs.Content), nil
}
