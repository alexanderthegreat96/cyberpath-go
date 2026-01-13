package main

import (
	"fmt"
	"log"
	"math/rand"
	"slices"
)

// breadth first search algorithm
// actually slightly better than depth first search
type BreadthFirstSearch struct {
	Frontier []*Node
	Game     *Maze
}

func (bfs *BreadthFirstSearch) GetFrontier() []*Node {
	return bfs.Frontier
}

func (bfs *BreadthFirstSearch) Add(i *Node) {
	bfs.Frontier = append(bfs.Frontier, i)
}

func (bfs *BreadthFirstSearch) ContainsState(i *Node) bool {
	for _, x := range bfs.Frontier {
		if x.State == i.State {
			return true
		}
	}

	return false
}

func (bfs *BreadthFirstSearch) Empty() bool {
	return len(bfs.Frontier) == 0
}

func (bfs *BreadthFirstSearch) Remove() (*Node, error) {
	if len(bfs.Frontier) == 0 {
		return nil, fmt.Errorf("frontier is empty")
	}

	if bfs.Game.Debug {
		fmt.Println("Frontier Before Remove:")
		for _, x := range bfs.Frontier {
			fmt.Println("Node: ", x.State)
		}
	}

	// first node
	node := bfs.Frontier[0]

	// no memory leak here
	bfs.Frontier[0] = nil

	// reslice
	bfs.Frontier = bfs.Frontier[1:]

	return node, nil
}

func (bfs *BreadthFirstSearch) Solve() {
	fmt.Println("Starting to solve maze using Breadth First Search...")
	bfs.Game.NumExplored = 0

	start := Node{State: bfs.Game.Start, Parent: nil, Action: ""}
	bfs.Add(&start)
	bfs.Game.CurrentNode = &start

	for {
		if bfs.Empty() {
			return
		}

		currentNode, err := bfs.Remove()
		if err != nil {
			log.Println(err)
			return
		}

		if bfs.Game.Debug {
			fmt.Println("Removed: ", currentNode.State)
			fmt.Println("---------------")
			fmt.Println()
		}

		bfs.Game.CurrentNode = currentNode
		bfs.Game.NumExplored += 1

		// have we found the solution?
		if bfs.Game.Goal == currentNode.State {
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

			bfs.Game.Solution = Solution{Actions: actions, Cells: cells}
			bfs.Game.Explored = append(bfs.Game.Explored, currentNode.State)

			break
		}

		bfs.Game.Explored = append(bfs.Game.Explored, currentNode.State)

		if bfs.Game.Animate {
			frameName := fmt.Sprintf("frames/frame_%04d.png", bfs.Game.NumExplored)
			bfs.Game.OutputImage(frameName)
		}

		for _, x := range bfs.Neighbors(currentNode) {
			if !bfs.ContainsState(x) {
				if !inExplored(x.State, bfs.Game.Explored) {
					bfs.Add(&Node{
						State:  x.State,
						Parent: currentNode,
						Action: x.Action,
					})
				}
			}
		}
	}
}

func (bfs *BreadthFirstSearch) Neighbors(node *Node) []*Node {
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
		if 0 <= x.State.Row && x.State.Row < bfs.Game.Height {
			if 0 <= x.State.Col && x.State.Col < bfs.Game.Width {
				if !bfs.Game.Walls[x.State.Row][x.State.Col].wall {
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
