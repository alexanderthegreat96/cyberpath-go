package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// search algorithms to implement
// DFS = Depth First Search
// BFS = Breadth First Search
// GBFS = Greedy Best First Search
// ASTAR = A* -> Graph Transversal Path Finding Algorithm
// DIJKSTRA = Darkstraz -> Algorithm for finding the shortest path

const (
	DFS      = iota // uninformed
	BFS             // uninformed
	GBFS            // informed
	ASTAR           // informed
	DIJKSTRA        // uninformed
	UNKNOWN
)

// simple x,y coordinates
type Point struct {
	Row   int
	Col   int
	Water bool
}

// keeps track of potential nodes
// that are walls and cannot be explored
// if wall is true  -> it can be explored
// if wall is false -> then it's an actual wall
type Wall struct {
	State Point
	wall  bool
}

// keep track of positions
type Node struct {
	index               int
	State               Point
	Parent              *Node
	Action              string
	CostToGoal          int
	EstimatedCostToGoal float64
}

func (n *Node) ManhattanDistance(goal Point) int {
	return Abs(n.State.Row-goal.Row) + Abs(n.State.Col-goal.Col)
}

// keeps track of all the info required to complete the maze
type Maze struct {
	Height      int
	Width       int
	Start       Point
	Goal        Point
	Walls       [][]Wall
	CurrentNode *Node
	Solution    Solution
	Explored    []Point
	Steps       int
	NumExplored int
	Debug       bool
	SearchType  int
	Animate     bool
	OutputDir   string // e.g. "results/dfs_run_1"
	FrameDir    string // e.g. "results/dfs_run_1/frames"
}

// how did we achieve movement through the maze
type Solution struct {
	Actions []string
	Cells   []Point
}

