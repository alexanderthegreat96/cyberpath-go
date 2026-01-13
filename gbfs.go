package main

import (
	"container/heap"
	"fmt"
	"log"
	"math/rand"
	"slices"
)

// greedy best first search
type GreedyBestFirstSearch struct {
	Frontier PriorityQueueGBFS
	Game     *Maze
}

func (gbfs *GreedyBestFirstSearch) GetFrontier() []*Node {
	return gbfs.Frontier
}

func (gbfs *GreedyBestFirstSearch) Add(i *Node) {
	// dijkstra uses the start instead of the goal / end
	i.CostToGoal = i.ManhattanDistance(gbfs.Game.Goal)
	gbfs.Frontier.Push(i)
	heap.Init(&gbfs.Frontier)
}

func (gbfs *GreedyBestFirstSearch) ContainsState(i *Node) bool {
	for _, x := range gbfs.Frontier {
		if x.State == i.State {
			return true
		}
	}

	return false
}

func (gbfs *GreedyBestFirstSearch) Empty() bool {
	return len(gbfs.Frontier) == 0
}

func (gbfs *GreedyBestFirstSearch) Remove() (*Node, error) {
	if len(gbfs.Frontier) == 0 {
		return nil, fmt.Errorf("frontier is empty")
	}

	if gbfs.Game.Debug {
		fmt.Println("Frontier Before Remove:")
		for _, x := range gbfs.Frontier {
			fmt.Println("Node: ", x.State)
		}
	}

	return heap.Pop(&gbfs.Frontier).(*Node), nil
}

func (gbfs *GreedyBestFirstSearch) Solve() {
	fmt.Println("Starting to solve maze using Greedy Best First Search...")
	gbfs.Game.NumExplored = 0

	start := Node{State: gbfs.Game.Start, Parent: nil, Action: ""}
	gbfs.Add(&start)
	gbfs.Game.CurrentNode = &start

	for {
		if gbfs.Empty() {
			return
		}

		currentNode, err := gbfs.Remove()
		if err != nil {
			log.Println(err)
			return
		}

		if gbfs.Game.Debug {
			fmt.Println("Removed: ", currentNode.State)
			fmt.Println("---------------")
			fmt.Println()
		}

		gbfs.Game.CurrentNode = currentNode
		gbfs.Game.NumExplored += 1

		// have we found the solution?
		if gbfs.Game.Goal == currentNode.State {
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

			gbfs.Game.Solution = Solution{Actions: actions, Cells: cells}
			gbfs.Game.Explored = append(gbfs.Game.Explored, currentNode.State)

			break
		}

		gbfs.Game.Explored = append(gbfs.Game.Explored, currentNode.State)

		// animating the results
		if gbfs.Game.Animate {
			frameName := fmt.Sprintf("frames/frame_%04d.png", gbfs.Game.NumExplored)
			gbfs.Game.OutputImage(frameName)
		}

		for _, x := range gbfs.Neighbors(currentNode) {
			if !gbfs.ContainsState(x) {
				if !inExplored(x.State, gbfs.Game.Explored) {
					gbfs.Add(&Node{
						State:  x.State,
						Parent: currentNode,
						Action: x.Action,
					})
				}
			}
		}
	}
}

func (gbfs *GreedyBestFirstSearch) Neighbors(node *Node) []*Node {
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
		if 0 <= x.State.Row && x.State.Row < gbfs.Game.Height {
			if 0 <= x.State.Col && x.State.Col < gbfs.Game.Width {
				if !gbfs.Game.Walls[x.State.Row][x.State.Col].wall {
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
