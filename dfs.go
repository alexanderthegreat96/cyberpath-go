package main

import (
	"fmt"
	"math/rand"
	"slices"
)

// depth first search algorithm
type DepthFirstSearch struct {
	Frontier []*Node
	Game     *Maze
}

func (dfs *DepthFirstSearch) GetFrontier() []*Node {
	return dfs.Frontier
}

func (dfs *DepthFirstSearch) Add(i *Node) {
	// last in first out method
	// LIFO
	dfs.Frontier = append(dfs.Frontier, i)
}

func (dfs *DepthFirstSearch) ContainsState(i *Node) bool {
	for _, x := range dfs.Frontier {
		if x.State == i.State {
			return true
		}
	}

	return false
}

func (dfs *DepthFirstSearch) Empty() bool {
	return len(dfs.Frontier) == 0
}

func (dfs *DepthFirstSearch) Remove() (*Node, error) {
	if len(dfs.Frontier) > 0 {
		if dfs.Game.Debug {
			fmt.Println("Frontier Before Remove:")

			for _, x := range dfs.Frontier {
				fmt.Println("Node: ", x.State)
			}
		}

		node := dfs.Frontier[len(dfs.Frontier)-1]
		dfs.Frontier = dfs.Frontier[:len(dfs.Frontier)-1]
		return node, nil
	}

	return nil, fmt.Errorf("frontier is empty")
}

func (dfs *DepthFirstSearch) Solve() {
	fmt.Println("Starting to solve maze using Depth First Search...")

	dfs.Game.NumExplored = 0

	start := Node{
		State:  dfs.Game.Start,
		Parent: nil,
		Action: "",
	}

	dfs.Add(&start)
	dfs.Game.CurrentNode = &start

	for {
		if dfs.Empty() {
			return
		}

		currentNode, err := dfs.Remove()
		if err != nil {
			fmt.Println(err)
			return
		}

		if dfs.Game.Debug {
			fmt.Println("Removed: ", currentNode.State)
			fmt.Println("--------------")
			fmt.Println("")
		}

		dfs.Game.CurrentNode = currentNode
		dfs.Game.NumExplored += 1

		// have we found the solution?
		if dfs.Game.Goal == currentNode.State {
			var actions []string
			var cells []Point

			for {
				if currentNode.Parent == nil {
					break
				}

				actions = append(actions, currentNode.Action)
				cells = append(cells, currentNode.State)
				currentNode = currentNode.Parent
			}

			slices.Reverse(actions)
			slices.Reverse(cells)

			dfs.Game.Solution = Solution{
				Actions: actions,
				Cells:   cells,
			}

			dfs.Game.Explored = append(dfs.Game.Explored, currentNode.State)
			break
		}

		dfs.Game.Explored = append(dfs.Game.Explored, currentNode.State)

		// build animation frame is appropriate
		if dfs.Game.Animate {
			frameName := fmt.Sprintf("frames/frame_%04d.png", dfs.Game.NumExplored)
			dfs.Game.OutputImage(frameName)
		}

		for _, x := range dfs.Neighbors(currentNode) {
			if !dfs.ContainsState(x) {
				if !inExplored(x.State, dfs.Game.Explored) {
					dfs.Add(&Node{
						State:  x.State,
						Parent: currentNode,
						Action: x.Action,
					})
				}
			}
		}
	}
}

func (dfs *DepthFirstSearch) Neighbors(node *Node) []*Node {
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
		if 0 <= x.State.Row && x.State.Row < dfs.Game.Height {
			if 0 <= x.State.Col && x.State.Col < dfs.Game.Width {
				if !dfs.Game.Walls[x.State.Row][x.State.Col].wall {
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
