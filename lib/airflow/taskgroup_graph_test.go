package airflow

import (
	"strings"
	"testing"

	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
	"gopkg.in/yaml.v3"
)

func taskGroup(name string, dependencies []string, tasks ...model.Task) model.TaskGroup {
	return model.TaskGroup{
		ID:           name + "-id",
		Name:         name,
		Dependencies: dependencies,
		Tasks:        tasks,
	}
}

func task(name string, dependencies ...string) model.Task {
	return model.Task{
		ID:           name + "-id",
		Name:         name,
		Dependencies: dependencies,
	}
}

func TestValidateTaskGroupBoundariesAcceptsLinearGroupChain(t *testing.T) {
	groups := []model.TaskGroup{
		taskGroup("A", nil, task("a1"), task("a2", "a1")),
		taskGroup("B", []string{"A"}, task("b1")),
		taskGroup("C", []string{"B"}, task("c1")),
	}

	if err := validateTaskGroupBoundaries(groups); err != nil {
		t.Fatalf("expected a linear group chain to validate, got: %v", err)
	}
}

// A task edge running against an explicit group edge deadlocks the DAG: gusty
// wires every task of B after A's downstream join, so A.a1 can never wait on
// B.b1.
func TestValidateTaskGroupBoundariesRejectsTaskEdgeAgainstGroupEdge(t *testing.T) {
	groups := []model.TaskGroup{
		taskGroup("A", nil, task("a1", "b1")),
		taskGroup("B", []string{"A"}, task("b1")),
	}

	err := validateTaskGroupBoundaries(groups)
	if err == nil {
		t.Fatal("expected a cycle error when a task edge contradicts a group edge")
	}
	if !strings.Contains(err.Error(), "cycle dependency found in task group") {
		t.Fatalf("expected a cycle error, got: %v", err)
	}
}

// Mutual task edges across two groups emit no group edge at all, so Airflow
// stays acyclic. It is a boundary leak, not a failure.
func TestValidateTaskGroupBoundariesAllowsImpliedOnlyCycle(t *testing.T) {
	groups := []model.TaskGroup{
		taskGroup("A", nil, task("a1"), task("a2", "b1")),
		taskGroup("B", nil, task("b1", "a1")),
	}

	if err := validateTaskGroupBoundaries(groups); err != nil {
		t.Fatalf("expected implied-only cycle to be tolerated, got: %v", err)
	}
}

func TestValidateTaskGroupBoundariesRejectsGroupCycle(t *testing.T) {
	groups := []model.TaskGroup{
		taskGroup("A", []string{"C"}, task("a1")),
		taskGroup("B", []string{"A"}, task("b1")),
		taskGroup("C", []string{"B"}, task("c1")),
	}

	if err := validateTaskGroupBoundaries(groups); err == nil {
		t.Fatal("expected a cycle error for A -> B -> C -> A")
	}
}

func TestValidateTaskGroupBoundariesRejectsSelfDependency(t *testing.T) {
	groups := []model.TaskGroup{
		taskGroup("A", []string{"A"}, task("a1")),
	}

	if err := validateTaskGroupBoundaries(groups); err == nil {
		t.Fatal("expected a cycle error for a self-dependent group")
	}
}

func TestValidateTaskGroupBoundariesRejectsUnknownGroupDependency(t *testing.T) {
	groups := []model.TaskGroup{
		taskGroup("A", []string{"Nope"}, task("a1")),
	}

	err := validateTaskGroupBoundaries(groups)
	if err == nil {
		t.Fatal("expected an error for a dependency on a group that does not exist")
	}
	if !strings.Contains(err.Error(), "wrong task group dependency") {
		t.Fatalf("expected a wrong-dependency error, got: %v", err)
	}
}

func TestValidateTaskGroupBoundariesRejectsDuplicatedGroupName(t *testing.T) {
	groups := []model.TaskGroup{
		taskGroup("A", nil, task("a1")),
		taskGroup("A", nil, task("a2")),
	}

	if err := validateTaskGroupBoundaries(groups); err == nil {
		t.Fatal("expected an error for duplicated task group names")
	}
}

// Existing single-group workflows and workflows whose groups only carry
// task-level edges must keep validating.
func TestValidateTaskGroupBoundariesAcceptsLegacyGraphs(t *testing.T) {
	single := []model.TaskGroup{
		taskGroup("only", nil, task("t1"), task("t2", "t1"), task("t3", "t2")),
	}
	if err := validateTaskGroupBoundaries(single); err != nil {
		t.Fatalf("expected a single-group workflow to validate, got: %v", err)
	}

	leaking := []model.TaskGroup{
		taskGroup("A", nil, task("a1")),
		taskGroup("B", nil, task("b1", "a1")),
	}
	if err := validateTaskGroupBoundaries(leaking); err != nil {
		t.Fatalf("expected an undeclared cross-group task edge to validate, got: %v", err)
	}
}

