package controller

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/cloud-barista/cm-cicada/internal/adapter"
	"github.com/cloud-barista/cm-cicada/internal/lib/airflow"
	"github.com/cloud-barista/cm-cicada/internal/service"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/common"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
	"github.com/jollaman999/utils/logger"
	"github.com/labstack/echo/v4"
	"github.com/mitchellh/mapstructure"
)

// CreateWorkflowRun godoc
//
//	@ID		create-workflow-run
//	@Summary	Create WorkflowRun (DB 저장)
//	@Description	Airflow에서 DAG 종료 시 POST로 보낸 workflow run 데이터를 DB에 저장. 실행내역 전부 INSERT.
//	@Tags		[Workflow]
//	@Accept		json
//	@Produce	json
//	@Param		wfId path string true "ID of the workflow."
//	@Param		body body model.CreateWorkflowRunReq true "Workflow run data (Airflow callback payload)"
//	@Success	200	{object}	model.WorkflowRun	"Successfully saved workflow run"
//	@Failure	400	{object}	common.ErrorResponse	"Sent bad request."
//	@Failure	500	{object}	common.ErrorResponse	"Failed to save workflow run"
//	@Router		/workflow/{wfId}/runs [post]
func CreateWorkflowRun(c echo.Context) error {
	wfId := c.Param("wfId")
	if wfId == "" {
		return common.ReturnErrorMsg(c, "Please provide the wfId.")
	}

	data, err := common.GetJSONRawBody(c)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	var req model.CreateWorkflowRunReq
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		DecodeHook: mapstructure.ComposeDecodeHookFunc(service.ToTimeHookFunc()),
		Result:     &req,
	})
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}
	if err = decoder.Decode(data); err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	if req.WorkflowID == "" {
		req.WorkflowID = wfId
	}

	resp, err := service.CreateWorkflowRun(req)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}
	return c.JSONPretty(http.StatusOK, resp, " ")
}

// GetWorkflowRuns godoc
//
//	@ID			get-workflow-runs
//	@Summary	Get workflowRuns
//	@Description	Get DAG runs from Airflow API.
//	@Tags	[Workflow]
//	@Accept	json
//	@Produce	json
//	@Param	wfId path string true "ID of the workflow."
//	@Success	200	{object}	model.WorkflowRun		"Successfully get the workflowRuns."
//	@Failure	400	{object}	common.ErrorResponse	"Sent bad request."
//	@Failure	500	{object}	common.ErrorResponse	"Failed to get the workflowRuns."
//	@Router	 /workflow/{wfId}/runs [get]
func GetWorkflowRuns(c echo.Context) error {
	wfId := c.Param("wfId")
	if wfId == "" {
		return common.ReturnErrorMsg(c, "Please provide the wfId.")
	}
	workflowRunList, err := service.GetWorkflowRunByWorkflowID(wfId)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}
	// client, err := airflow.GetClient()
	// if err != nil {
	// 	return common.ReturnErrorMsg(c, err.Error())
	// }

	// runList, err := client.GetDAGRuns(wfId)
	// if err != nil {
	// 	return common.ReturnErrorMsg(c, "Failed to get the workflow runs: "+err.Error())
	// }

	return c.JSONPretty(http.StatusOK, workflowRunList, " ")
}

