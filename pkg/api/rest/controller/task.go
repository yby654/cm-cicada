package controller

import (
	"net/http"

	"github.com/cloud-barista/cm-cicada/internal/service"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/common"
	"github.com/labstack/echo/v4"
)

// ListTaskFromTaskGroup godoc
//
//	@ID		list-task-from-task-group
//	@Summary	List Task from Task Group
//	@Description	Get a task list from the task group.
//	@Tags	[Workflow]
//	@Accept	json
//	@Produce	json
//	@Param	wfId path string true "ID of the workflow."
//	@Param	tgId path string true "ID of the task group."
//	@Success	200	{object}	[]model.Task		"Successfully get a task list from the task group."
//	@Failure	400	{object}	common.ErrorResponse	"Sent bad request."
//	@Failure	500	{object}	common.ErrorResponse	"Failed to get a task list from the task group."
//	@Router	/workflow/{wfId}/task_group/{tgId}/task [get]
func ListTaskFromTaskGroup(c echo.Context) error {
	wfId := c.Param("wfId")
	if wfId == "" {
		return common.ReturnErrorMsg(c, "wfId is required.")
	}
	tgId := c.Param("tgId")
	if tgId == "" {
		return common.ReturnErrorMsg(c, "tgId is required.")
	}

	tasks, err := service.ListTaskFromTaskGroup(wfId, tgId)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	return c.JSONPretty(http.StatusOK, tasks, " ")
}

// GetTaskFromTaskGroup godoc
//
//	@ID		get-task-from-task-group
//	@Summary	Get Task from Task Group
//	@Description	Get the task from the task group.
//	@Tags		[Workflow]
//	@Accept		json
//	@Produce	json
//	@Param		wfId path string true "ID of the workflow."
//	@Param		tgId path string true "ID of the task group."
//	@Param		taskId path string true "ID of the task."
//	@Success	200	{object}	model.Task		"Successfully get the task from the task group."
//	@Failure	400	{object}	common.ErrorResponse	"Sent bad request."
//	@Failure	500	{object}	common.ErrorResponse	"Failed to get the task from the task group."
//	@Router		/workflow/{wfId}/task_group/{tgId}/task/{taskId} [get]
func GetTaskFromTaskGroup(c echo.Context) error {

	wfId := c.Param("wfId")
	if wfId == "" {
		return common.ReturnErrorMsg(c, "wfId is required.")
	}
	tgId := c.Param("tgId")
	if tgId == "" {
		return common.ReturnErrorMsg(c, "tgId is required.")
	}
	taskId := c.Param("taskId")
	if taskId == "" {
		return common.ReturnErrorMsg(c, "taskId is required.")
	}

	workflowService := &service.WorkflowService{}
	validationInput := service.WorkflowValidationInput{
		WorkflowID:  &wfId,
		TaskGroupID: &tgId,
		TaskID:      &taskId,
	}
	err := workflowService.ValidateWorkflowInfo(validationInput)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	task, err := service.GetTaskFromTaskGroup(wfId, tgId, taskId)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	return c.JSONPretty(http.StatusOK, task, " ")
}

// ListTask godoc
//
//	@ID		list-task
//	@Summary	List Task
//	@Description	Get a task list of the workflow.
//	@Tags		[Workflow]
//	@Accept		json
//	@Produce	json
//	@Param		wfId path string true "ID of the workflow."
//	@Success	200	{object}	[]model.Task		"Successfully get a task list."
//	@Failure	400	{object}	common.ErrorResponse	"Sent bad request."
//	@Failure	500	{object}	common.ErrorResponse	"Failed to get a task list."
//	@Router		/workflow/{wfId}/task [get]
func ListTask(c echo.Context) error {
	wfId := c.Param("wfId")
	if wfId == "" {
		return common.ReturnErrorMsg(c, "wfId is required.")
	}
	tasks, err := service.ListTask(wfId)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	return c.JSONPretty(http.StatusOK, tasks, " ")
}

// GetTask godoc
//
//	@ID		get-task
//	@Summary	Get Task
//	@Description	Get the task.
//	@Tags		[Workflow]
//	@Accept		json
//	@Produce	json
//	@Param		wfId path string true "ID of the workflow."
//	@Param		taskId path string true "ID of the task."
//	@Success	200	{object}	model.Task		"Successfully get the task."
//	@Failure	400	{object}	common.ErrorResponse	"Sent bad request."
//	@Failure	500	{object}	common.ErrorResponse	"Failed to get the task."
//	@Router		/workflow/{wfId}/task/{taskId} [get]
func GetTask(c echo.Context) error {
	wfId := c.Param("wfId")
	if wfId == "" {
		return common.ReturnErrorMsg(c, "wfId is required.")
	}
	taskId := c.Param("taskId")
	if taskId == "" {
		return common.ReturnErrorMsg(c, "taskId is required.")
	}
	task, err := service.GetTask(wfId, taskId)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	return c.JSONPretty(http.StatusOK, task, " ")
}

