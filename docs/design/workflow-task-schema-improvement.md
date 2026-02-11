# Workflow / Task / TaskGroup / TaskComponent 스키마 개선 제안

## 1. 현재 문제 요약

| 문제 | 현상 |
|------|------|
| **역할 불명확** | task_group / task / task_component 용도가 섞여 있음 |
| **참조 약함** | Workflow ↔ Task/TaskGroup은 ID로만 연결, DB FK/객체 관계 미흡 |
| **Task ↔ TaskGroup** | Task가 어떤 TaskGroup 소속인지 DB에서 직접 관리되지 않음 → 그룹별 Task 조회 어려움 |
| **Task ↔ TaskComponent** | 연결 없음 → Task가 어떤 컴포넌트를 쓰는지 추적 불가 |
| **Task 직접 저장 없음** | Task 상세(name, request_body, path_params 등)가 Workflow.data Blob에만 있음, task 테이블에는 ID/Name/WorkflowID/TaskGroupID 수준만 저장 |
| **Task 단위 조회/이력 불가** | Task 이력은 WorkflowVersion.data로만 가능, Task 단위 직접 조회 불가 |
| **Template vs 실행 혼재** | Template이 실행 가능한 전체 값을 갖고, 실행 시 그대로 Workflow로 취급 → “정의”와 “실행” 구분 없음 |
| **실행 기준 구조 추적 불가** | “이 실행이 어떤 구조(버전) 기준이었는지” 연결되지 않음 → 구조 변경 시 과거 실행 해석/재현/장애 분석 어려움 |

---

## 2. 목표 개념 정리

### 2.1 역할 정의

| 엔티티 | 역할 | 비유 |
|--------|------|------|
| **TaskComponent** | **재사용 가능한 작업 컴포넌트 정의** (endpoint, method, 파라미터 스키마 등). 여러 Task에서 참조. | “HTTP GET 컴포넌트”, “이메일 발송 컴포넌트” |
| **WorkflowTemplate** | **실행 전의 워크플로 정의**(이름, 설명, 그룹/태스크 구조). 실제 request_body 등은 Template 또는 실행 시점에 바인딩. | “마이그레이션 워크플로 템플릿” |
| **Workflow** | Template을 기반으로 한 **실제 워크플로 인스턴스**(현재 “살아 있는” 정의). | “프로젝트 A 마이그레이션 워크플로” |
| **WorkflowVersion** | **특정 시점의 워크플로 정의 스냅샷**. 실행/변경 이력 해석의 기준. | “v1”, “v2” |
| **TaskGroup** | **Task 묶음 단위**(이름, 설명). Workflow(또는 Version) 소속. | “준비 단계”, “마이그레이션 단계” |
| **Task** | **실제 작업 단위**. TaskGroup 소속, TaskComponent 참조, request_body/path_params 등 **전부 DB에 저장**. | “소스 정보 조회”, “타겟 생성” |
| **WorkflowRun** | **한 번의 실행**. “어떤 버전(WorkflowVersion) 기준으로 돌았는지” 명시. | “2024-01-15 10:00 실행” |
| **TaskRun** | **한 Run 안에서의 Task 실행 이력**. Task + WorkflowRun으로 연결. | “위 Run에서의 ‘소스 조회’ 실행” |

### 2.2 관계 원칙

- **정의 계층**: Workflow → WorkflowVersion(스냅샷) → TaskGroup → Task → TaskComponent(참조)
- **실행 계층**: WorkflowRun → WorkflowVersion(기준 버전), WorkflowRun → TaskRun → Task
- **Task는 항상**  
  - 하나의 TaskGroup에 속하고 (task_group_id),  
  - 하나의 TaskComponent를 참조하며 (task_component_id 또는 name),  
  - **모든 입력(request_body, path_params, dependencies 등)을 task 테이블에 저장**해 Task 단위 직접 조회/이력 가능하게 함.

---

## 3. 권장 스키마 (요지)

### 3.1 명시적 FK 및 역할