func (g *Maze) Load(fileName string) error {
	f, err := os.Open(fileName)
	if err != nil {
		return fmt.Errorf("error opening file: %s. err: %s", fileName, err.Error())
	}
	defer f.Close()

	var fileContents []string
	scanner := bufio.NewScanner(f)
	maxWidth := 0

	for scanner.Scan() {
		line := scanner.Text()
		if len(line) > maxWidth {
			maxWidth = len(line)
		}
		fileContents = append(fileContents, line)
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	g.Height = len(fileContents)
	g.Width = maxWidth
	g.Walls = make([][]Wall, g.Height)

	foundStart, foundEnd := false, false

	for i, line := range fileContents {
		g.Walls[i] = make([]Wall, g.Width)

		for j := 0; j < g.Width; j++ {
			var char byte = '#'
			if j < len(line) {
				char = line[j]
			}

			var isWall bool
			var isWater bool

			switch char {
			case 'A', 'S':
				g.Start = Point{Row: i, Col: j}
				foundStart = true
				isWall = false
			case 'B', 'E':
				g.Goal = Point{Row: i, Col: j}
				foundEnd = true
				isWall = false
			case 'w', 'W':
				isWall = false
				isWater = true
			case ' ':
				isWall = false
			case '#':
				isWall = true
			default:
				isWall = true
			}

			g.Walls[i][j] = Wall{
				State: Point{Row: i, Col: j, Water: isWater},
				wall:  isWall,
			}
		}
	}

	if !foundStart || !foundEnd {
		return errors.New("maze must contain both a start (A) and a goal (B)")
	}

	return nil
}

func printHeader(maze string, algo string) {
	cyan := "\033[36m"
	gray := "\033[90m"
	reset := "\033[0m"

	banner := `
█▀▀ █▄█ █▄▄ █▀▀ █▀█ █▀█ ▄▀█ ▀█▀ █░█
█▄▄ ░█░ █▄█ ██▄ █▀▄ █▀▀ █▀█ ░█░ █▀█
───────────────────────────────────`

	fmt.Printf("%s%s%s\n", cyan, banner, reset)
	fmt.Printf("%s%-12s%s %s\n", gray, "CORE_STATUS:", reset, "ACTIVE")
	fmt.Printf("%s%-12s%s %s\n", gray, "TARGET_FILE:", reset, maze)
	fmt.Printf("%s%-12s%s %s\n", gray, "LOGIC_CORE :", reset, strings.ToUpper(algo))
	fmt.Println(gray + strings.Repeat("━", 45) + reset)
}

func main() {
	var m Maze
	var mazeFile, searchType string
	var wipeResults bool

	flag.StringVar(&mazeFile, "file", "mazes/maze.txt", "Path to the maze text file")
	flag.StringVar(&searchType, "search", "dfs", "Search algorithm: [dfs, bfs, ds]")
	flag.BoolVar(&m.Debug, "debug", false, "Enable verbose logging")
	flag.BoolVar(&m.Animate, "animate", false, "Generate an APNG animation")
	flag.BoolVar(&wipeResults, "clear", false, "Wipe Results / Clear results dir")
	flag.Parse()

	fmt.Print("\033[H\033[2J")

	printHeader(mazeFile, searchType)

	m.SearchType = getSearchType(searchType)
	if m.SearchType == UNKNOWN {
		fmt.Printf("\033[31m[!] INVALID_PROTOCOL: '%s' is not a recognized algorithm.\033[0m\n", searchType)
		os.Exit(1)
	}

	if _, err := os.Stat(mazeFile); os.IsNotExist(err) {
		fmt.Printf("\033[31m[!] CRITICAL_ERROR: File '%s' not found.\033[0m\n", mazeFile)
		os.Exit(1)
	}

	// clear results if required
	if wipeResults {
		fmt.Printf("%-25s", "Wiping Results...")

		err := EmptyContents("results")
		if err != nil {
			fmt.Printf("\r%-25s [\033[31mFAIL\033[0m]\n", "Wiping Results...")
			fmt.Printf("\033[31m   -> Error: %s\033[0m\n", err.Error())
		} else {
			fmt.Printf("\r%-25s [\033[32m OK \033[0m]\n", "Wiping Results...")
		}
	}

	timestamp := time.Now().Format("20060102_150405")
	m.OutputDir = filepath.Join("results", fmt.Sprintf("%s_%s", strings.ToLower(searchType), timestamp))
	m.FrameDir = filepath.Join(m.OutputDir, "frames")

	if err := os.MkdirAll(m.FrameDir, 0755); err != nil {
		log.Fatalf("\033[31m[!] ACCESS_DENIED: Directory creation failed: %v\033[0m\n", err)
	}

	fmt.Printf("%-25s", "Initializing Grid...")
	if err := m.Load(mazeFile); err != nil {
		fmt.Printf("\r%-25s [\033[31mFAIL\033[0m]\n", "Initializing Grid...")
		os.Exit(1)
	}
	fmt.Printf("\r%-25s [\033[32m OK \033[0m]\n", "Initializing Grid...")

	fmt.Printf("%-25s", "Executing Search...")
	startTime := time.Now()

	switch strings.ToLower(searchType) {
	case "dfs":
		solveDFS(&m)
	case "bfs":
		solveBFS(&m)
	case "ds", "dijkstra":
		solveDijkstra(&m)
	case "gbfs":
		solveGBFS(&m)
	case "astar":
		solveAstar(&m)
	default:
		fmt.Printf("\r%-25s [\033[31mINVALID\033[0m]\n", "Executing Search...")
		os.Exit(1)
	}
	fmt.Printf("\r%-25s [\033[32mDONE\033[0m]\n", "Executing Search...")

	duration := time.Since(startTime)

	if len(m.Solution.Actions) > 0 {
		fmt.Printf("\n\033[36m// SOLUTION_METRICS_FOUND //\033[0m\n")
		m.PrintMaze()

		fmt.Println("\033[90m" + strings.Repeat("─", 45) + "\033[0m")
		fmt.Printf("%-18s %d units\n", "PATH_LENGTH:", len(m.Solution.Cells))
		fmt.Printf("%-18s %d units\n", "NODES_VISITED:", len(m.Explored))
		fmt.Printf("%-18s %v\n", "PROCESS_TIME:", duration)
		fmt.Printf("%-18s %s\n", "STORAGE_ROOT:", m.OutputDir)
		fmt.Println("\033[90m" + strings.Repeat("─", 45) + "\033[0m")

		m.OutputImage()
	} else {
		fmt.Println("\n\033[31m// PATH_REJECTED: ENDPOINT_UNREACHABLE //\033[0m")
	}

	if m.Animate {
		fmt.Printf("%-25s", "Compiling Animation...")
		m.OutputAnimateImage()
		fmt.Printf("\r%-25s [\033[32mDONE\033[0m]\n", "Compiling Animation...")
	}

	fmt.Println("\n\033[90m[SESSION_END]\033[0m")
}

func solveDFS(m *Maze) {
	fmt.Println("solving using dfs")
	var s DepthFirstSearch

	s.Game = m
	fmt.Println("Goal is: ", s.Game.Goal)
	s.Solve()
}

func solveBFS(m *Maze) {
	var s BreadthFirstSearch

	s.Game = m
	fmt.Println("Goal is: ", s.Game.Goal)
	s.Solve()
}

func solveDijkstra(m *Maze) {
	var s DijkstraSearch

	s.Game = m
	fmt.Println("Goal is: ", s.Game.Goal)
	s.Solve()
}

func solveGBFS(m *Maze) {
	var s GreedyBestFirstSearch

	s.Game = m
	fmt.Println("Goal is: ", s.Game.Goal)
	s.Solve()
}

func solveAstar(m *Maze) {
	var s AstarSearch

	s.Game = m
	fmt.Println("Goal is: ", s.Game.Goal)
	s.Solve()
}

func getSearchType(algo string) int {
	switch strings.ToLower(algo) {
	case "dfs":
		return DFS
	case "bfs":
		return BFS
	case "gbfs":
		return GBFS
	case "astar":
		return ASTAR
	case "ds", "dijkstra":
		return DIJKSTRA
	default:
		return UNKNOWN
	}
}

func (g *Maze) PrintMaze() {
	for r, row := range g.Walls {
		for c, col := range row {
			if col.wall {
				fmt.Print("█")
			} else if col.State.Water {
				fmt.Print("F")
			} else if g.Start.Row == col.State.Row && g.Start.Col == col.State.Col {
				fmt.Print("A")
			} else if g.Goal.Row == col.State.Row && g.Goal.Col == col.State.Col {
				fmt.Print("B")
			} else if g.inSolution(Point{r, c, false}) {
				fmt.Print("*")
			} else {
				fmt.Print(" ")
			}
		}

		fmt.Println()
	}
}

func (g *Maze) inSolution(x Point) bool {
	for _, step := range g.Solution.Cells {
		if step.Row == x.Row && step.Col == x.Col {
			return true
		}
	}

	return false
}
