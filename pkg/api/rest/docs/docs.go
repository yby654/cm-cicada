// Package docs Code generated by swaggo/swag. DO NOT EDIT
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {},
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/cicada/health": {
            "get": {
                "description": "Check Cicada is alive",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "[Admin] System management"
                ],
                "summary": "Check Cicada is alive",
                "responses": {
                    "200": {
                        "description": "Successfully get heath state.",
                        "schema": {
                            "$ref": "#/definitions/pkg_api_rest_controller.SimpleMsg"
                        }
                    },
                    "500": {
                        "description": "Failed to check health.",
                        "schema": {
                            "$ref": "#/definitions/github_com_cloud-barista_cm-cicada_pkg_api_rest_common.ErrorResponse"
                        }
                    }
                }
            }
        },
        "/task_template": {
            "get": {
                "description": "Get a list of task template.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "[Task Template]"
                ],
                "summary": "List TaskTemplate",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Page of the task template list.",
                        "name": "page",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "Row of the task template list.",
                        "name": "row",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "UUID of the task template.",
                        "name": "uuid",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "Task template name.",
                        "name": "name",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Successfully get a list of task template.",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/github_com_cloud-barista_cm-cicada_pkg_api_rest_model.TaskTemplate"
                            }
                        }
                    },
                    "400": {
                        "description": "Sent bad request.",
                        "schema": {
                            "$ref": "#/definitions/github_com_cloud-barista_cm-cicada_pkg_api_rest_common.ErrorResponse"
                        }
                    },
                    "500": {
                        "description": "Failed to get a list of task template.",
                        "schema": {
                            "$ref": "#/definitions/github_com_cloud-barista_cm-cicada_pkg_api_rest_common.ErrorResponse"
                        }
                    }
                }
            },
            "post": {
                "description": "Register the task template.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "[Task Template]"
                ],
                "summary": "Create TaskTemplate",
                "parameters": [
                    {
                        "description": "task template of the node.",
                        "name": "TaskTemplate",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/github_com_cloud-barista_cm-cicada_pkg_api_rest_model.TaskTemplate"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Successfully register the task template",
                        "schema": {
                            "$ref": "#/definitions/github_com_cloud-barista_cm-cicada_pkg_api_rest_model.TaskTemplate"
                        }
                    },
                    "400": {
                        "description": "Sent bad request.",
                        "schema": {
                            "$ref": "#/definitions/github_com_cloud-barista_cm-cicada_pkg_api_rest_common.ErrorResponse"
                        }
                    },
                    "500": {
                        "description": "Failed to register the task template",
                        "schema": {
                            "$ref": "#/definitions/github_com_cloud-barista_cm-cicada_pkg_api_rest_common.ErrorResponse"
                        }
                    }
                }
            }
        },
        "/task_template/{uuid}": {
            "get": {
                "description": "Get the task template.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "[Task Template]"
                ],
                "summary": "Get TaskTemplate",
                "parameters": [
                    {
                        "type": "string",
                        "description": "UUID of the TaskTemplate",
                        "name": "uuid",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Successfully get the task template",
                        "schema": {
                            "$ref": "#/definitions/github_com_cloud-barista_cm-cicada_pkg_api_rest_model.TaskTemplate"
                        }
                    },
                    "400": {
                        "description": "Sent bad request.",
                        "schema": {
                            "$ref": "#/definitions/github_com_cloud-barista_cm-cicada_pkg_api_rest_common.ErrorResponse"
                        }
                    },
                    "500": {
                        "description": "Failed to get the task template",
                        "schema": {
                            "$ref": "#/definitions/github_com_cloud-barista_cm-cicada_pkg_api_rest_common.ErrorResponse"
                        }
                    }
                }
            },
            "put": {
                "description": "Update the task template.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "[Task Template]"
                ],
                "summary": "Update TaskTemplate",
                "parameters": [
                    {
                        "description": "task template to modify.",
                        "name": "TaskTemplate",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/github_com_cloud-barista_cm-cicada_pkg_api_rest_model.TaskTemplate"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Successfully update the task template",
                        "schema": {
                            "$ref": "#/definitions/github_com_cloud-barista_cm-cicada_pkg_api_rest_model.TaskTemplate"
                        }
                    },
                    "400": {
                        "description": "Sent bad request.",
                        "schema": {
                            "$ref": "#/definitions/github_com_cloud-barista_cm-cicada_pkg_api_rest_common.ErrorResponse"
                        }
                    },
                    "500": {
                        "description": "Failed to update the task template",
                        "schema": {
                            "$ref": "#/definitions/github_com_cloud-barista_cm-cicada_pkg_api_rest_common.ErrorResponse"
                        }
                    }
                }
            },
            "delete": {
                "description": "Delete the task template.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "[Task Template]"
                ],
                "summary": "Delete TaskTemplate",
                "responses": {
                    "200": {
                        "description": "Successfully delete the task template",
                        "schema": {
                            "$ref": "#/definitions/github_com_cloud-barista_cm-cicada_pkg_api_rest_model.TaskTemplate"
                        }
                    },
                    "400": {
                        "description": "Sent bad request.",
                        "schema": {
                            "$ref": "#/definitions/github_com_cloud-barista_cm-cicada_pkg_api_rest_common.ErrorResponse"
                        }
                    },
                    "500": {
                        "description": "Failed to delete the task template",
                        "schema": {
                            "$ref": "#/definitions/github_com_cloud-barista_cm-cicada_pkg_api_rest_common.ErrorResponse"
                        }
                    }
                }
            }
        },
        "/workflow": {
            "get": {
                "description": "Get a list of DAGs from Airflow",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "[Workflow]"
                ],
                "summary": "List Workflow",
                "responses": {
                    "200": {
                        "description": "Successfully get a workflow list.",
                        "schema": {
                            "$ref": "#/definitions/airflow.DAGCollection"
                        }
                    },
                    "400": {
                        "description": "Sent bad request.",
                        "schema": {
                            "$ref": "#/definitions/github_com_cloud-barista_cm-cicada_pkg_api_rest_common.ErrorResponse"
                        }
                    },
                    "500": {
                        "description": "Failed to get a workflow list.",
                        "schema": {
                            "$ref": "#/definitions/github_com_cloud-barista_cm-cicada_pkg_api_rest_common.ErrorResponse"
                        }
                    }
                }
            },
            "post": {
                "description": "Create a DAG in Airflow.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "[Workflow]"
                ],
                "summary": "Create Workflow",
                "parameters": [
                    {
                        "description": "query params",
                        "name": "request",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/github_com_cloud-barista_cm-cicada_pkg_api_rest_model.Workflow"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Successfully create the DAG.",
                        "schema": {
                            "$ref": "#/definitions/github_com_cloud-barista_cm-cicada_pkg_api_rest_model.Workflow"
                        }
                    },
                    "400": {
                        "description": "Sent bad request.",
                        "schema": {
                            "$ref": "#/definitions/github_com_cloud-barista_cm-cicada_pkg_api_rest_common.ErrorResponse"
                        }
                    },
                    "500": {
                        "description": "Failed to create DAG.",
                        "schema": {
                            "$ref": "#/definitions/github_com_cloud-barista_cm-cicada_pkg_api_rest_common.ErrorResponse"
                        }
                    }
                }
            }
        },
        "/workflow/run/{id}": {
            "post": {
                "description": "Run the DAG in Airflow",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "[Workflow]"
                ],
                "summary": "Run Workflow",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Workflow ID",
                        "name": "dag_id",
                        "in": "query",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Successfully run the DAG.",
                        "schema": {
                            "$ref": "#/definitions/github_com_cloud-barista_cm-cicada_pkg_api_rest_model.Workflow"
                        }
                    },
                    "400": {
                        "description": "Sent bad request.",
                        "schema": {
                            "$ref": "#/definitions/github_com_cloud-barista_cm-cicada_pkg_api_rest_common.ErrorResponse"
                        }
                    },
                    "500": {
                        "description": "Failed to run Workflow",
                        "schema": {
                            "$ref": "#/definitions/github_com_cloud-barista_cm-cicada_pkg_api_rest_common.ErrorResponse"
                        }
                    }
                }
            }
        },
        "/workflow/{id}": {
            "get": {
                "description": "Get the DAG from Airflow.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "[Workflow]"
                ],
                "summary": "Get Workflow",
                "responses": {
                    "200": {
                        "description": "Successfully get the DAG.",
                        "schema": {
                            "$ref": "#/definitions/github_com_cloud-barista_cm-cicada_pkg_api_rest_model.Workflow"
                        }
                    },
                    "400": {
                        "description": "Sent bad request.",
                        "schema": {
                            "$ref": "#/definitions/github_com_cloud-barista_cm-cicada_pkg_api_rest_common.ErrorResponse"
                        }
                    },
                    "500": {
                        "description": "Failed to get the DAG.",
                        "schema": {
                            "$ref": "#/definitions/github_com_cloud-barista_cm-cicada_pkg_api_rest_common.ErrorResponse"
                        }
                    }
                }
            }
        },
        "/workflow_template": {
            "get": {
                "description": "Get a list of workflow template.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "[Workflow Template]"
                ],
                "summary": "List WorkflowTemplate",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Page of the workflow template list.",
                        "name": "page",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "Row of the workflow template list.",
                        "name": "row",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "UUID of the workflow template.",
                        "name": "uuid",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "Migration group name.",
                        "name": "name",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Successfully get a list of workflow template.",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/github_com_cloud-barista_cm-cicada_pkg_api_rest_model.Workflow"
                            }
                        }
                    },
                    "400": {
                        "description": "Sent bad request.",
                        "schema": {
                            "$ref": "#/definitions/github_com_cloud-barista_cm-cicada_pkg_api_rest_common.ErrorResponse"
                        }
                    },
                    "500": {
                        "description": "Failed to get a list of workflow template.",
                        "schema": {
                            "$ref": "#/definitions/github_com_cloud-barista_cm-cicada_pkg_api_rest_common.ErrorResponse"
                        }
                    }
                }
            }
        },
        "/workflow_template/{id}": {
            "get": {
                "description": "Get the workflow template.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "[Workflow Template]"
                ],
                "summary": "Get WorkflowTemplate",
                "parameters": [
                    {
                        "type": "string",
                        "description": "UUID of the WorkflowTemplate",
                        "name": "uuid",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Successfully get the workflow template",
                        "schema": {
                            "$ref": "#/definitions/github_com_cloud-barista_cm-cicada_pkg_api_rest_model.Workflow"
                        }
                    },
                    "400": {
                        "description": "Sent bad request.",
                        "schema": {
                            "$ref": "#/definitions/github_com_cloud-barista_cm-cicada_pkg_api_rest_common.ErrorResponse"
                        }
                    },
                    "500": {
                        "description": "Failed to get the workflow template",
                        "schema": {
                            "$ref": "#/definitions/github_com_cloud-barista_cm-cicada_pkg_api_rest_common.ErrorResponse"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "airflow.DAG": {
            "type": "object",
            "properties": {
                "dag_id": {
                    "description": "The ID of the DAG.",
                    "type": "string"
                },
                "default_view": {
                    "description": "Default view of the DAG inside the webserver  *New in version 2.3.0*",
                    "allOf": [
                        {
                            "$ref": "#/definitions/airflow.NullableString"
                        }
                    ]
                },
                "description": {
                    "description": "User-provided DAG description, which can consist of several sentences or paragraphs that describe DAG contents.",
                    "allOf": [
                        {
                            "$ref": "#/definitions/airflow.NullableString"
                        }
                    ]
                },
                "file_token": {
                    "description": "The key containing the encrypted path to the file. Encryption and decryption take place only on the server. This prevents the client from reading an non-DAG file. This also ensures API extensibility, because the format of encrypted data may change.",
                    "type": "string"
                },
                "fileloc": {
                    "description": "The absolute path to the file.",
                    "type": "string"
                },
                "has_import_errors": {
                    "description": "Whether the DAG has import errors  *New in version 2.3.0*",
                    "allOf": [
                        {
                            "$ref": "#/definitions/airflow.NullableBool"
                        }
                    ]
                },
                "has_task_concurrency_limits": {
                    "description": "Whether the DAG has task concurrency limits  *New in version 2.3.0*",
                    "allOf": [
                        {
                            "$ref": "#/definitions/airflow.NullableBool"
                        }
                    ]
                },
                "is_active": {
                    "description": "Whether the DAG is currently seen by the scheduler(s).  *New in version 2.1.1*  *Changed in version 2.2.0*\u0026#58; Field is read-only.",
                    "allOf": [
                        {
                            "$ref": "#/definitions/airflow.NullableBool"
                        }
                    ]
                },
                "is_paused": {
                    "description": "Whether the DAG is paused.",
                    "allOf": [
                        {
                            "$ref": "#/definitions/airflow.NullableBool"
                        }
                    ]
                },
                "is_subdag": {
                    "description": "Whether the DAG is SubDAG.",
                    "type": "boolean"
                },
                "last_expired": {
                    "description": "Time when the DAG last received a refresh signal (e.g. the DAG's \\\"refresh\\\" button was clicked in the web UI)  *New in version 2.3.0*",
                    "allOf": [
                        {
                            "$ref": "#/definitions/airflow.NullableTime"
                        }
                    ]
                },
                "last_parsed_time": {
                    "description": "The last time the DAG was parsed.  *New in version 2.3.0*",
                    "allOf": [
                        {
                            "$ref": "#/definitions/airflow.NullableTime"
                        }
                    ]
                },
                "last_pickled": {
                    "description": "The last time the DAG was pickled.  *New in version 2.3.0*",
                    "allOf": [
                        {
                            "$ref": "#/definitions/airflow.NullableTime"
                        }
                    ]
                },
                "max_active_runs": {
                    "description": "Maximum number of active DAG runs for the DAG  *New in version 2.3.0*",
                    "allOf": [
                        {
                            "$ref": "#/definitions/airflow.NullableInt32"
                        }
                    ]
                },
                "max_active_tasks": {
                    "description": "Maximum number of active tasks that can be run on the DAG  *New in version 2.3.0*",
                    "allOf": [
                        {
                            "$ref": "#/definitions/airflow.NullableInt32"
                        }
                    ]
                },
                "next_dagrun": {
                    "description": "The logical date of the next dag run.  *New in version 2.3.0*",
                    "allOf": [
                        {
                            "$ref": "#/definitions/airflow.NullableTime"
                        }
                    ]
                },
                "next_dagrun_create_after": {
                    "description": "Earliest time at which this ` + "`" + `` + "`" + `next_dagrun` + "`" + `` + "`" + ` can be created.  *New in version 2.3.0*",
                    "allOf": [
                        {
                            "$ref": "#/definitions/airflow.NullableTime"
                        }
                    ]
                },
                "next_dagrun_data_interval_end": {
                    "description": "The end of the interval of the next dag run.  *New in version 2.3.0*",
                    "allOf": [
                        {
                            "$ref": "#/definitions/airflow.NullableTime"
                        }
                    ]
                },
                "next_dagrun_data_interval_start": {
                    "description": "The start of the interval of the next dag run.  *New in version 2.3.0*",
                    "allOf": [
                        {
                            "$ref": "#/definitions/airflow.NullableTime"
                        }
                    ]
                },
                "owners": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "pickle_id": {
                    "description": "Foreign key to the latest pickle_id  *New in version 2.3.0*",
                    "allOf": [
                        {
                            "$ref": "#/definitions/airflow.NullableString"
                        }
                    ]
                },
                "root_dag_id": {
                    "description": "If the DAG is SubDAG then it is the top level DAG identifier. Otherwise, null.",
                    "allOf": [
                        {
                            "$ref": "#/definitions/airflow.NullableString"
                        }
                    ]
                },
                "schedule_interval": {
                    "$ref": "#/definitions/airflow.NullableScheduleInterval"
                },
                "scheduler_lock": {
                    "description": "Whether (one of) the scheduler is scheduling this DAG at the moment  *New in version 2.3.0*",
                    "allOf": [
                        {
                            "$ref": "#/definitions/airflow.NullableBool"
                        }
                    ]
                },
                "tags": {
                    "description": "List of tags.",
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/airflow.Tag"
                    }
                },
                "timetable_description": {
                    "description": "Timetable/Schedule Interval description.  *New in version 2.3.0*",
                    "allOf": [
                        {
                            "$ref": "#/definitions/airflow.NullableString"
                        }
                    ]
                }
            }
        },
        "airflow.DAGCollection": {
            "type": "object",
            "properties": {
                "dags": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/airflow.DAG"
                    }
                },
                "total_entries": {
                    "description": "Count of objects in the current result set.",
                    "type": "integer"
                }
            }
        },
        "airflow.NullableBool": {
            "type": "object"
        },
        "airflow.NullableInt32": {
            "type": "object"
        },
        "airflow.NullableScheduleInterval": {
            "type": "object"
        },
        "airflow.NullableString": {
            "type": "object"
        },
        "airflow.NullableTime": {
            "type": "object"
        },
        "airflow.Tag": {
            "type": "object",
            "properties": {
                "name": {
                    "type": "string"
                }
            }
        },
        "github_com_cloud-barista_cm-cicada_pkg_api_rest_common.ErrorResponse": {
            "type": "object",
            "properties": {
                "error": {
                    "type": "string"
                }
            }
        },
        "github_com_cloud-barista_cm-cicada_pkg_api_rest_model.Data": {
            "type": "object",
            "required": [
                "default_args",
                "task_groups"
            ],
            "properties": {
                "default_args": {
                    "$ref": "#/definitions/github_com_cloud-barista_cm-cicada_pkg_api_rest_model.DefaultArgs"
                },
                "description": {
                    "type": "string"
                },
                "task_groups": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/github_com_cloud-barista_cm-cicada_pkg_api_rest_model.TaskGroup"
                    }
                }
            }
        },
        "github_com_cloud-barista_cm-cicada_pkg_api_rest_model.DefaultArgs": {
            "type": "object",
            "required": [
                "owner",
                "start_date"
            ],
            "properties": {
                "email": {
                    "type": "string"
                },
                "email_on_failure": {
                    "type": "boolean"
                },
                "email_on_retry": {
                    "type": "boolean"
                },
                "owner": {
                    "type": "string"
                },
                "retries": {
                    "description": "default: 1",
                    "type": "integer"
                },
                "retry_delay_sec": {
                    "description": "default: 300",
                    "type": "integer"
                },
                "start_date": {
                    "type": "string"
                }
            }
        },
        "github_com_cloud-barista_cm-cicada_pkg_api_rest_model.Task": {
            "type": "object",
            "required": [
                "operator",
                "operator_options",
                "task_component",
                "task_name"
            ],
            "properties": {
                "dependencies": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "operator": {
                    "type": "string"
                },
                "operator_options": {
                    "type": "array",
                    "items": {
                        "type": "object",
                        "required": [
                            "name",
                            "value"
                        ],
                        "properties": {
                            "name": {
                                "type": "string"
                            },
                            "value": {}
                        }
                    }
                },
                "task_component": {
                    "type": "string"
                },
                "task_name": {
                    "type": "string"
                }
            }
        },
        "github_com_cloud-barista_cm-cicada_pkg_api_rest_model.TaskGroup": {
            "type": "object",
            "required": [
                "task_group_name",
                "tasks"
            ],
            "properties": {
                "description": {
                    "type": "string"
                },
                "task_group_name": {
                    "type": "string"
                },
                "tasks": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/github_com_cloud-barista_cm-cicada_pkg_api_rest_model.Task"
                    }
                }
            }
        },
        "github_com_cloud-barista_cm-cicada_pkg_api_rest_model.TaskTemplate": {
            "type": "object",
            "required": [
                "id",
                "name",
                "task"
            ],
            "properties": {
                "created_at": {
                    "type": "string"
                },
                "id": {
                    "type": "string"
                },
                "name": {
                    "type": "string"
                },
                "task": {
                    "$ref": "#/definitions/github_com_cloud-barista_cm-cicada_pkg_api_rest_model.Task"
                },
                "updated_at": {
                    "type": "string"
                }
            }
        },
        "github_com_cloud-barista_cm-cicada_pkg_api_rest_model.Workflow": {
            "type": "object",
            "required": [
                "data",
                "id",
                "name"
            ],
            "properties": {
                "created_at": {
                    "type": "string"
                },
                "data": {
                    "$ref": "#/definitions/github_com_cloud-barista_cm-cicada_pkg_api_rest_model.Data"
                },
                "id": {
                    "type": "string"
                },
                "name": {
                    "type": "string"
                },
                "updated_at": {
                    "type": "string"
                }
            }
        },
        "pkg_api_rest_controller.SimpleMsg": {
            "type": "object",
            "properties": {
                "message": {
                    "type": "string"
                }
            }
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "",
	Host:             "",
	BasePath:         "",
	Schemes:          []string{},
	Title:            "",
	Description:      "",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
