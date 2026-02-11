package airflow

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/cloud-barista/cm-cicada/common"
	"github.com/cloud-barista/cm-cicada/internal/db"
	"github.com/cloud-barista/cm-cicada/internal/domain"
	"github.com/cloud-barista/cm-cicada/internal/lib/config"
	"github.com/jollaman999/utils/fileutil"
	"github.com/jollaman999/utils/logger"
	"gopkg.in/yaml.v3"
)

// unmarshalTaskFields unmarshals []byte fields from Task to their proper types
func unmarshalTaskFields(t *domain.Task) (map[string]string, map[string]string, map[string]interface{}, []string, error) {
	var pathParams map[string]string
	var queryParams map[string]string
	var extra map[string]interface{}
	var dependencies []string

	// Unmarshal PathParams
	if len(t.PathParams) > 0 {
		if err := json.Unmarshal(t.PathParams, &pathParams); err != nil {
			return nil, nil, nil, nil, fmt.Errorf("failed to unmarshal path_params: %w", err)
		}
	}
	if pathParams == nil {
		pathParams = make(map[string]string)
	}

	// Unmarshal QueryParams
	if len(t.QueryParams) > 0 {
		if err := json.Unmarshal(t.QueryParams, &queryParams); err != nil {
			return nil, nil, nil, nil, fmt.Errorf("failed to unmarshal query_params: %w", err)
		}
	}
	if queryParams == nil {
		queryParams = make(map[string]string)
	}

	// Unmarshal Extra
	if len(t.Extra) > 0 {
		if err := json.Unmarshal(t.Extra, &extra); err != nil {
			return nil, nil, nil, nil, fmt.Errorf("failed to unmarshal extra: %w", err)
		}
	}

	// Unmarshal Dependencies
	if len(t.Dependencies) > 0 {
		if err := json.Unmarshal(t.Dependencies, &dependencies); err != nil {
			return nil, nil, nil, nil, fmt.Errorf("failed to unmarshal dependencies: %w", err)
		}
	}
	if dependencies == nil {
		dependencies = []string{}
	}

	return pathParams, queryParams, extra, dependencies, nil
}

func checkWorkflow(workflow *domain.Workflow) error {
	var taskNames []string

	for _, tg := range workflow.Data.TaskGroups {
		if tg.Name == "" {
			return errors.New("task group name should not be empty")
		}

		for _, t := range tg.Tasks {
			if t.Name == "" {
				return errors.New("task name should not be empty")
			}

			taskNames = append(taskNames, t.Name)
		}
	}

	for _, tg := range workflow.Data.TaskGroups {
		for _, t := range tg.Tasks {
			taskComponent := db.TaskComponentGetByName(t.TaskComponent)
			if taskComponent == nil {
				return errors.New("task component '" + t.TaskComponent + "' not found")
			}

			// Unmarshal Dependencies from []byte to []string
			_, _, _, dependencies, err := unmarshalTaskFields(&t)
			if err != nil {
				return fmt.Errorf("failed to unmarshal task fields for %s.%s: %w", tg.Name, t.Name, err)
			}

			for _, dep := range dependencies {
				if t.Name == dep {
					return errors.New("cycle dependency found in " + tg.Name + "." + t.Name)
				}

				var depFound bool
				for _, tName := range taskNames {
					if tName == dep {
						depFound = true
						break
					}
				}
				if !depFound {
					return errors.New("wrong dependency found in " + tg.Name + "." + t.Name + " (" + dep + ")")
				}
			}
		}
	}

	return nil
}

func isTaskExist(workflow *domain.Workflow, taskID string) bool {
	for _, tg := range workflow.Data.TaskGroups {
		for _, t := range tg.Tasks {
			if t.Name == taskID {
				return true
			}
		}
	}

	return false
}

func parseEndpoint(pathParams map[string]string, queryParams map[string]string, endpoint string) (string, error) {
	pathParamKeys := reflect.ValueOf(pathParams).MapKeys()
	for _, key := range pathParamKeys {
		if pathParams[key.String()] == "" {
			return endpoint, fmt.Errorf("path parameter %s is empty", pathParams[key.String()])
		}
		endpoint = strings.ReplaceAll(endpoint, "{"+key.String()+"}", pathParams[key.String()])
	}

	queryParamKeys := reflect.ValueOf(queryParams).MapKeys()
	if len(queryParamKeys) > 0 {
		var queryParamsString string

		for _, key := range queryParamKeys {
			if queryParams[key.String()] == "" {
				continue
			}
			queryParamsString += fmt.Sprintf("%v=%v&", key.String(), queryParams[key.String()])
		}

		if queryParamsString != "" {
			queryParamsString = strings.TrimRight(queryParamsString, "&")

			if !strings.HasSuffix(endpoint, "?") {
				endpoint += "?"
			}
			endpoint += queryParamsString
		}
	}

	return endpoint, nil
}

func writedomainToYAMLFile(domain any, filePath string) error {
	bytes, err := yaml.Marshal(domain)
	if err != nil {
		return err
	}
	parsed := string(bytes)

	return fileutil.WriteFile(filePath, parsed)
}