// Group dependencies are written with the directory names gusty turns into
// group_ids, not with the human-facing group names.
func TestBuildTaskGroupDirNameByNameUsesGroupID(t *testing.T) {
	workflow := &model.Workflow{
		Data: model.Data{
			TaskGroups: []model.TaskGroup{
				{ID: "uuid-a", Name: "A"},
				{ID: "", Name: "B"},
			},
		},
	}

	dirNameByName := buildTaskGroupDirNameByName(workflow)
	if dirNameByName["A"] != "uuid-a" {
		t.Fatalf("expected group A to resolve to its ID, got %q", dirNameByName["A"])
	}
	if dirNameByName["B"] != "B" {
		t.Fatalf("expected an ID-less group to fall back to its name, got %q", dirNameByName["B"])
	}

	resolved := resolveDependencies([]string{"A", "B"}, dirNameByName)
	if len(resolved) != 2 || resolved[0] != "uuid-a" || resolved[1] != "B" {
		t.Fatalf("unexpected resolved group dependencies: %v", resolved)
	}
}

// gusty rewrites the task_id of every spec to the YAML file's base name, so a
// task's file name and the value its dependents point at must come from one
// place. They used to come from two (t.ID for the file, the DB task key for
// dependencies), which silently dropped edges whenever the two diverged.
func TestResolveAirflowIDMatchesResolvedDependencies(t *testing.T) {
	// A workflow update used to hand out a second UUID as the task key.
	taskAirflowIDByName := map[string]string{"t1": "key-uuid-1", "t2": "key-uuid-2"}

	fileName := resolveAirflowID("t1", taskAirflowIDByName)
	dependents := resolveDependencies([]string{"t1"}, taskAirflowIDByName)

	if len(dependents) != 1 || dependents[0] != fileName {
		t.Fatalf("dependency %v does not match the task file name %q", dependents, fileName)
	}

	// Names with no mapping fall back to the name itself, on both sides.
	unmapped := resolveAirflowID("t3", taskAirflowIDByName)
	if unmapped != "t3" {
		t.Fatalf("expected an unmapped name to fall back to itself, got %q", unmapped)
	}
	if got := resolveDependencies([]string{"t3"}, taskAirflowIDByName); got[0] != unmapped {
		t.Fatalf("expected the same fallback for dependencies, got %v", got)
	}

	// An empty mapping must not produce an empty file name.
	if got := resolveAirflowID("t4", map[string]string{"t4": ""}); got != "t4" {
		t.Fatalf("expected an empty mapping to fall back to the name, got %q", got)
	}
}

// The group METADATA.yml gusty reads: dependencies carry directory names, and
// prefix_group_id is pinned so a hand edit cannot re-enable task_id prefixing.
func TestBuildTaskGroupMetadataYAML(t *testing.T) {
	dirNameByName := map[string]string{"A": "uuid-a", "B": "uuid-b"}

	withDeps, err := yaml.Marshal(buildTaskGroupMetadata(
		model.TaskGroup{ID: "uuid-b", Name: "B", Description: "group B", Dependencies: []string{"A"}},
		dirNameByName,
	))
	if err != nil {
		t.Fatalf("failed to marshal group metadata: %v", err)
	}

	expected := "tooltip: group B\ntask_display_name: B\nprefix_group_id: false\ndependencies:\n    - uuid-a\n"
	if string(withDeps) != expected {
		t.Fatalf("unexpected group metadata YAML:\n%s\nwant:\n%s", withDeps, expected)
	}

	withoutDeps, err := yaml.Marshal(buildTaskGroupMetadata(
		model.TaskGroup{ID: "uuid-a", Name: "A", Description: "group A"},
		dirNameByName,
	))
	if err != nil {
		t.Fatalf("failed to marshal group metadata: %v", err)
	}

	if strings.Contains(string(withoutDeps), "dependencies") {
		t.Fatalf("expected no dependencies key for a group without dependencies:\n%s", withoutDeps)
	}
	if !strings.Contains(string(withoutDeps), "prefix_group_id: false") {
		t.Fatalf("expected prefix_group_id to be pinned:\n%s", withoutDeps)
	}
}
