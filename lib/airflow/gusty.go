package airflow

import (
	"errors"
	"github.com/cloud-barista/cm-cicada/common"
	"github.com/cloud-barista/cm-cicada/lib/config"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
	"github.com/jollaman999/utils/fileutil"
	"gopkg.in/yaml.v3"
	"strings"
	"time"
)

func checkDAG(dag *model.Workflow) error {
	var taskNames []string

	for _, tg := range dag.Data.TaskGroups {
		if tg.ID == "" {
			return errors.New("task group id should not be empty")
		}

		for _, t := range tg.Tasks {
			if t.ID == "" {
				return errors.New("task id should not be empty")
			}

			taskNames = append(taskNames, t.ID)
		}
	}

	for _, tg := range dag.Data.TaskGroups {
		for _, t := range tg.Tasks {
			for _, dep := range t.Dependencies {
				var depFound bool
				for _, tName := range taskNames {
					if tName == dep {
						depFound = true
						break
					}
				}
				if !depFound {
					return errors.New("wrong dependency found in " + tg.ID + "." + t.ID + " (" + dep + ")")
				}
			}
		}
	}

	return nil
}

func writeModelToYAMLFile(model any, filePath string) error {
	bytes, err := yaml.Marshal(model)
	if err != nil {
		return err
	}
	parsed := string(bytes)

	return fileutil.WriteFile(filePath, parsed)
}

func writeGustyYAMLs(dag *model.Workflow) error {
	err := checkDAG(dag)
	if err != nil {
		return err
	}

	dagDir := config.CMCicadaConfig.CMCicada.DAGDirectoryHost + "/" + dag.UUID
	err = fileutil.CreateDirIfNotExist(dagDir)
	if err != nil {
		return errors.New("failed to create the Workflow directory (Workflow ID=" + dag.ID +
			", Workflow UUID=" + dag.UUID + ", Description: " + dag.Data.Description)
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
	dagInfo.Description = dag.Data.Description

	filePath := dagDir + "/METADATA.yml"

	err = writeModelToYAMLFile(dagInfo, filePath)
	if err != nil {
		return errors.New("failed to write YAML file (FilePath: " + filePath + ", Error: " + err.Error() + ")")
	}

	for _, tg := range dag.Data.TaskGroups {
		err = fileutil.CreateDirIfNotExist(dagDir + "/" + tg.ID)
		if err != nil {
			return err
		}

		var taskGroup struct {
			Tooltip string `yaml:"tooltip"`
		}

		taskGroup.Tooltip = tg.Description

		filePath = dagDir + "/" + tg.ID + "/METADATA.yml"

		err = writeModelToYAMLFile(taskGroup, filePath)
		if err != nil {
			return errors.New("failed to write YAML file (FilePath: " + filePath + ", Error: " + err.Error() + ")")
		}

		for _, t := range tg.Tasks {
			taskOptions := make(map[string]any)

			taskOptions["operator"] = "airflow.providers.http.operators.http.SimpleHttpOperator"

			type headers struct {
				ContentType string `json:"Content-Type"`
			}
			taskOptions["headers"] = headers{
				ContentType: "application/json",
			}

			taskOptions["dependencies"] = t.Dependencies

			taskOptions["task_id"] = t.ID
			taskOptions["http_conn_id"] = t.Options.APIConnectionID
			taskOptions["endpoint"] = t.Options.Endpoint
			taskOptions["method"] = t.Options.Method

			filePath = dagDir + "/" + tg.ID + "/" + t.ID + ".yml"

			err = writeModelToYAMLFile(taskOptions, filePath)
			if err != nil {
				return errors.New("failed to write YAML file (FilePath: " + filePath + ", Error: " + err.Error() + ")")
			}
		}
	}

	return nil
}
