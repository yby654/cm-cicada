package airflow

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"

	"github.com/apache/airflow-client-go/airflow"
	"github.com/cloud-barista/cm-cicada/lib/config"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
	"github.com/jollaman999/utils/fileutil"
	"github.com/jollaman999/utils/logger"
)

var dagRequests = make(map[string]*sync.Mutex)
var dagRequestsLock = sync.Mutex{}

func callDagRequestLock(workflowID string) func() {
	dagRequestsLock.Lock()
	_, exist := dagRequests[workflowID]
	if !exist {
		dagRequests[workflowID] = new(sync.Mutex)
	}
	dagRequestsLock.Unlock()

	dagRequests[workflowID].Lock()

	return func() {
		dagRequests[workflowID].Unlock()

		dagRequestsLock.Lock()
		delete(dagRequests, workflowID)
		dagRequestsLock.Unlock()
	}
}

func (client *client) CreateDAG(workflow *model.Workflow) error {
	deferFunc := callDagRequestLock(workflow.ID)
	defer func() {
		deferFunc()
	}()

	err := writeGustyYAMLs(workflow)
	if err != nil {
		return err
	}

	return nil
}

func (client *client) GetDAG(dagID string) (airflow.DAG, error) {
	deferFunc := callDagRequestLock(dagID)
	defer func() {
		deferFunc()
	}()

	ctx, cancel := Context()
	defer cancel()
	resp, _, err := client.api.DAGApi.GetDag(ctx, dagID).Execute()
	if err != nil {
		logger.Println(logger.ERROR, false,
			"AIRFLOW: Error occurred while getting the DAG. (Error: "+err.Error()+").")
	}

	return resp, err
}

func (client *client) GetDAGs() (airflow.DAGCollection, error) {
	ctx, cancel := Context()
	defer cancel()
	resp, _, err := client.api.DAGApi.GetDags(ctx).Execute()
	if err != nil {
		logger.Println(logger.ERROR, false,
			"AIRFLOW: Error occurred while getting DAGs. (Error: "+err.Error()+").")
	}
	return resp, err
}

func (client *client) RunDAG(dagID string) (airflow.DAGRun, error) {
	deferFunc := callDagRequestLock(dagID)
	defer func() {
		deferFunc()
	}()

	dags, err := client.GetDAGs()
	if err != nil {
		return airflow.DAGRun{}, err
	}

	var found = false

	for _, dag := range *dags.Dags {
		if *dag.DagId == dagID {
			found = true
			break
		}
	}

	if !found {
		logger.Println(logger.DEBUG, false,
			"AIRFLOW: Received the request with none existing DAG ID. (DAG ID: "+dagID+")")
		return airflow.DAGRun{}, errors.New("provided dag_id is not exist")
	}

	ctx, cancel := Context()
	defer cancel()
	resp, _, err := client.api.DAGRunApi.PostDagRun(ctx, dagID).DAGRun(*airflow.NewDAGRun()).Execute()
	if err != nil {
		errMsg := "AIRFLOW: Error occurred while running the DAG. (DAG ID: " + dagID + ", Error: " + err.Error() + ")"
		logger.Println(logger.ERROR, false, errMsg)

		return airflow.DAGRun{}, errors.New(errMsg)
	}

	logger.Println(logger.INFO, false, "AIRFLOW: Running the DAG. (DAG ID: "+dagID+")")

	return resp, err
}

func (client *client) DeleteDAG(dagID string, deleteFolderOnly bool) error {
	deferFunc := callDagRequestLock(dagID)
	defer func() {
		deferFunc()
	}()

	dagDir := config.CMCicadaConfig.CMCicada.DAGDirectoryHost + "/" + dagID
	err := fileutil.DeleteDir(dagDir)
	if err != nil {
		logger.Println(logger.ERROR, true,
			"AIRFLOW: Failed to delete dag directory. (Error: "+err.Error()+").")
	}

	if !deleteFolderOnly {
		ctx, cancel := Context()
		defer cancel()
		_, err = client.api.DAGApi.DeleteDag(ctx, dagID).Execute()
		if err != nil {
			logger.Println(logger.ERROR, false,
				"AIRFLOW: Error occurred while deleting the DAG. (Error: "+err.Error()+").")
		}
	}

	return err
}
func (client *client) GetDAGRuns(dagID string) (airflow.DAGRunCollection, error) {
	deferFunc := callDagRequestLock(dagID)
	defer func() {
		deferFunc()
	}()
	ctx, cancel := Context()
	defer cancel()
	resp, _, err := client.api.DAGRunApi.GetDagRuns(ctx, dagID).Execute()
	if err != nil {
		logger.Println(logger.ERROR, false,
			"AIRFLOW: Error occurred while getting DAGRuns. (Error: "+err.Error()+").")
	}
	return resp, err
}

func (client *client) GetTaskInstances(dagID string, dagRunId string) (airflow.TaskInstanceCollection, error) {
	deferFunc := callDagRequestLock(dagID)
	defer func() {
		deferFunc()
	}()
	ctx, cancel := Context()
	defer cancel()
	resp, http, err := client.api.TaskInstanceApi.GetTaskInstances(ctx, dagID, dagRunId).Execute()
	fmt.Println("test : ", http)
	if err != nil {
		logger.Println(logger.ERROR, false,
			"AIRFLOW: Error occurred while getting TaskInstances. (Error: "+err.Error()+").")
	}
	return resp, err
}

