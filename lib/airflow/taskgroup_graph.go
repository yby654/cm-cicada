package airflow

import (
	"errors"
	"strings"

	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
	"github.com/jollaman999/utils/logger"
)

// groupDepEdge is one upstream -> downstream edge of the task group graph.
// Explicit edges come from TaskGroup.Dependencies and are emitted into the
// group's METADATA.yml, so gusty turns them into real Airflow edges between the
// groups' join nodes. Implied edges are inferred from task dependencies that
// cross a group boundary; they order the tasks but leave no group edge behind.
type groupDepEdge struct {
	to       string
	explicit bool
}

type taskGroupGraph struct {
	groupNames []string
	adjacency  map[string][]groupDepEdge
	// boundaryLeaks lists task edges that cross a group boundary that no task
	// group dependency declares, in "A.a1 -> B.b1" form.
	boundaryLeaks []string
}

// buildTaskGroupGraph folds the explicit group dependencies and the task
// dependencies crossing a boundary into a single graph.
func buildTaskGroupGraph(taskGroups []model.TaskGroup) (*taskGroupGraph, error) {
	graph := &taskGroupGraph{
		groupNames: make([]string, 0, len(taskGroups)),
		adjacency:  make(map[string][]groupDepEdge),
	}

	groupNames := make(map[string]struct{}, len(taskGroups))
	groupNameByTask := make(map[string]string)

	for _, tg := range taskGroups {
		if _, exists := groupNames[tg.Name]; exists {
			return nil, errors.New("Duplicated task group name: " + tg.Name)
		}
		groupNames[tg.Name] = struct{}{}
		graph.groupNames = append(graph.groupNames, tg.Name)

		for _, t := range tg.Tasks {
			groupNameByTask[t.Name] = tg.Name
		}
	}

	explicitEdges := make(map[string]struct{})
	for _, tg := range taskGroups {
		for _, dep := range tg.Dependencies {
			if dep == tg.Name {
				return nil, errors.New("cycle dependency found in task group " + tg.Name)
			}
			if _, exists := groupNames[dep]; !exists {
				return nil, errors.New("wrong task group dependency found in " + tg.Name + " (" + dep + ")")
			}

			edgeKey := dep + " -> " + tg.Name
			if _, exists := explicitEdges[edgeKey]; exists {
				continue
			}
			explicitEdges[edgeKey] = struct{}{}
			graph.adjacency[dep] = append(graph.adjacency[dep], groupDepEdge{to: tg.Name, explicit: true})
		}
	}

	impliedEdges := make(map[string]struct{})
	for _, tg := range taskGroups {
		for _, t := range tg.Tasks {
			for _, dep := range t.Dependencies {
				upstreamGroup, exists := groupNameByTask[dep]
				if !exists || upstreamGroup == tg.Name {
					continue
				}

				edgeKey := upstreamGroup + " -> " + tg.Name
				if _, declared := explicitEdges[edgeKey]; !declared {
					graph.boundaryLeaks = append(graph.boundaryLeaks,
						upstreamGroup+"."+dep+" -> "+tg.Name+"."+t.Name)
				}
				if _, seen := impliedEdges[edgeKey]; seen {
					continue
				}
				impliedEdges[edgeKey] = struct{}{}
				graph.adjacency[upstreamGroup] = append(graph.adjacency[upstreamGroup],
					groupDepEdge{to: tg.Name, explicit: false})
			}
		}
	}

	return graph, nil
}

// validateTaskGroupBoundaries rejects a group graph that cannot run: an explicit
// group edge combined with a task edge running the other way deadlocks the DAG,
// because gusty wires the group's upstream_join_id ahead of every task inside it.
//
// A cycle made of implied edges alone is not an Airflow cycle (no group edge is
// emitted for it), so it is reported as a boundary leak instead of an error.
func validateTaskGroupBoundaries(taskGroups []model.TaskGroup) error {
	graph, err := buildTaskGroupGraph(taskGroups)
	if err != nil {
		return err
	}

	if err := graph.checkCycles(); err != nil {
		return err
	}

	if len(graph.boundaryLeaks) > 0 {
		logger.Println(logger.WARN, true,
			"TaskGroup boundary leak: task dependencies cross a group boundary without a "+
				"task group dependency declaring it ("+strings.Join(graph.boundaryLeaks, ", ")+")")
	}

	return nil
}

const (
	unvisited = 0
	visiting  = 1
	visited   = 2
)

// checkCycles walks the combined graph and fails only on cycles that contain at
// least one explicit group edge, i.e. the ones Airflow would actually deadlock
// on.
func (g *taskGroupGraph) checkCycles() error {
	visitState := make(map[string]int, len(g.groupNames))
	var pathGroups []string
	var pathEdges []groupDepEdge

	cycleHasExplicitEdge := func(start string, closing groupDepEdge) bool {
		if closing.explicit {
			return true
		}
		for i, name := range pathGroups {
			if name != start {
				continue
			}
			for _, edge := range pathEdges[i:] {
				if edge.explicit {
					return true
				}
			}
			break
		}
		return false
	}

	var dfs func(groupName string) error
	dfs = func(groupName string) error {
		visitState[groupName] = visiting
		pathGroups = append(pathGroups, groupName)

		for _, edge := range g.adjacency[groupName] {
			switch visitState[edge.to] {
			case visiting:
				if cycleHasExplicitEdge(edge.to, edge) {
					return errors.New("cycle dependency found in task group " + edge.to)
				}
				logger.Println(logger.WARN, true,
					"TaskGroup boundary leak: task dependencies form a cycle between task groups "+
						groupName+" and "+edge.to+" (no task group dependency is emitted for it)")
			case unvisited:
				pathEdges = append(pathEdges, edge)
				if err := dfs(edge.to); err != nil {
					return err
				}
				pathEdges = pathEdges[:len(pathEdges)-1]
			}
		}

		visitState[groupName] = visited
		pathGroups = pathGroups[:len(pathGroups)-1]
		return nil
	}

	for _, groupName := range g.groupNames {
		if visitState[groupName] == unvisited {
			if err := dfs(groupName); err != nil {
				return err
			}
		}
	}

	return nil
}
