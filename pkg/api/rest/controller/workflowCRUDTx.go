package controller

import (
	"net/http"

	"github.com/cloud-barista/cm-cicada/internal/domain"
	"github.com/cloud-barista/cm-cicada/internal/lib/airflow"
	"github.com/cloud-barista/cm-cicada/internal/service"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/common"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/mitchellh/mapstructure"
)

// CreateWorkflowTx godoc
//
//	@ID			create-workflow-tx
//	@Summary		Create Workflow (DB Tx)
//	@Description	Create a workflow with TaskGroups/Tasks in a single DB transaction.
//	@Tags			[Workflow]
//	@Accept			json
//	@Produce		json
//	@Param			request body 	model.CreateWorkflowReq true "Workflow content"
//	@Success		200	{object}	model.Workflow	"Successfully create the workflow."
//	@Failure		400	{object}	common.ErrorResponse	"Sent bad request."
//	@Failure		500	{object}	common.ErrorResponse	"Failed to create workflow."
//	@Router			/workflow/tx [post]
func CreateWorkflowTx(c echo.Context) error {
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

	if err := decoder.Decode(data); err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	if createWorkflowReq.Name == "" {
		return common.ReturnErrorMsg(c, "Please provide the name.")
	}

	specVersion := model.WorkflowSpecVersion_LATEST
	if createWorkflowReq.SpecVersion != "" {
		specVersion = createWorkflowReq.SpecVersion
	}

	workflowData, err := service.CreateDataReqToData(specVersion, createWorkflowReq.Data)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	workflow := domain.Workflow{
		ID:          uuid.New().String(),
		SpecVersion: specVersion,
		Name:        createWorkflowReq.Name,
		Data:        workflowData,
	}

	// 1) DB transaction: Workflow + TaskGroups + Tasks
	if err := service.CreateWorkflowGraphTx(&workflow); err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	// 2) External side-effect: Airflow DAG creation (NOT transactional)
	client, err := airflow.GetClient()
	if err != nil {
		_ = service.DeleteWorkflowGraph(workflow.ID)
		return common.ReturnErrorMsg(c, err.Error())
	}

	if err := client.CreateDAG(&workflow); err != nil {
		_ = service.DeleteWorkflowGraph(workflow.ID)
		return common.ReturnErrorMsg(c, "Failed to create the workflow. (Error:"+err.Error()+")")
	}

	return c.JSONPretty(http.StatusOK, workflow, " ")
}