```
TaskComponent (기존 유지)
  - id, name, data(options, path_params, query_params, body_params), ...
  - 역할: 재사용 가능 컴포넌트 정의

WorkflowTemplate (기존 유지)
  - id, name, spec_version, data, ...
  - 역할: 템플릿 수준 정의

Workflow (현재 유지 + 선택적 정리)
  - id, name, spec_version, data(JSON, 캐시/호환용), created_at, updated_at
  - data는 “현재 정의” 캐시로 두거나, 점진적으로 WorkflowVersion + TaskGroup + Task 조합으로 대체

WorkflowVersion (강화)
  - id, workflow_id, version, spec_version, definition_snapshot, created_at
  - workflow_id → Workflow.id (FK)
  - 역할: “이 버전 기준으로 실행했다”의 기준

TaskGroup (강화)
  - id, workflow_id, name, description, sort_order(선택), created_at
  - workflow_id → Workflow.id (FK)  ※ 현재는 workflow_version_id로 Workflow.id 넣고 있음 → 네이밍 정리 권장
  - 역할: 그룹 메타, Task 묶음

Task (강화, 핵심)
  - id, task_group_id, task_component_id, name, request_body, path_params, query_params, extra, dependencies, sort_order(선택), created_at
  - task_group_id → TaskGroup.id (FK)
  - task_component_id → TaskComponent.id (FK)  ※ 현재는 task_component name만 저장 → id 참조 권장
  - 역할: “실제 작업 정의” 전부를 보유 → Task 단위 직접 조회/이력 가능

WorkflowRun (강화)
  - id, workflow_id, workflow_version_id(신규), ... (기존 컬럼 유지)
  - workflow_id → Workflow.id
  - workflow_version_id → WorkflowVersion.id (FK)  ※ “이 실행이 어떤 버전 기준인지” 명시
  - 역할: 한 번의 실행 + 버전 추적

TaskRun (현재 유지)
  - id, workflow_run_id, task_id, state, start_date, end_date, try_number, ...
  - workflow_run_id → WorkflowRun.id, task_id → Task.id
  - 역할: Run 단위 Task 실행 이력
```

### 3.2 Task “직접 저장” 규칙

- **현재**: Task 상세는 `Workflow.data` Blob에만 있고, `task` 테이블에는 ID/Name/WorkflowID/TaskGroupID 등 일부만 저장.
- **권장**:  
  - **task 테이블에** task_component_id, request_body, path_params, query_params, extra, dependencies 등 **전부 저장**.  
  - **Workflow.data**는  
    - 기존 API/호환을 위해 “현재 정의” JSON으로 유지하되, **WorkflowVersion 생성 시점에 TaskGroup+Task를 조회해 조합**하거나,  
    - 점진적으로 “캐시”로만 두고, **단일 소스 오브 트루스는 WorkflowVersion + task_group + task**로 이전.

이렇게 하면:

- Task 단위 직접 조회 가능.
- TaskGroup 기준 Task 목록 조회: `WHERE task_group_id = ?`.
- TaskComponent 기준 Task 목록: `WHERE task_component_id = ?`.
- Task 이력은 TaskRun + Task + WorkflowVersion으로 “어떤 구조로 실행되었는지”까지 추적 가능.

---

## 4. Template vs Workflow vs Run 구분

### 4.1 정의와 실행 분리

| 구분 | 엔티티 | 설명 |
|------|--------|------|
| **정의** | WorkflowTemplate | 실행 가능한 “형태”만 정의. 실제 인스턴스 아님. |
| **인스턴스 정의** | Workflow | Template 기반 생성된 “현재” 워크플로. TaskGroup/Task가 여기 또는 Version에 소속. |
| **정의 스냅샷** | WorkflowVersion | 생성/수정 시점의 정의 고정. 실행 시 “이 버전 기준”으로 사용. |
| **실행** | WorkflowRun | 한 번의 실행. **workflow_version_id**로 “어떤 버전 기준이었는지” 저장. |
| **실행 상세** | TaskRun | Run 내 개별 Task 실행 결과. |

### 4.2 실행 시 권장 흐름

1. 사용자가 “실행” 요청 시  
   - 현재 Workflow의 **최신 WorkflowVersion**을 결정하거나,  
   - 지정된 version_id를 사용.
2. **WorkflowRun** 생성 시  
   - `workflow_id` + `workflow_version_id` 저장.  
   - 이후 “이 실행은 이 버전 정의 기준”이라고 명확히 함.
