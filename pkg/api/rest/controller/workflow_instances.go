package controller

import (
	"net/http"

	"github.com/cloud-barista/cm-cicada/pkg/api/rest/common"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/service"
	"github.com/labstack/echo/v4"
	"github.com/mitchellh/mapstructure"
)

// GetTaskInstances godoc
//
//	@ID			get-task-instances
//	@Summary	List Task Instances
//	@Description	Get the task Logs.
//	@Tags	[Workflow Execution]
//	@Accept	json
//	@Produce	json
//	@Param	wfId path string true "DB workflow ID."
//	@Param	wfRunId path string true "DB workflow ID."
//	@Success	200	{object}	model.TaskInstance		"Successfully get the taskInstances."
//	@Failure	400	{object}	common.ErrorResponse	"Sent bad request."
//	@Failure	500	{object}	common.ErrorResponse	"Failed to get the taskInstances."
//	@Router	 /workflow/{wfId}/workflowRun/{wfRunId}/taskInstances [get]
func GetTaskInstances(c echo.Context) error {
	wfId, err := requireParam(c, "wfId", "wfId")
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}
	wfRunId, err := requireParam(c, "wfRunId", "wfRunId")
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	svc := service.NewWorkflowRuntimeService()
	taskInstances, err := svc.GetTaskInstances(wfId, wfRunId)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	return c.JSONPretty(http.StatusOK, taskInstances, " ")
}

// ClearTaskInstances godoc
//
//	@ID			clear-task-instances
//	@Summary	Clear Task Instances
//	@Description	Clear the task Instance.
//	@Tags	[Workflow Execution]
//	@Accept	json
//	@Produce	json
//	@Param	wfId path string true "DB workflow ID."
//	@Param	wfRunId path string true "ID of the wfRunId."
//
// @Param		request body 	model.TaskClearOption true "Workflow content"
// @Success	200	{object}	model.TaskInstanceReference		"Successfully clear the taskInstances."
// @Failure	400	{object}	common.ErrorResponse	"Sent bad request."
// @Failure	500	{object}	common.ErrorResponse	"Failed to clear the taskInstances."
// @Router	 /workflow/{wfId}/workflowRun/{wfRunId}/range [post]
func ClearTaskInstances(c echo.Context) error {
	var taskClearOption model.TaskClearOption

	data, err := common.GetJSONRawBody(c)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Metadata: nil,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			toTimeHookFunc()),
		Result: &taskClearOption,
	})
	if err != nil {
		return err
	}

	if err := decoder.Decode(data); err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	wfId, err := requireParam(c, "wfId", "wfId")
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}
	wfRunId, err := requireParam(c, "wfRunId", "wfRunId")
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	if err := common.ValidateTaskClearOptions(taskClearOption); err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	svc := service.NewWorkflowRuntimeService()
	references, err := svc.ClearTaskInstances(wfId, wfRunId, taskClearOption)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	return c.JSONPretty(http.StatusOK, references, " ")
}

// RerunTaskGroups godoc
//
//	@ID			rerun-task-groups
//	@Summary	Rerun Task Groups
//	@Description	Re-run whole task groups of a workflow run. Every task of the named groups is cleared, so the scheduler runs those groups again while the rest of the run keeps its state. Each entry of taskGroups takes a task group ID or name. Pass includeDownstream to re-run everything that depends on them as well, which re-runs a group and every group after it. A run that already finished is moved back to running, so the groups really do run again. Pass dryRun to get the list of tasks that would be re-run without touching the run.
//	@Tags	[Workflow Execution]
//	@Accept	json
//	@Produce	json
//	@Param	wfId path string true "DB workflow ID."
//	@Param	wfRunId path string true "ID of the wfRunId."
//
// @Param		request body 	model.TaskGroupRerunOption true "Task groups to re-run"
// @Success	200	{object}	model.TaskInstanceReference		"Successfully cleared the task groups."
// @Failure	400	{object}	common.ErrorResponse	"Sent bad request."
// @Failure	500	{object}	common.ErrorResponse	"Failed to clear the task groups."
// @Router	 /workflow/{wfId}/workflowRun/{wfRunId}/task_group_range [post]
func RerunTaskGroups(c echo.Context) error {
	var rerunOption model.TaskGroupRerunOption

	data, err := common.GetJSONRawBody(c)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Metadata: nil,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			toTimeHookFunc()),
		Result: &rerunOption,
	})
	if err != nil {
		return err
	}

	if err := decoder.Decode(data); err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	wfId, err := requireParam(c, "wfId", "wfId")
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}
	wfRunId, err := requireParam(c, "wfRunId", "wfRunId")
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	if err := common.ValidateTaskGroupRerunOptions(rerunOption); err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	svc := service.NewWorkflowRuntimeService()
	references, err := svc.RerunTaskGroups(wfId, wfRunId, rerunOption)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	return c.JSONPretty(http.StatusOK, references, " ")
}