func (client *client) GetTaskLogs(dagID, dagRunID, taskID string, taskTryNumber int) (airflow.InlineResponse200, error) {
	deferFunc := callDagRequestLock(dagID)
	defer func() {
		deferFunc()
	}()
	ctx, cancel := Context()
	defer cancel()

	// TaskInstanceApi 인스턴스를 사용하여 로그 요청
	logs, _, err := client.api.TaskInstanceApi.GetLog(ctx, dagID, dagRunID, taskID, int32(taskTryNumber)).FullContent(true).Execute()
	logger.Println(logger.INFO, false, logs)
	if err != nil {
		logger.Println(logger.ERROR, false,
			"AIRFLOW: Error occurred while getting TaskInstance logs. (Error: "+err.Error()+").")
	}

	return logs, nil
}

func (client *client) ClearTaskInstance(dagID string, dagRunID string, taskID string) (airflow.TaskInstanceReferenceCollection, error) {
	deferFunc := callDagRequestLock(dagID)
	defer func() {
		deferFunc()
	}()
	ctx, cancel := Context()
	defer cancel()

	dryRun := false
	taskIds := []string{taskID}
	includeDownstream := true
	includeFuture := false
	includeParentdag := false
	includePast := false
	includeSubdags := true
	includeUpstream := false
	onlyFailed := false
	onlyRunning := false
	resetDagRuns := true

	clearTask := airflow.ClearTaskInstances{
		DryRun:            &dryRun,
		TaskIds:           &taskIds,
		IncludeSubdags:    &includeSubdags,
		IncludeParentdag:  &includeParentdag,
		IncludeUpstream:   &includeUpstream,
		IncludeDownstream: &includeDownstream,
		IncludeFuture:     &includeFuture,
		IncludePast:       &includePast,
		OnlyFailed:        &onlyFailed,
		OnlyRunning:       &onlyRunning,
		ResetDagRuns:      &resetDagRuns,
		DagRunId:          *airflow.NewNullableString(&dagRunID),
	}

	// 요청 생성
	request := client.api.DAGApi.PostClearTaskInstances(ctx, dagID)

	// ClearTaskInstances 데이터 설정
	request = request.ClearTaskInstances(clearTask)

	// 요청 실행
	logs, _, err := client.api.DAGApi.PostClearTaskInstancesExecute(request)
	if err != nil {
		logger.Println(logger.ERROR, false,
			"AIRFLOW: Error occurred while clearing TaskInstance. (Error: "+err.Error()+").")
		return airflow.TaskInstanceReferenceCollection{}, err
	}

	return logs, nil
}

func (client *client) GetEventLogs(dagID string, dagRunId string, taskId string) (airflow.EventLogCollection, error) {
	deferFunc := callDagRequestLock(dagID)
	defer func() {
		deferFunc()
	}()
	ctx, cancel := Context()
	defer cancel()

	localBasePath, err := client.api.GetConfig().ServerURLWithContext(ctx, "EventLogApiService.GetEventLog")
	if err != nil {
		fmt.Println("Error occurred while getting event logs:", err)
	}

	baseURL := "http://" + client.api.GetConfig().Host + localBasePath + "/eventLogs"
	queryParams := map[string]string{
		"offset":          "0",
		"limit":           "100",
		"dag_id":          dagID,
		"run_id":          dagRunId,
		"task_id":         taskId,
		"order_by":        "-when",
		"excluded_events": "gantt,landing_times,tries,duration,calendar,graph,grid,tree,tree_data",
	}
	query := url.Values{}
	for key, value := range queryParams {
		query.Add(key, value)
	}
	queryString := query.Encode()
	fullURL := fmt.Sprintf("%s?%s", baseURL, queryString)
	httpclient := client.api.GetConfig().HTTPClient
	// fmt.Println(&httpclient.)

	// 요청 생성
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		fmt.Println("Error occurred while creating the request:", err)
	}
	cred := ctx.Value(airflow.ContextBasicAuth).(airflow.BasicAuth)
	addBasicAuth(req, cred.UserName, cred.Password)
	res, err := httpclient.Do(req)
	if err != nil {
		fmt.Println("Error occurred while sending the request:", err)
	}
	defer func() {
		_ = res.Body.Close()
	}()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
	}

	var eventlogs airflow.EventLogCollection
	err = json.Unmarshal(body, &eventlogs)
	if err != nil {
		fmt.Println("Error unmarshal response body:", err)
	}

	return eventlogs, err
}

func (client *client) GetImportErrors() (airflow.ImportErrorCollection, error) {
	ctx, cancel := Context()
	defer cancel()

	// TaskInstanceApi 인스턴스를 사용하여 로그 요청
	logs, _, err := client.api.ImportErrorApi.GetImportErrors(ctx).Execute()
	logger.Println(logger.INFO, false, logs)
	if err != nil {
		logger.Println(logger.ERROR, false,
			"AIRFLOW: Error occurred while getting import dag errors. (Error: "+err.Error()+").")
	}

	return logs, nil
}

func (client *client) PatchDag(dagID string, dagBody airflow.DAG) (airflow.DAG, error) {
	ctx, cancel := Context()
	defer cancel()

	// TaskInstanceApi 인스턴스를 사용하여 로그 요청
	logs, _, err := client.api.DAGApi.PatchDag(ctx, dagID).DAG(dagBody).Execute()
	logger.Println(logger.INFO, false, logs)
	if err != nil {
		logger.Println(logger.ERROR, false,
			"AIRFLOW: Error occurred while getting import dag errors. (Error: "+err.Error()+").")
	}

	return logs, nil
}

func addBasicAuth(req *http.Request, username, password string) {
	auth := username + ":" + password
	encodedAuth := base64.StdEncoding.EncodeToString([]byte(auth))
	req.Header.Add("Authorization", "Basic "+encodedAuth)
}