3. 실행 엔진(Airflow 등)은  
   - WorkflowVersion.definition_snapshot 또는  
   - 해당 Version에 연결된 TaskGroup + Task를 조회해 실행.
4. Task 구조가 나중에 바뀌어도  
   - 과거 Run은 `workflow_version_id`로 스냅샷/정의를 조회해 해석·재현·장애 분석 가능.

---

## 5. 단계별 적용 방안

### Phase 1: FK와 역할 명시 (단기)

- **domain**  
  - TaskGroup: `workflow_id`(또는 현재처럼 workflow_version_id에 Workflow.id 저장) + `Workflow` association 유지.  
  - Task: `task_group_id`, `task_component_id`(또는 name) + TaskGroup/TaskComponent association 명시.  
  - WorkflowRun: `workflow_version_id` 컬럼 + WorkflowVersion association 추가.
- **DB**  
  - SQLite에서 FK를 켜고 싶다면 `gorm.Config{ DisableForeignKeyConstraintWhenMigrating: false }` 등으로 마이그레이션 시 FK 생성 검토 (SQLite 제한 있음).  
  - 최소한 **애플리케이션 레벨**에서 Create/Update/Delete 시 FK 일관성 유지.

### Phase 2: Task 전량 저장 (중기)

- **task 테이블**  
  - request_body, path_params, query_params, extra, dependencies 등 **전부 컬럼(또는 JSON blob)**으로 저장.  
  - Task 생성/수정 시 Workflow.data에서 복사해 넣거나, API에서 Task를 “주인”으로 두고 저장.
- **Workflow.data**  
  - “현재 정의” 캐시로 유지.  
  - WorkflowVersion 생성 시: task_group + task를 조회해 definition_snapshot 구성하거나, 동일 정보를 Task 테이블에서 읽어오도록 변경.

### Phase 3: WorkflowRun ↔ WorkflowVersion (중기)

- **WorkflowRun**  
  - `workflow_version_id` 추가.  
  - 실행 시 “현재 최신 버전” 또는 지정 버전을 넣어 저장.  
- **조회/분석**  
  - “이 실행이 어떤 구조였는지” → WorkflowRun.workflow_version_id → WorkflowVersion.definition_snapshot 또는 Version 기준 TaskGroup/Task 조회.

### Phase 4: TaskComponent FK (단기~중기)

- **Task**  
  - `task_component_id` (TaskComponent.id) 추가.  
  - 기존 `task_component`(name)는 호환용으로 유지 후 점진적 이전.  
- **TaskComponent**  
  - 기존 id 사용.  
  - “이 Task가 어떤 컴포넌트를 쓰는지” 조회: Task.task_component_id 또는 name.

### Phase 5: Template vs Workflow 분리 (장기, 선택)

- **WorkflowTemplate**  
  - “실행 가능한 전체 값”을 갖지 않고, **구조(그룹/태스크 목록, 컴포넌트 참조)**만 갖도록 정리.  
- **Workflow**  
  - Template에서 “인스턴스화”할 때 생성.  
  - 실제 request_body 등은 Workflow 소속 Task에 저장.  
- 실행 이력은 항상 WorkflowRun → WorkflowVersion으로 “정의 스냅샷”과 연결.

---

## 6. 체크리스트 (현재 상태 대비)

- [ ] TaskGroup: workflow_id(또는 workflow_version_id) + Workflow association 명시
- [ ] Task: task_group_id, task_component_id + TaskGroup/TaskComponent association, **전 필드 저장**
- [ ] WorkflowRun: workflow_version_id 추가, 실행 시 버전 기록
- [ ] Workflow.data: “캐시” 또는 호환용으로만 사용, 단일 소스는 Version+TaskGroup+Task로 이전
- [ ] Task 단위 조회/이력: task 테이블 + TaskRun으로 가능
- [ ] “이 실행이 어떤 구조였는지”: WorkflowRun.workflow_version_id → WorkflowVersion으로 추적
- [ ] Template: 정의 위주, 실행 인스턴스는 Workflow + Task

이 순서대로 적용하면, 말씀하신 “역할 불명확, 참조 약함, Task 직접 저장 없음, Template/실행 혼재, 실행 기준 구조 추적 불가” 문제를 단계적으로 해소할 수 있습니다.