// GetTaskInstances godoc
//
//	@ID			get-task-instances
//	@Summary	Get taskInstances
//	@Description	Get the task Logs.
//	@Tags	[Workflow]
//	@Accept	json
//	@Produce	json
//	@Param	wfId path string true "ID of the workflow."
//	@Param	wfRunId path string true "ID of the workflow."
//	@Success	200	{object}	model.TaskInstance		"Successfully get the taskInstances."
//	@Failure	400	{object}	common.ErrorResponse	"Sent bad request."
//	@Failure	500	{object}	common.ErrorResponse	"Failed to get the taskInstances."
//	@Router	 /workflow/{wfId}/workflowRun/{wfRunId}/taskInstances [get]
func GetTaskInstances(c echo.Context) error {
	wfId := c.Param("wfId")
	if wfId == "" {
		return common.ReturnErrorMsg(c, "Please provide the wfId.")
	}
	wfRunId := c.Param("wfRunId")
	if wfRunId == "" {
		return common.ReturnErrorMsg(c, "Please provide the wfRunId.")
	}
	client, err := airflow.GetClient()
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}
	runList, err := client.GetTaskInstances(common.UrlDecode(wfId), common.UrlDecode(wfRunId))
	if err != nil {
		return common.ReturnErrorMsg(c, "Failed to get the taskInstances: "+err.Error())
	}
	var taskInstances []model.TaskInstance
	layout := time.RFC3339Nano

	workflow, err := service.GetWorkflowFromDB(wfId)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	for _, taskInstance := range *runList.TaskInstances {
		taskDBInfo, err := adapter.TaskGetByWorkflowIDAndName(taskInstance.GetDagId(), taskInstance.GetTaskId())
		if err != nil {
			return common.ReturnErrorMsg(c, "Failed to get the taskInstances: "+err.Error())
		}
		taskId := &taskDBInfo.ID
		executionDate, err := time.Parse(layout, taskInstance.GetExecutionDate())
		if err != nil {
			fmt.Println("Error parsing execution date:", err)
			continue
		}
		startDate, err := time.Parse(layout, taskInstance.GetExecutionDate())
		if err != nil {
			fmt.Println("Error parsing start date:", err)
			continue
		}
		endDate, err := time.Parse(layout, taskInstance.GetExecutionDate())
		if err != nil {
			fmt.Println("Error parsing end date:", err)
			continue
		}

		var isSoftwareMigrationTask bool
		var executionID string
		for _, tg := range workflow.Data.TaskGroups {
			for _, task := range tg.Tasks {
				if strings.Contains(task.TaskComponent, "grasshopper") &&
					strings.Contains(task.TaskComponent, "software") &&
					strings.Contains(task.TaskComponent, "migration") &&
					task.ID == *taskId {
					isSoftwareMigrationTask = true

					// software migration task인 경우 xcom에서 execution_id 조회
					xcomData, err := client.GetXComValue(
						taskInstance.GetDagId(),
						taskInstance.GetDagRunId(),
						taskInstance.GetTaskId(),
						"return_value",
					)
					if err != nil {
						logger.Println(logger.WARN, false,
							"Failed to get xcom data for task: "+taskInstance.GetTaskId()+" (Error: "+err.Error()+")")
					} else if xcomData != nil {
						if execID, ok := xcomData["execution_id"].(string); ok {
							executionID = execID
						}
					}
					break
				}
			}
		}

		taskInfo := model.TaskInstance{
			WorkflowID:                   taskInstance.DagId,
			WorkflowRunID:                taskInstance.GetDagRunId(),
			TaskID:                       *taskId,
			TaskName:                     taskInstance.GetTaskId(),
			State:                        string(taskInstance.GetState()),
			ExecutionDate:                executionDate,
			StartDate:                    startDate,
			EndDate:                      endDate,
			DurationDate:                 float64(taskInstance.GetDuration()),
			TryNumber:                    int(taskInstance.GetTryNumber()),
			IsSoftwareMigrationTask:      isSoftwareMigrationTask,
			SoftwareMigrationExecutionID: executionID,
		}
		taskInstances = append(taskInstances, taskInfo)
	}
	return c.JSONPretty(http.StatusOK, taskInstances, " ")
}