// GetTaskDirectly godoc
//
//	@ID		get-task-directly
//	@Summary	Get Task Directly
//	@Description	Get the task directly.
//	@Tags		[Workflow]
//	@Accept		json
//	@Produce	json
//	@Param		taskId path string true "ID of the task."
//	@Success	200	{object}	model.Task	"Successfully get the task."
//	@Failure	400	{object}	common.ErrorResponse	"Sent bad request."
//	@Failure	500	{object}	common.ErrorResponse	"Failed to get the task."
//	@Router		/task/{taskId} [get]
func GetTaskDirectly(c echo.Context) error {
	taskId := c.Param("taskId")
	task, err := service.GetTaskDirectly(taskId)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	// tgDB, err := adapter.TaskGroupGet(tDB.TaskGroupID)
	// if err != nil {
	// 	return common.ReturnErrorMsg(c, err.Error())
	// }

	// workflow, err := adapter.WorkflowGet(tgDB.WorkflowID)
	// if err != nil {
	// 	return common.ReturnErrorMsg(c, err.Error())
	// }

	// for _, tg := range workflow.Data.TaskGroups {
	// 	if tg.ID == tgDB.ID {
	// 		for _, task := range tg.Tasks {
	// 			if task.ID == taskId {
	// 				return c.JSONPretty(http.StatusOK, model.TaskDirectly{
	// 					ID:            task.ID,
	// 					WorkflowID:    tDB.WorkflowID,
	// 					TaskGroupID:   tDB.TaskGroupID,
	// 					Name:          task.Name,
	// 					TaskComponent: task.TaskComponent,
	// 					RequestBody:   task.RequestBody,
	// 					PathParams:    task.PathParams,
	// 					QueryParams:   task.QueryParams,
	// 					Extra:         task.Extra,
	// 					Dependencies:  task.Dependencies,
	// 				}, " ")
	// 			}
	// 		}
	// 	}
	// }
	return c.JSONPretty(http.StatusOK, task, "")
	// return common.ReturnErrorMsg(c, "task not found.")
}

// GetTaskLogs godoc
//
//	@ID			get-task-logs
//	@Summary	Get Task Logs
//	@Description	Get the task Logs.
//	@Tags	[Workflow]
//	@Accept	json
//	@Produce	json
//	@Param	wfId path string true "ID of the workflow."
//	@Param	wfRunId path string true "ID of the workflowRunId."
//	@Param	taskId path string true "ID of the task."
//	@Param	taskTryNum path string true "ID of the taskTryNum."
//	@Success	200	{object}	model.TaskLog	"Successfully get the task Logs."
//	@Failure	400	{object}	common.ErrorResponse	"Sent bad request."
//	@Failure	500	{object}	common.ErrorResponse	"Failed to get the task Logs."
//	@Router	 /workflow/{wfId}/workflowRun/{wfRunId}/task/{taskId}/taskTryNum/{taskTryNum}/logs [get]
func GetTaskLogs(c echo.Context) error {
	wfId := c.Param("wfId")
	wfRunId := c.Param("wfRunId")
	taskId := c.Param("taskId")
	taskTryNum := c.Param("taskTryNum")

	taskLog, err := service.GetTaskLogs(wfId, wfRunId, taskId, taskTryNum)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	return c.JSONPretty(http.StatusOK, taskLog, " ")
}

// GetTaskLogDownload godoc
//
//	@ID			get-task-logs-download
//	@Summary	Download Task Logs
//	@Description	Download the task logs as a file.
//	@Tags		[Workflow]
//	@Accept		json
//	@Produce	text/plain
//	@Param		wfId path string true "ID of the workflow."
//	@Param		wfRunId path string true "ID of the workflowRunId."
//	@Param		taskId path string true "ID of the task."
//	@Param		taskTryNum path string true "ID of the taskTryNum."
//	@Success	200 {file} file "Log file downloaded successfully."
//	@Failure	400 {object} common.ErrorResponse "Sent bad request."
//	@Failure	500 {object} common.ErrorResponse "Failed to get the task Logs."
//	@Router		/workflow/{wfId}/workflowRun/{wfRunId}/task/{taskId}/taskTryNum/{taskTryNum}/logs/download [get]
func GetTaskLogDownload(c echo.Context) error {
	wfId := c.Param("wfId")
	wfRunId := c.Param("wfRunId")
	taskId := c.Param("taskId")
	taskTryNum := c.Param("taskTryNum")

	filename, content, err := service.GetTaskLogDownload(wfId, wfRunId, taskId, taskTryNum)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	c.Response().Header().Set("Content-Disposition", "attachment; filename="+filename)
	c.Response().Header().Set("Content-Type", "text/plain")
	return c.Blob(http.StatusOK, "text/plain", content)
}
