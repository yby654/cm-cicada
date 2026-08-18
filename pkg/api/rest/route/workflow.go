package route

import (
	"strings"

	"github.com/cloud-barista/cm-cicada/common"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/controller"
	"github.com/labstack/echo/v4"
)

func Workflow(e *echo.Echo) {
	api := e.Group("/" + strings.ToLower(common.ShortModuleName))

	// Workflow
	api.POST("/workflow", controller.CreateWorkflow)
	api.GET("/workflow", controller.ListWorkflow)
	api.GET("/workflow/:wfId", controller.GetWorkflow)
	api.PUT("/workflow/:wfId", controller.UpdateWorkflow)
	api.DELETE("/workflow/:wfId", controller.DeleteWorkflow)
	api.GET("/workflow/name/:wfName", controller.GetWorkflowByName)
	api.POST("/workflow/:wfId/clone", controller.CloneWorkflow)

	// Workflow version
	api.GET("/workflow/:wfId/version", controller.ListWorkflowVersion)
	api.GET("/workflow/:wfId/version/:verId", controller.GetWorkflowVersion)
	api.POST("/workflow/:wfId/version/:versionNo/rollback", controller.RollbackWorkflow)

	// Task group and task
	api.GET("/workflow/:wfId/task_group", controller.ListTaskGroup)
	api.GET("/workflow/:wfId/task_group/:tgId", controller.GetTaskGroup)
	api.GET("/workflow/:wfId/task_group/:tgId/task", controller.ListTaskFromTaskGroup)
	api.GET("/workflow/:wfId/task_group/:tgId/task/:taskId", controller.GetTaskFromTaskGroup)
	api.GET("/workflow/:wfId/task", controller.ListTask)
	api.GET("/workflow/:wfId/task/:taskId", controller.GetTask)
	api.GET("/task_group/:tgId", controller.GetTaskGroupDirectly)
	api.GET("/task/:taskId", controller.GetTaskDirectly)

	// Scheduling
	api.POST("/workflow/:wfId/schedule", controller.ScheduleWorkflow)
	api.GET("/workflow/:wfId/schedule", controller.GetWorkflowSchedule)
	api.DELETE("/workflow/:wfId/schedule", controller.CancelWorkflowSchedule)

	// Execution
	api.POST("/workflow/:wfId/run", controller.RunWorkflow)
	api.GET("/workflow/:wfId/status", controller.GetWorkflowStatus)
	api.GET("/workflow/:wfId/runs", controller.GetWorkflowRuns)
	api.GET("/workflow/:wfId/workflowRun/:wfRunId/taskInstances", controller.GetTaskInstances)
	api.POST("/workflow/:wfId/workflowRun/:wfRunId/range", controller.ClearTaskInstances)
	api.POST("/workflow/:wfId/workflowRun/:wfRunId/task_group_range", controller.RerunTaskGroups)

	// Logs
	api.GET("/workflow/:wfId/workflowRun/:wfRunId/task/:taskId/taskTryNum/:taskTryNum/logs", controller.GetTaskLogs)
	api.GET("/workflow/:wfId/workflowRun/:wfRunId/task/:taskId/taskTryNum/:taskTryNum/logs/download", controller.GetTaskLogDownload)
	api.GET("/workflow/:wfId/eventlogs", controller.GetEventLogs)
	api.GET("/importErrors", controller.GetImportErrors)

	// Utilities
	api.POST("/run_script", controller.RunScript)
	api.POST("/sleep_time", controller.SleepTime)
}
