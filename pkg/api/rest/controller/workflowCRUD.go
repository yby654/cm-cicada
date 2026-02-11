package controller

import (
	"fmt"
	"net/http"

	"github.com/cloud-barista/cm-cicada/internal/adapter"
	"github.com/cloud-barista/cm-cicada/internal/domain"
	"github.com/cloud-barista/cm-cicada/internal/lib/airflow"
	"github.com/cloud-barista/cm-cicada/internal/service"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/common"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
	"github.com/google/uuid"
	"github.com/jollaman999/utils/logger"
	"github.com/labstack/echo/v4"
	"github.com/mitchellh/mapstructure"
)

// CreateWorkflow godoc
//
//	@ID		create-workflow
//	@Summary	Create Workflow
//	@Description	Create a workflow.
//	@Tags		[Workflow]
//	@Accept		json
//	@Produce	json
//	@Param		request body 	model.CreateWorkflowReq true "Workflow content"
//	@Success	200	{object}	model.WorkflowTemplate	"Successfully create the workflow."
//	@Failure	400	{object}	common.ErrorResponse	"Sent bad request."
//	@Failure	500	{object}	common.ErrorResponse	"Failed to create workflow."
//	@Router		/workflow [post]
func CreateWorkflow(c echo.Context) error {
	var createWorkflowReq model.CreateWorkflowReq

	data, err := common.GetJSONRawBody(c)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Metadata: nil,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			service.ToTimeHookFunc()),
		Result: &createWorkflowReq,
	})
	if err != nil {
		return err
	}

	err = decoder.Decode(data)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	if createWorkflowReq.Name == "" {
		return common.ReturnErrorMsg(c, "Please provide the name.")
	}

	var specVersion = model.WorkflowSpecVersion_LATEST
	if createWorkflowReq.SpecVersion != "" {
		specVersion = createWorkflowReq.SpecVersion
	}

	workflowData, err := service.CreateDataReqToData(specVersion, createWorkflowReq.Data)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	var workflow domain.Workflow
	workflow.ID = uuid.New().String()
	workflow.SpecVersion = specVersion
	workflow.Name = createWorkflowReq.Name
	workflow.Data = workflowData

	var success bool
	_, err = adapter.WorkflowCreate(&workflow)
	if err != nil {
		{
			return common.ReturnErrorMsg(c, err.Error())
		}
	}
	defer func() {
		if !success {
			_ = adapter.WorkflowDelete(&workflow)
		}
	}()

	client, err := airflow.GetClient()
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	err = client.CreateDAG(&workflow)
	if err != nil {
		return common.ReturnErrorMsg(c, "Failed to create the workflow. (Error:"+err.Error()+")")
	}

	for _, tg := range workflow.Data.TaskGroups {
		_, err = adapter.TaskGroupCreate(&domain.TaskGroup{
			ID:                tg.ID,
			Name:              tg.Name,
			WorkflowVersionID: workflow.ID,
		})
		if err != nil {
			return common.ReturnErrorMsg(c, err.Error())
		}

		for _, t := range tg.Tasks {
			_, err = adapter.TaskCreate(&domain.Task{
				ID:          t.ID,
				Name:        t.Name,
				WorkflowID:  workflow.ID,
				TaskGroupID: tg.ID,
			})
			if err != nil {
				return common.ReturnErrorMsg(c, err.Error())
			}
		}
	}
	success = true

	return c.JSONPretty(http.StatusOK, workflow, " ")
}

// GetWorkflow godoc
//
//	@ID		get-workflow
//	@Summary	Get Workflow
//	@Description	Get the workflow.
//	@Tags		[Workflow]
//	@Accept		json
//	@Produce	json
//	@Param		wfId path string true "ID of the workflow."
//	@Success	200	{object}	model.Workflow		"Successfully get the workflow."
//	@Failure	400	{object}	common.ErrorResponse	"Sent bad request."
//	@Failure	500	{object}	common.ErrorResponse	"Failed to get the workflow."
//	@Router		/workflow/{wfId} [get]
func GetWorkflow(c echo.Context) error {
	wfId := c.Param("wfId")
	if wfId == "" {
		return common.ReturnErrorMsg(c, "Please provide the wfId.")
	}

	workflow, err := service.GetWorkflowFromDB(wfId)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	client, err := airflow.GetClient()
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	_, err = client.GetDAG(wfId)
	if err != nil {
		return common.ReturnErrorMsg(c, "Failed to get the workflow from the airflow server.")
	}

	return c.JSONPretty(http.StatusOK, workflow, " ")
}

