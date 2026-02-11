package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/cloud-barista/cm-cicada/internal/adapter"
	"github.com/cloud-barista/cm-cicada/internal/domain"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
	"github.com/google/uuid"
	"github.com/jollaman999/utils/logger"
	"github.com/mitchellh/mapstructure"
)

// toTimeHookFunc returns a mapstructure decode hook function for time conversion
func ToTimeHookFunc() mapstructure.DecodeHookFunc {
	return func(
		f reflect.Type,
		t reflect.Type,
		data interface{}) (interface{}, error) {
		if t != reflect.TypeOf(time.Time{}) {
			return data, nil
		}

		switch f.Kind() {
		case reflect.String:
			return time.Parse(time.RFC3339, data.(string))
		case reflect.Float64:
			return time.Unix(0, int64(data.(float64))*int64(time.Millisecond)), nil
		case reflect.Int64:
			return time.Unix(0, data.(int64)*int64(time.Millisecond)), nil
		default:
			return data, nil
		}
		// Convert it by parsing
	}
}

func marshalToRawMessage(data any) json.RawMessage {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return json.RawMessage{}
	}
	return json.RawMessage(jsonData)
}

// createDataReqToData converts CreateDataReq to Data based on spec version
func CreateDataReqToData(specVersion string, createDataReq model.CreateDataReq) (domain.Data, error) {
	specVersionSpilit := strings.Split(specVersion, ".")
	if len(specVersionSpilit) != 2 {
		return domain.Data{}, errors.New("invalid workflow spec version: " + specVersion)
	}

	specVersionMajor, err := strconv.Atoi(specVersionSpilit[0])
	if err != nil {
		return domain.Data{}, errors.New("invalid workflow spec version: " + specVersion)
	}

	specVersionMinor, err := strconv.Atoi(specVersionSpilit[1])
	if err != nil {
		return domain.Data{}, errors.New("invalid workflow spec version: " + specVersion)
	}

	var taskGroups []domain.TaskGroup
	var allTasks []domain.Task

	if specVersionMajor > 0 && specVersionMajor <= 1 {
		if specVersionMinor == 0 {
			// v1.0
			for _, tgReq := range createDataReq.TaskGroups {
				fmt.Println("tgReq", tgReq)
				var tasks []domain.Task
				for _, tReq := range tgReq.Tasks {
					tasks = append(tasks, domain.Task{
						ID:            uuid.New().String(),
						Name:          tReq.Name,
						TaskComponent: tReq.TaskComponent,
						RequestBody:   tReq.RequestBody,
						PathParams:    marshalToRawMessage(tReq.PathParams),
						QueryParams:   marshalToRawMessage(tReq.QueryParams),
						Extra:         marshalToRawMessage(tReq.Extra),
						Dependencies:  marshalToRawMessage(tReq.Dependencies),
					})
				}

				allTasks = append(allTasks, tasks...)
				taskGroups = append(taskGroups, domain.TaskGroup{
					ID:          uuid.New().String(),
					Name:        tgReq.Name,
					Description: tgReq.Description,
					Tasks:       tasks,
				})
			}

			for i, tgReq := range createDataReq.TaskGroups {
				for j, tg := range taskGroups {
					if tgReq.Name == tg.Name {
						if i == j {
							continue
						}

						return domain.Data{}, errors.New("Duplicated task group name: " + tg.Name)
					}
				}
			}

			for i, tCheck := range allTasks {
				for j, t := range allTasks {
					if tCheck.Name == t.Name {
						if i == j {
							continue
						}

						return domain.Data{}, errors.New("Duplicated task name: " + t.Name)
					}
				}
			}
		} else {
			return domain.Data{}, errors.New("Unsupported workflow spec version: " + specVersion)
		}
	} else {
		return domain.Data{}, errors.New("Unsupported workflow spec version: " + specVersion)
	}

	return domain.Data{
		Description: createDataReq.Description,
		TaskGroups:  taskGroups,
	}, nil
}

// getWorkflowFromDB retrieves workflow from DB and populates TaskGroup and Task IDs
func GetWorkflowFromDB(workflowID string) (*domain.Workflow, error) {
	workflow, err := adapter.WorkflowGet(workflowID)
	if err != nil {
		return nil, fmt.Errorf("failed to get the workflow from DB. Error: %s", err.Error())
	}

	for i, tg := range workflow.Data.TaskGroups {
		_, err = adapter.TaskGroupGetByWorkflowIDAndName(workflowID, tg.Name)
		if err != nil {
			logger.Println(logger.ERROR, true, err)
		}

		workflow.Data.TaskGroups[i].ID = tg.ID

		for j, t := range tg.Tasks {
			_, err = adapter.TaskGetByWorkflowIDAndName(workflowID, t.Name)
			if err != nil {
				logger.Println(logger.ERROR, true, err)
			}

			workflow.Data.TaskGroups[i].Tasks[j].ID = t.ID
		}
	}

	return workflow, nil
}
