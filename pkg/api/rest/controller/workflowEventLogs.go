package controller

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/cloud-barista/cm-cicada/internal/adapter"
	"github.com/cloud-barista/cm-cicada/internal/lib/airflow"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/common"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
	"github.com/labstack/echo/v4"
)

// GetEventLogs godoc
//
//	@ID				get-event-logs
//	@Summary		Get Eventlog
//	@Description	Get Eventlog.
//	@Tags		[Workflow]
//	@Accept		json
//	@Produce	json
//	@Param		wfId path string true "ID of the workflow."
//	@Param		wfRunId query string false "ID of the workflow run."
//	@Param		taskId query string false "ID of the task."
//	@Success	200	{object}	[]model.EventLog			"Successfully get the workflow."
//	@Failure	400	{object}	common.ErrorResponse	"Sent bad request."
//	@Failure	500	{object}	common.ErrorResponse	"Failed to get the workflow."
//	@Router	/workflow/{wfId}/eventlogs [get]
func GetEventLogs(c echo.Context) error {
	wfId := c.Param("wfId")
	if wfId == "" {
		return common.ReturnErrorMsg(c, "Please provide the wfId.")
	}

	var wfRunId, taskId, taskName string

	if c.QueryParam("wfRunId") != "" {
		wfRunId = c.QueryParam("wfRunId")
	}
	if c.QueryParam("taskId") != "" {
		taskId = c.QueryParam("taskId")
		taskDBInfo, err := adapter.TaskGet(taskId)
		if err != nil {
			return common.ReturnErrorMsg(c, "Failed to get the taskInstances: "+err.Error())
		}
		taskName = taskDBInfo.Name
	}
	var eventLogs model.EventLogs
	client, err := airflow.GetClient()
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}
	logs, err := client.GetEventLogs(wfId, wfRunId, taskName)
	if err != nil {
		return common.ReturnErrorMsg(c, "Failed to get the taskInstances: "+err.Error())
	}
	err = json.Unmarshal(logs, &eventLogs)
	if err != nil {
		fmt.Println(err)
	}
	var logList []model.EventLog
	for _, eventlog := range eventLogs.EventLogs {
		var taskID, RunId string
		if eventlog.TaskID != "" {
			taskDBInfo, err := adapter.TaskGetByWorkflowIDAndName(wfId, eventlog.TaskID)
			if err != nil {
				return common.ReturnErrorMsg(c, "Failed to get the taskInstances: "+err.Error())
			}
			taskID = taskDBInfo.ID
		}
		eventlog.WorkflowID = wfId
		if eventlog.RunID != "" {
			RunId = eventlog.RunID
		}

		log := model.EventLog{
			WorkflowID:    eventlog.WorkflowID,
			WorkflowRunID: RunId,
			TaskID:        taskID,
			TaskName:      eventlog.TaskID,
			Extra:         eventlog.Extra,
			Event:         eventlog.Event,
			When:          eventlog.When,
		}
		logList = append(logList, log)
	}
	return c.JSONPretty(http.StatusOK, logList, " ")
}