// GetWorkflowByName godoc
//
//	@ID		get-workflow-by-name
//	@Summary	Get Workflow by Name
//	@Description	Get the workflow by name.
//	@Tags		[Workflow]
//	@Accept		json
//	@Produce	json
//	@Param		wfName path string true "Name of the workflow."
//	@Success	200	{object}	model.Workflow		"Successfully get the workflow."
//	@Failure	400	{object}	common.ErrorResponse	"Sent bad request."
//	@Failure	500	{object}	common.ErrorResponse	"Failed to get the workflow."
//	@Router		/workflow/name/{wfName} [get]
func GetWorkflowByName(c echo.Context) error {
	wfName := c.Param("wfName")
	if wfName == "" {
		return common.ReturnErrorMsg(c, "Please provide the wfName.")
	}

	workflow, err := adapter.WorkflowGetByName(wfName)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	for i, tg := range workflow.Data.TaskGroups {
		_, err = adapter.TaskGroupGetByWorkflowIDAndName(workflow.ID, tg.Name)
		if err != nil {
			logger.Println(logger.ERROR, true, err)
		}

		workflow.Data.TaskGroups[i].ID = tg.ID

		for j, t := range tg.Tasks {
			_, err = adapter.TaskGetByWorkflowIDAndName(workflow.ID, tg.Name)
			if err != nil {
				logger.Println(logger.ERROR, true, err)
			}

			workflow.Data.TaskGroups[i].Tasks[j].ID = t.ID
		}
	}

	client, err := airflow.GetClient()
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	_, err = client.GetDAG(workflow.ID)
	if err != nil {
		return common.ReturnErrorMsg(c, "Failed to get the workflow from the airflow server.")
	}

	return c.JSONPretty(http.StatusOK, workflow, " ")
}

// ListWorkflow godoc
//
//	@ID		list-workflow
//	@Summary	List Workflow
//	@Description	Get a workflow list.
//	@Tags		[Workflow]
//	@Accept		json
//	@Produce	json
//	@Param		name query string false "Name of the workflow"
//	@Param		page query string false "Page of the workflow list."
//	@Param		row query string false "Row of the workflow list."
//	@Success	200	{object}	[]model.Workflow	"Successfully get a workflow list."
//	@Failure	400	{object}	common.ErrorResponse	"Sent bad request."
//	@Failure	500	{object}	common.ErrorResponse	"Failed to get a workflow list."
//	@Router		/workflow [get]
func ListWorkflow(c echo.Context) error {
	page, row, err := common.CheckPageRow(c)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	workflow := &domain.Workflow{
		Name: c.QueryParam("name"),
	}

	workflows, err := adapter.WorkflowGetList(workflow, page, row)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}
	fmt.Println(workflows)

	// for i, w := range *workflows {
	// 	for j, tg := range workflow.Data.TaskGroups {
	// 		_, err = adapter.TaskGroupGetByWorkflowIDAndName(w.ID, tg.Name)
	// 		if err != nil {
	// 			logger.Println(logger.ERROR, true, err)
	// 		}

	// 		(*workflows)[i].Data.TaskGroups[j].ID = tg.ID

	// 		for k, t := range tg.Tasks {
	// 			_, err = adapter.TaskGetByWorkflowIDAndName(w.ID, tg.Name)
	// 			if err != nil {
	// 				logger.Println(logger.ERROR, true, err)
	// 			}

	// 			(*workflows)[i].Data.TaskGroups[j].Tasks[k].ID = t.ID
	// 		}
	// 	}
	// }

	return c.JSONPretty(http.StatusOK, workflows, " ")
}

// RunWorkflow godoc
//
//	@ID		run-workflow
//	@Summary	Run Workflow
//	@Description	Run the workflow.
//	@Tags		[Workflow]
//	@Accept		json
//	@Produce	json
//	@Param		wfId path string true "ID of the workflow."
//	@Success	200	{object}	model.SimpleMsg		"Successfully run the workflow."
//	@Failure	400	{object}	common.ErrorResponse	"Sent bad request."
//	@Failure	500	{object}	common.ErrorResponse	"Failed to run the Workflow"
//	@Router		/workflow/{wfId}/run [post]
func RunWorkflow(c echo.Context) error {
	wfId := c.Param("wfId")
	if wfId == "" {
		return common.ReturnErrorMsg(c, "Please provide the id.")
	}

	workflow, err := adapter.WorkflowGet(wfId)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	client, err := airflow.GetClient()
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	_, err = client.RunDAG(workflow.ID)
	if err != nil {
		return common.ReturnInternalError(c, err, "Failed to run the workflow.")
	}

	return c.JSONPretty(http.StatusOK, model.SimpleMsg{Message: "success"}, " ")
}