// ClearTaskInstances godoc
//
//	@ID			clear-task-instances
//	@Summary	Clear taskInstances
//	@Description	Clear the task Instance.
//	@Tags	[Workflow]
//	@Accept	json
//	@Produce	json
//	@Param	wfId path string true "ID of the workflow."
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
			service.ToTimeHookFunc()),
		Result: &taskClearOption,
	})
	if err != nil {
		return err
	}
	err = decoder.Decode(data)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}
	var taskNameList []string
	for _, taskId := range taskClearOption.TaskIds {
		taskInfo, err := adapter.TaskGet(taskId)
		if err != nil {
			return fmt.Errorf("failed to get task info for ID %s: %w", taskId, err)
		}
		taskNameList = append(taskNameList, taskInfo.Name)
	}
	taskClearOption.TaskIds = taskNameList
	if err := common.ValidateTaskClearOptions(taskClearOption); err != nil {
		fmt.Printf("옵션 검증 실패: %v\n", err)
		return common.ReturnErrorMsg(c, err.Error())
	}
	wfId := c.Param("wfId")
	if wfId == "" {
		return common.ReturnErrorMsg(c, "Please provide the wfId.")
	}
	wfRunId := c.Param("wfRunId")
	if wfRunId == "" {
		return common.ReturnErrorMsg(c, "Please provide the wfRunId.")
	}

	TaskInstanceReferences := make([]model.TaskInstanceReference, 0)
	client, err := airflow.GetClient()
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}
	clearList, err := client.ClearTaskInstance(wfId, common.UrlDecode(wfRunId), taskClearOption)
	if err != nil {
		return common.ReturnErrorMsg(c, "Failed to get the taskInstances: "+err.Error())
	}
	logger.Println(logger.DEBUG, false, "clearList 요청 내용 : {} ", &clearList)
	if clearList.TaskInstances == nil || len(*clearList.TaskInstances) == 0 {
		logger.Println(logger.DEBUG, false, "TaskInstances is nil or empty")

	}
	for _, taskInstance := range *clearList.TaskInstances {
		taskDBInfo, err := adapter.TaskGetByWorkflowIDAndName(taskInstance.GetDagId(), taskInstance.GetTaskId())
		if err != nil {
			return common.ReturnErrorMsg(c, "Failed to get the taskInstances: "+err.Error())
		}
		taskId := &taskDBInfo.ID
		taskInfo := model.TaskInstanceReference{
			WorkflowID:    taskInstance.DagId,
			WorkflowRunID: taskInstance.DagRunId,
			TaskId:        taskId,
			TaskName:      taskInstance.GetTaskId(),
			ExecutionDate: taskInstance.ExecutionDate,
		}
		logger.Println(logger.DEBUG, false, "TaskInstanceReferences  ", TaskInstanceReferences)
		TaskInstanceReferences = append(TaskInstanceReferences, taskInfo)
	}
	logger.Println(logger.DEBUG, false, "TaskInstanceReferences ", TaskInstanceReferences)

	return c.JSONPretty(http.StatusOK, TaskInstanceReferences, " ")
}

// GetWorkflowStatus godoc
//
//	@ID		get-WorkflowStatus
//	@Summary	Get WorkflowStatus
//	@Description	Get the WorkflowStatus.
//	@Tags		[Workflow]
//	@Accept		json
//	@Produce	json
//	@Param		wfId path string true "wfId of the workflow"
//	@Success	200	{object}	[]model.WorkflowStatus		"Successfully get the WorkflowVersion."
//	@Failure	400	{object}	common.ErrorResponse	"Sent bad request."
//	@Failure	500	{object}	common.ErrorResponse	"Failed to get the WorkflowVersion."
//	@Router		/workflow/{wfId}/status [get]
func GetWorkflowStatus(c echo.Context) error {
	wfId := c.Param("wfId")
	if wfId == "" {
		return common.ReturnErrorMsg(c, "Please provide the wfId.")
	}
	client, err := airflow.GetClient()
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}
	enumStatus := client.GetAllowedDagStateEnumValues()
	var statusList []model.WorkflowStatus
	for _, v := range enumStatus {

		resp, err := client.GetDagStatus(wfId, string(*v.Ptr()))
		if err != nil {
			logger.Println(logger.ERROR, false,
				"AIRFLOW: Error occurred while getting DAGRuns. (Error: "+err.Error()+").")
		}
		statusList = append(statusList, model.WorkflowStatus{
			State: string(*v.Ptr()),
			Count: len(*resp.DagRuns),
		})
	}

	return c.JSONPretty(http.StatusOK, statusList, " ")
}
