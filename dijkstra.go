package main

import (
	"container/heap"
	"fmt"
	"log"
	"math/rand"
	"slices"
)

// dijkstra search
type DijkstraSearch struct {
	Frontier PriorityQueueDijkstra
	Game     *Maze
}

func (ds *DijkstraSearch) GetFrontier() []*Node {
	return ds.Frontier
}

func (ds *DijkstraSearch) Add(i *Node) {
	i.CostToGoal = i.ManhattanDistance(ds.Game.Start)
	ds.Frontier.Push(i)
	heap.Init(&ds.Frontier)
}

func (ds *DijkstraSearch) ContainsState(i *Node) bool {
	for _, x := range ds.Frontier {
		if x.State == i.State {
			return true
		}
	}

	return false
}

func (ds *DijkstraSearch) Empty() bool {
	return len(ds.Frontier) == 0
}

func (ds *DijkstraSearch) Remove() (*Node, error) {
	if len(ds.Frontier) == 0 {
		return nil, fmt.Errorf("frontier is empty")
	}

	if ds.Game.Debug {
		fmt.Println("Frontier Before Remove:")
		for _, x := range ds.Frontier {
			fmt.Println("Node: ", x.State)
		}
	}

	return heap.Pop(&ds.Frontier).(*Node), nil
}

func (ds *DijkstraSearch) Solve() {
	fmt.Println("Starting to solve maze using Dijkstra Search...")
	ds.Game.NumExplored = 0

	start := Node{State: ds.Game.Start, Parent: nil, Action: ""}
	ds.Add(&start)
	ds.Game.CurrentNode = &start

	for {
		if ds.Empty() {
			return
		}

		currentNode, err := ds.Remove()
		if err != nil {
			log.Println(err)
			return
		}

		if ds.Game.Debug {
			fmt.Println("Removed: ", currentNode.State)
			fmt.Println("---------------")
			fmt.Println()
		}

		ds.Game.CurrentNode = currentNode
		ds.Game.NumExplored += 1

		// have we found the solution?
		if ds.Game.Goal == currentNode.State {
			var actions []string
			var cells []Point

			for {
				// if we're at the starting point break the loop
				if currentNode.Parent == nil {
					break
				}

				actions = append(actions, currentNode.Action)
				cells = append(cells, currentNode.State)
				currentNode = currentNode.Parent
			}

			slices.Reverse(actions)
			slices.Reverse(cells)

			ds.Game.Solution = Solution{Actions: actions, Cells: cells}
			ds.Game.Explored = append(ds.Game.Explored, currentNode.State)

			break
		}

		ds.Game.Explored = append(ds.Game.Explored, currentNode.State)

		// animating the results
		if ds.Game.Animate {
			frameName := fmt.Sprintf("frames/frame_%04d.png", ds.Game.NumExplored)
			ds.Game.OutputImage(frameName)
		}

		for _, x := range ds.Neighbors(currentNode) {
			if !ds.ContainsState(x) {
				if !inExplored(x.State, ds.Game.Explored) {
					ds.Add(&Node{
						State:  x.State,
						Parent: currentNode,
						Action: x.Action,
					})
				}
			}
		}
	}
}

func (ds *DijkstraSearch) Neighbors(node *Node) []*Node {
	row := node.State.Row
	col := node.State.Col

	candidates := []*Node{
		{State: Point{Row: row - 1, Col: col}, Parent: node, Action: "UP"},
		{State: Point{Row: row + 1, Col: col}, Parent: node, Action: "DOWN"},
		{State: Point{Row: row, Col: col - 1}, Parent: node, Action: "LEFT"},
		{State: Point{Row: row, Col: col + 1}, Parent: node, Action: "RIGHT"},
	}

	var neighbors []*Node

	for _, x := range candidates {
		if 0 <= x.State.Row && x.State.Row < ds.Game.Height {
			if 0 <= x.State.Col && x.State.Col < ds.Game.Width {
				if !ds.Game.Walls[x.State.Row][x.State.Col].wall {
					neighbors = append(neighbors, x)
				}
			}
		}
	}

	for i := range neighbors {
		j := rand.Intn(i + 1)

		neighbors[i], neighbors[j] = neighbors[j], neighbors[i]
	}

	return neighbors
}