func writeGustyYAMLs(workflow *domain.Workflow) error {
	err := checkWorkflow(workflow)
	if err != nil {
		return err
	}

	dagDir := config.CMCicadaConfig.CMCicada.DAGDirectoryHost + "/" + workflow.ID
	err = fileutil.CreateDirIfNotExist(dagDir)
	if err != nil {
		return errors.New("failed to create the Workflow directory (Workflow ID=" + workflow.ID +
			", Workflow Name=" + workflow.Name + ", Description: " + workflow.Data.Description)
	}

	type defaultArgs struct {
		Owner         string `yaml:"owner"`
		StartDate     string `yaml:"start_date"`
		Retries       int    `yaml:"retries"`
		RetryDelaySec int    `yaml:"retry_delay_sec"`
	}

	var dagInfo struct {
		defaultArgs defaultArgs `yaml:"default_args"`
		Description string      `yaml:"description"`
	}

	dagInfo.defaultArgs = defaultArgs{
		Owner:         strings.ToLower(common.ModuleName),
		StartDate:     time.Now().Format(time.DateOnly),
		Retries:       0,
		RetryDelaySec: 0,
	}
	dagInfo.Description = workflow.Data.Description

	filePath := dagDir + "/METADATA.yml"

	err = writedomainToYAMLFile(dagInfo, filePath)
	if err != nil {
		return errors.New("failed to write YAML file (FilePath: " + filePath + ", Error: " + err.Error() + ")")
	}

	for _, tg := range workflow.Data.TaskGroups {
		err = fileutil.CreateDirIfNotExist(dagDir + "/" + tg.Name)
		if err != nil {
			return err
		}

		var taskGroup struct {
			Tooltip string `yaml:"tooltip"`
		}

		taskGroup.Tooltip = tg.Description

		filePath = dagDir + "/" + tg.Name + "/METADATA.yml"

		err = writedomainToYAMLFile(taskGroup, filePath)
		if err != nil {
			return errors.New("failed to write YAML file (FilePath: " + filePath + ", Error: " + err.Error() + ")")
		}

		for _, t := range tg.Tasks {
			// Unmarshal []byte fields to proper types
			pathParams, queryParams, extra, dependencies, err := unmarshalTaskFields(&t)
			if err != nil {
				return fmt.Errorf("failed to unmarshal task fields for %s.%s: %w", tg.Name, t.Name, err)
			}

			taskOptions := make(map[string]any)
			taskComponent := db.TaskComponentGetByName(t.TaskComponent)
			if taskComponent == nil {
				return errors.New("task component '" + t.TaskComponent + "' not found")
			}
			logger.Println(logger.INFO, true, fmt.Sprintf("task component extra: %v", taskComponent.Data.Options.Extra))
			logger.Println(logger.INFO, true, fmt.Sprintf("task extra: %v", extra))

			if taskComponent.Data.Options.Extra != nil {
				// taskComponent의 Extra를 복사
				taskOptions = copyMap(taskComponent.Data.Options.Extra)

				// task의 extra가 있으면 병합
				if len(extra) > 0 {
					mergeMaps(taskOptions, extra)
				}
			} else {
				if isTaskExist(workflow, t.RequestBody) {
					taskOptions["operator"] = "local.JsonHttpRequestOperator"
					taskOptions["xcom_task"] = t.RequestBody
				} else {
					taskOptions["operator"] = "airflow.providers.http.operators.http.SimpleHttpOperator"

					type headers struct {
						ContentType string `json:"Content-Type" yaml:"Content-Type"`
					}
					taskOptions["headers"] = headers{
						ContentType: "application/json",
					}

					taskOptions["log_response"] = true

					taskOptions["data"] = t.RequestBody
				}

				taskOptions["http_conn_id"] = taskComponent.Data.Options.APIConnectionID
				endpoint, err := parseEndpoint(pathParams, queryParams, taskComponent.Data.Options.Endpoint)
				if err != nil {
					return errors.New("failed to write YAML file (FilePath: " + filePath + ", Error: " + err.Error() + ")")
				}
				taskOptions["endpoint"] = endpoint
				taskOptions["method"] = taskComponent.Data.Options.Method
			}

			taskOptions["dependencies"] = dependencies

			taskOptions["task_id"] = t.Name

			filePath = dagDir + "/" + tg.Name + "/" + t.Name + ".yml"

			err = writedomainToYAMLFile(taskOptions, filePath)
			if err != nil {
				return errors.New("failed to write YAML file (FilePath: " + filePath + ", Error: " + err.Error() + ")")
			}
		}
	}

	return nil
}

func copyMap(src map[string]any) map[string]any {
	if src == nil {
		return nil
	}
	dst := make(map[string]any, len(src))
	for k, v := range src {
		if nestedMap, ok := v.(map[string]any); ok {
			dst[k] = copyMap(nestedMap)
		} else {
			dst[k] = v
		}
	}
	return dst
}

func mergeMaps(dst map[string]any, src map[string]any) {
	if dst == nil {
		return
	}
	if len(src) == 0 {
		return
	}

	for k, srcValue := range src {
		// dst에 같은 키가 없으면 그대로 추가
		if _, exists := dst[k]; !exists {
			dst[k] = srcValue
			continue
		}

		// dst에 같은 키가 있는 경우
		dstValue := dst[k]

		if dstMap, dstOk := dstValue.(map[string]any); dstOk {
			if srcMap, srcOk := srcValue.(map[string]any); srcOk {
				mergeMaps(dstMap, srcMap)
				continue
			}
		}

		dst[k] = srcValue
	}
}