// UpdateWorkflow godoc
//
//	@ID		update-workflow
//	@Summary	Update Workflow
//	@Description	Update the workflow content.
//	@Tags		[Workflow]
//	@Accept		json
//	@Produce	json
//	@Param		wfId path string true "ID of the workflow."
//	@Param		Workflow body 	model.CreateWorkflowReq true "Workflow to modify."
//	@Success	200	{object}	model.Workflow	"Successfully update the workflow"
//	@Failure	400	{object}	common.ErrorResponse	"Sent bad request."
//	@Failure	500	{object}	common.ErrorResponse	"Failed to update the workflow"
//	@Router		/workflow/{wfId} [put]
func UpdateWorkflow(c echo.Context) error {
	var updateWorkflowReq model.CreateWorkflowReq

	data, err := common.GetJSONRawBody(c)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Metadata: nil,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			service.ToTimeHookFunc()),
		Result: &updateWorkflowReq,
	})
	if err != nil {
		return err
	}

	err = decoder.Decode(data)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	wfId := c.Param("wfId")
	oldWorkflow, err := adapter.WorkflowGet(wfId)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	if updateWorkflowReq.Name != "" {
		oldWorkflow.Name = updateWorkflowReq.Name
	}

	var specVersion = model.WorkflowSpecVersion_LATEST
	if updateWorkflowReq.SpecVersion != "" {
		specVersion = updateWorkflowReq.SpecVersion
	}

	workflowData, err := service.CreateDataReqToData(specVersion, updateWorkflowReq.Data)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	// Remove old task groups and tasks from the database
	for _, tg := range oldWorkflow.Data.TaskGroups {
		taskGroup, err := adapter.TaskGroupGet(tg.ID)
		if err != nil {
			logger.Println(logger.ERROR, true, err)
		}
		err = adapter.TaskGroupDelete(taskGroup)
		if err != nil {
			logger.Println(logger.ERROR, true, err)
		}

		for _, t := range tg.Tasks {
			task, err := adapter.TaskGet(t.ID)
			if err != nil {
				logger.Println(logger.ERROR, true, err)
			}
			err = adapter.TaskDelete(task)
			if err != nil {
				logger.Println(logger.ERROR, true, err)
			}
		}
	}

	// Create task groups and tasks to the database
	for _, tg := range workflowData.TaskGroups {
		_, err = adapter.TaskGroupCreate(&domain.TaskGroup{
			ID:                tg.ID,
			Name:              tg.Name,
			WorkflowVersionID: wfId,
		})
		if err != nil {
			return common.ReturnErrorMsg(c, err.Error())
		}

		for _, t := range tg.Tasks {
			_, err = adapter.TaskCreate(&domain.Task{
				ID:          t.ID,
				Name:        t.Name,
				WorkflowID:  wfId,
				TaskGroupID: tg.ID,
			})
			if err != nil {
				return common.ReturnErrorMsg(c, err.Error())
			}
		}
	}

	oldWorkflow.Data = workflowData

	err = adapter.WorkflowUpdate(oldWorkflow)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	client, err := airflow.GetClient()
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	err = client.DeleteDAG(oldWorkflow.ID, true)
	if err != nil {
		return common.ReturnErrorMsg(c, "Failed to update the workflow. (Error:"+err.Error()+")")
	}

	err = client.CreateDAG(oldWorkflow)
	if err != nil {
		return common.ReturnErrorMsg(c, "Failed to update the workflow. (Error:"+err.Error()+")")
	}

	return c.JSONPretty(http.StatusOK, oldWorkflow, " ")
}

// DeleteWorkflow godoc
//
//	@ID		delete-workflow
//	@Summary	Delete Workflow
//	@Description	Delete the workflow.
//	@Tags		[Workflow]
//	@Accept		json
//	@Produce	json
//	@Param		wfId path string true "ID of the workflow."
//	@Success	200	{object}	model.SimpleMsg	"Successfully delete the workflow"
//	@Failure	400	{object}	common.ErrorResponse	"Sent bad request."
//	@Failure	500	{object}	common.ErrorResponse	"Failed to delete the workflow"
//	@Router		/workflow/{wfId} [delete]
func DeleteWorkflow(c echo.Context) error {
	wfId := c.Param("wfId")
	if wfId == "" {
		return common.ReturnErrorMsg(c, "Please provide the wfId.")
	}

	workflow, err := adapter.WorkflowGet(wfId)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	client, err := airflow.GetClient()
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	err = client.DeleteDAG(workflow.ID, false)
	if err != nil {
		logger.Println(logger.ERROR, true, "AIRFLOW: "+err.Error())
	}

	for _, tg := range workflow.Data.TaskGroups {
		taskGroup, err := adapter.TaskGroupGet(tg.ID)
		if err != nil {
			logger.Println(logger.ERROR, true, err)
		}
		err = adapter.TaskGroupDelete(taskGroup)
		if err != nil {
			logger.Println(logger.ERROR, true, err)
		}

		for _, t := range tg.Tasks {
			task, err := adapter.TaskGet(t.ID)
			if err != nil {
				logger.Println(logger.ERROR, true, err)
			}
			err = adapter.TaskDelete(task)
			if err != nil {
				logger.Println(logger.ERROR, true, err)
			}
		}
	}

	err = adapter.WorkflowDelete(workflow)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	return c.JSONPretty(http.StatusOK, model.SimpleMsg{Message: "success"}, " ")
}
