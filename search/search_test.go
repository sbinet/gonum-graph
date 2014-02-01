package search_test

import (
	"math"
	"testing"

	"github.com/gonum/graph"
	"github.com/gonum/graph/concrete"
	"github.com/gonum/graph/search"
	//	"github.com/gonum/graph/set"
)

func TestSimpleAStar(t *testing.T) {
	tg, err := concrete.GenerateTileGraph("▀  ▀\n▀▀ ▀\n▀▀ ▀\n▀▀ ▀")
	if err != nil {
		t.Fatal("Couldn't generate tilegraph")
	}

	path, cost, _ := search.AStar(concrete.GonumNode(1), concrete.GonumNode(14), tg, nil, nil)
	if math.Abs(cost-4.0) > .00001 {
		t.Errorf("A* reports incorrect cost for simple tilegraph search")
	}

	if path == nil {
		t.Fatalf("A* fails to find path for for simple tilegraph search")
	} else {
		correctPath := []int{1, 2, 6, 10, 14}
		if len(path) != len(correctPath) {
			t.Fatalf("Astar returns wrong length path for simple tilegraph search")
		}
		for i, node := range path {
			if node.ID() != correctPath[i] {
				t.Errorf("Astar returns wrong path at step", i, "got:", node, "actual:", correctPath[i])
			}
		}
	}
}

func TestBiggerAStar(t *testing.T) {
	tg := concrete.NewTileGraph(3, 3, true)

	path, cost, _ := search.AStar(concrete.GonumNode(0), concrete.GonumNode(8), tg, nil, nil)

	if math.Abs(cost-4.0) > .00001 || !search.IsPath(path, tg) {
		t.Error("Non-optimal or impossible path found for 3x3 grid")
	}

	tg = concrete.NewTileGraph(1000, 1000, true)
	path, cost, _ = search.AStar(concrete.GonumNode(00), concrete.GonumNode(999*1000+999), tg, nil, nil)
	if !search.IsPath(path, tg) || cost != 1998.0 {
		t.Error("Non-optimal or impossible path found for 100x100 grid; cost:", cost, "path:\n"+tg.PathString(path))
	}
}

func TestObstructedAStar(t *testing.T) {
	tg := concrete.NewTileGraph(10, 10, true)

	// Creates a partial "wall" down the middle row with a gap down the left side
	tg.SetPassability(4, 1, false)
	tg.SetPassability(4, 2, false)
	tg.SetPassability(4, 3, false)
	tg.SetPassability(4, 4, false)
	tg.SetPassability(4, 5, false)
	tg.SetPassability(4, 6, false)
	tg.SetPassability(4, 7, false)
	tg.SetPassability(4, 8, false)
	tg.SetPassability(4, 9, false)

	rows, cols := tg.Dimensions()
	path, cost1, expanded := search.AStar(concrete.GonumNode(5), tg.CoordsToNode(rows-1, cols-1), tg, nil, nil)

	if !search.IsPath(path, tg) {
		t.Error("Path doesn't exist in obstructed graph")
	}

	ManhattanHeuristic := func(n1, n2 graph.Node) float64 {
		id1, id2 := n1.ID(), n2.ID()
		r1, c1 := tg.IDToCoords(id1)
		r2, c2 := tg.IDToCoords(id2)

		return math.Abs(float64(r1)-float64(r2)) + math.Abs(float64(c1)-float64(c2))
	}

	path, cost2, expanded2 := search.AStar(concrete.GonumNode(5), tg.CoordsToNode(rows-1, cols-1), tg, nil, ManhattanHeuristic)
	if !search.IsPath(path, tg) {
		t.Error("Path doesn't exist when using heuristic on obstructed graph")
	}

	if math.Abs(cost1-cost2) > .00001 {
		t.Error("Cost when using admissible heuristic isn't approximately equal to cost without it")
	}

	if expanded2 > expanded {
		t.Error("Using admissible, consistent heuristic expanded more nodes than null heuristic (possible, but unlikely -- suggests an error somewhere)")
	}

}

func TestNoPathAStar(t *testing.T) {
	tg := concrete.NewTileGraph(5, 5, true)

	// Creates a "wall" down the middle row
	tg.SetPassability(2, 0, false)
	tg.SetPassability(2, 1, false)
	tg.SetPassability(2, 2, false)
	tg.SetPassability(2, 3, false)
	tg.SetPassability(2, 4, false)

	rows, _ := tg.Dimensions()
	path, _, _ := search.AStar(tg.CoordsToNode(0, 2), tg.CoordsToNode(rows-1, 2), tg, nil, nil)

	if len(path) > 0 { // Note that a nil slice will return len of 0, this won't panic
		t.Error("A* finds path where none exists")
	}
}

/* BUG Test fails, presumably indicating a bug.
func TestDStarLite(t *testing.T) {
	tg := graph.NewTileGraph(10, 10, true)
	realGraph := graph.NewTileGraph(10, 10, true)
	realGraph.SetPassability(4, 1, false)
	realGraph.SetPassability(4, 2, false)
	realGraph.SetPassability(4, 3, false)
	realGraph.SetPassability(4, 4, false)
	realGraph.SetPassability(4, 5, false)
	realGraph.SetPassability(4, 6, false)
	realGraph.SetPassability(4, 7, false)
	realGraph.SetPassability(4, 8, false)
	realGraph.SetPassability(4, 9, false)

	rows, cols := tg.Dimensions()
	dStarInstance := graph.InitDStar(graph.GonumNode(5), tg.CoordsToNode(rows-1, cols-1), tg, nil, nil)

	path := []graph.Node{graph.GonumNode(5)}

	var succ graph.Node
	var err error
	for succ, err = dStarInstance.Step(); err != nil && succ != nil && succ.ID() != tg.CoordsToID(rows-1, cols-1); succ, err = dStarInstance.Step() {
		path = append(path, succ)

		knownSuccs := tg.Neighbors(succ)
		realSuccs := realGraph.Neighbors(succ)
		knownSet := set.NewSet()
		for _, k := range knownSuccs {
			knownSet.Add(k)
		}
		realSet := set.NewSet()
		for _, k := range realSuccs {
			realSet.Add(k)
		}

		knownSet.Diff(knownSet, realSet)
		updatedEdges := []graph.Edge{}
		for _, toRemove := range knownSet.Elements() {
			node := toRemove.(graph.Node)
			updatedEdges = append(updatedEdges, graph.GonumEdge{H: succ, T: node})
			row, col := tg.IDToCoords(node.ID())
			tg.SetPassability(row, col, false)
		}

		dStarInstance.Update(nil, updatedEdges)
	}

	if succ == nil || err != nil {
		t.Error("Got erroneous error: %s\nPath before error encountered: %#v", err, path)
	}
}
*/