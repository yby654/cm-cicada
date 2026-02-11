import os
import airflow
from gusty import create_dag
from airflow.providers.http.hooks.http import HttpHook
from airflow.utils.state import State



########################
# cicada callback 먼저 #
########################
def cicada_workflowrun_callback(context):

    dr = context["dag_run"]
    wf_id = dr.dag_id
    payload = {
        "workflow_run_id": dr.run_id,
        "workflow_id": dr.dag_id,
        "execution_date": _dt(dr.execution_date),
        "start_date": _dt(dr.start_date),
        "end_date": _dt(dr.end_date),
        "run_type": str(dr.run_type),
        "state": dr.state
    }

    hook = HttpHook(
        method="POST",
        http_conn_id="cicada_api"
    )

    try:
        hook.run(
            endpoint="/cicada/workflow/{wf_id}/runs",
            json=payload,
            headers={"Content-Type": "application/json"}
        )
    except Exception as e:
        print(f"[WARN] cicada workflow run report failed: {e}")


def _dt(v):
    return v.isoformat() if v else None




#####################
## DAG Directories ##
#####################

# point to your dags directory
dag_parent_dir = os.path.join(os.environ['AIRFLOW_HOME'], "dags")

# assumes any subdirectories in the dags directory are Gusty DAGs (with METADATA.yml) (excludes subdirectories like __pycache__)
dag_directories = [os.path.join(dag_parent_dir, name) for name in os.listdir(dag_parent_dir) if os.path.isdir(os.path.join(dag_parent_dir, name)) and not name.endswith('__')]

####################
## DAG Generation ##
####################
for dag_directory in dag_directories:

    dag_id = os.path.basename(dag_directory)

    dag = create_dag(
        dag_directory,
        tags=['default', 'tags'],
        task_group_defaults={"tooltip": "default tooltip"},
        wait_for_defaults={"retries": 10, "check_existence": True},
        latest_only=False
    )

    # ✅ 전 DAG 공통 이력 적재
    dag.on_success_callback = cicada_workflowrun_callback
    dag.on_failure_callback = cicada_workflowrun_callback

    globals()[dag_id] = dag

# for dag_directory in dag_directories:
#     dag_id = os.path.basename(dag_directory)
#     globals()[dag_id] = create_dag(dag_directory,
#                                    tags = ['default', 'tags'],
#                                    task_group_defaults={"tooltip": "default tooltip"},
#                                    wait_for_defaults={"retries": 10, "check_existence": True},
#                                    latest_only=False)

 