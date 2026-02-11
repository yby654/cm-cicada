package controller

import (
	"net/http"

	"github.com/cloud-barista/cm-cicada/internal/adapter"
	"github.com/cloud-barista/cm-cicada/internal/lib/airflow"
	"github.com/cloud-barista/cm-cicada/internal/service"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/common"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
	"github.com/labstack/echo/v4"
	"github.com/mitchellh/mapstructure"
)

// UpdateWorkflowVersioned godoc
//
//	@ID		update-workflow-versioned
//	@Summary	Update Workflow (버전 생성)
//	@Description	Workflow를 덮어쓰지 않고 새 WorkflowVersion을 생성한 뒤, 삭제분은 soft delete, 나머지는 upsert. Workflow = 현재 최신 정의 포인터, WorkflowVersion = 불변 스냅샷(append-only).
//	@Tags		[Workflow]
//	@Accept		json
//	@Produce	json
//	@Param		wfId path string true "ID of the workflow."
//	@Param		Workflow body model.CreateWorkflowReq true "Workflow to modify."
//	@Success	200	{object}	model.Workflow	"Successfully update the workflow"
//	@Failure	400	{object}	common.ErrorResponse	"Sent bad request."
//	@Failure	500	{object}	common.ErrorResponse	"Failed to update the workflow"
//	@Router		/workflow/{wfId}/versioned [put]
func UpdateWorkflowVersioned(c echo.Context) error {
	var req model.CreateWorkflowReq

	data, err := common.GetJSONRawBody(c)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Metadata: nil,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			service.ToTimeHookFunc()),
		Result: &req,
	})
	if err != nil {
		return err
	}

	if err = decoder.Decode(data); err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	wfId := c.Param("wfId")
	workflow, err := adapter.WorkflowGet(wfId)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	if req.Name != "" {
		workflow.Name = req.Name
	}

	specVersion := model.WorkflowSpecVersion_LATEST
	if req.SpecVersion != "" {
		specVersion = req.SpecVersion
	}

	workflowData, err := service.CreateDataReqToData(specVersion, req.Data)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	// 1~6: DB Tx 내 버전 생성 + soft delete + upsert + Workflow 캐시 갱신
	if err := service.UpdateWorkflowGraphVersionedTx(workflow, workflowData, specVersion); err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	// 7: Airflow DAG 업데이트 (외부)
	client, err := airflow.GetClient()
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	if err = client.DeleteDAG(workflow.ID, true); err != nil {
		return common.ReturnErrorMsg(c, "Failed to update the workflow (Airflow DAG delete): "+err.Error())
	}

	if err = client.CreateDAG(workflow); err != nil {
		// 8: 외부 실패 시 보상 — DB는 이미 커밋됨, 클라이언트에 실패 반환
		return common.ReturnErrorMsg(c, "Workflow DB updated but Airflow DAG update failed. "+err.Error())
	}

	return c.JSONPretty(http.StatusOK, workflow, " ")
}
