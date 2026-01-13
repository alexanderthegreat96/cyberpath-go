package main

import (
	"container/heap"
	"fmt"
	"log"
	"math/rand"
	"slices"
)

// flooded cells asbstraction cost
const FLOODED_COST = 100

// astar algorithm
type AstarSearch struct {
	Frontier PriorityQueueAstar
	Game     *Maze
}

func (as *AstarSearch) GetFrontier() []*Node {
	return as.Frontier
}

func (as *AstarSearch) Add(i *Node) {
	i.CostToGoal = i.ManhattanDistance(as.Game.Start)
	i.EstimatedCostToGoal = EuclideanDistance(i.State, as.Game.Goal) + float64(i.CostToGoal)

	// flooded cost
	if i.State.Water {
		i.EstimatedCostToGoal += FLOODED_COST
	}

	as.Frontier.Push(i)
	heap.Init(&as.Frontier)
}

func (as *AstarSearch) ContainsState(i *Node) bool {
	for _, x := range as.Frontier {
		if x.State == i.State {
			return true
		}
	}

	return false
}

func (as *AstarSearch) Empty() bool {
	return len(as.Frontier) == 0
}

func (as *AstarSearch) Remove() (*Node, error) {
	if len(as.Frontier) == 0 {
		return nil, fmt.Errorf("frontier is empty")
	}

	if as.Game.Debug {
		fmt.Println("Frontier Before Remove:")
		for _, x := range as.Frontier {
			fmt.Println("Node: ", x.State)
		}
	}

	return heap.Pop(&as.Frontier).(*Node), nil
}

func (as *AstarSearch) Solve() {
	fmt.Println("Starting to solve maze using Astar Search...")
	as.Game.NumExplored = 0

	start := Node{State: as.Game.Start, Parent: nil, Action: ""}
	as.Add(&start)
	as.Game.CurrentNode = &start

	for {
		if as.Empty() {
			return
		}

		currentNode, err := as.Remove()
		if err != nil {
			log.Println(err)
			return
		}

		if as.Game.Debug {
			fmt.Println("Removed: ", currentNode.State)
			fmt.Println("---------------")
			fmt.Println()
		}

		as.Game.CurrentNode = currentNode
		as.Game.NumExplored += 1

		// have we found the solution?
		if as.Game.Goal == currentNode.State {
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

			as.Game.Solution = Solution{Actions: actions, Cells: cells}
			as.Game.Explored = append(as.Game.Explored, currentNode.State)

			break
		}

		as.Game.Explored = append(as.Game.Explored, currentNode.State)

		// animating the results
		if as.Game.Animate {
			frameName := fmt.Sprintf("frames/frame_%04d.png", as.Game.NumExplored)
			as.Game.OutputImage(frameName)
		}

		for _, x := range as.Neighbors(currentNode) {
			if !as.ContainsState(x) {
				if !inExplored(x.State, as.Game.Explored) {
					as.Add(&Node{
						State:  x.State,
						Parent: currentNode,
						Action: x.Action,
					})
				}
			}
		}
	}
}

func (as *AstarSearch) Neighbors(node *Node) []*Node {
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
		if 0 <= x.State.Row && x.State.Row < as.Game.Height {
			if 0 <= x.State.Col && x.State.Col < as.Game.Width {
				if !as.Game.Walls[x.State.Row][x.State.Col].wall {
					// flooded cells
					if as.Game.Walls[x.State.Row][x.State.Col].State.Water {
						x.State.Water = true
					}
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
