package actor

import (
	"container/heap"
	"fmt"
	"math"
)

// Node represents a node in the pathfinding algorithm
type Node struct {
	X, Y    int
	G, H, F float64
	Parent  *Node
	Index   int // For heap operations
}

// NodeHeap implements a priority queue for A* pathfinding
type NodeHeap []*Node

func (h NodeHeap) Len() int           { return len(h) }
func (h NodeHeap) Less(i, j int) bool { return h[i].F < h[j].F }
func (h NodeHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].Index = i
	h[j].Index = j
}

func (h *NodeHeap) Push(x interface{}) {
	n := len(*h)
	node := x.(*Node)
	node.Index = n
	*h = append(*h, node)
}

func (h *NodeHeap) Pop() interface{} {
	old := *h
	n := len(old)
	node := old[n-1]
	node.Index = -1
	*h = old[0 : n-1]
	return node
}

// AStar implements the A* pathfinding algorithm
func (m *Monster) AStar(targetX, targetY int, level LevelCollisionChecker) []Node {
	startX, startY := m.Position.X, m.Position.Y

	// Early exit if target is unreachable
	if !level.IsInBounds(targetX, targetY) || !level.IsWalkable(targetX, targetY) {
		return nil
	}

	// Early exit if we're already at the target
	if startX == targetX && startY == targetY {
		return []Node{{X: startX, Y: startY}}
	}

	openSet := &NodeHeap{}
	heap.Init(openSet)
	closedSet := make(map[string]bool)
	nodeMap := make(map[string]*Node)

	// Create start node
	startNode := &Node{
		X: startX,
		Y: startY,
		G: 0,
		H: m.heuristic(startX, startY, targetX, targetY),
	}
	startNode.F = startNode.G + startNode.H

	heap.Push(openSet, startNode)
	nodeMap[m.nodeKey(startX, startY)] = startNode

	// A* main loop
	for openSet.Len() > 0 {
		current := heap.Pop(openSet).(*Node)
		currentKey := m.nodeKey(current.X, current.Y)

		// Check if we've reached the target
		if current.X == targetX && current.Y == targetY {
			return m.reconstructPath(current)
		}

		closedSet[currentKey] = true

		// Check all neighbors
		for _, neighbor := range m.getNeighbors(current.X, current.Y, level) {
			neighborKey := m.nodeKey(neighbor.X, neighbor.Y)

			// Skip if already processed
			if closedSet[neighborKey] {
				continue
			}

			// Calculate tentative G score
			tentativeG := current.G + m.moveCost(current.X, current.Y, neighbor.X, neighbor.Y)

			// Check if this path to neighbor is better
			existingNode, exists := nodeMap[neighborKey]
			if !exists {
				existingNode = &Node{X: neighbor.X, Y: neighbor.Y, Index: -1}
				nodeMap[neighborKey] = existingNode
			}

			if !exists || tentativeG < existingNode.G {
				existingNode.G = tentativeG
				existingNode.H = m.heuristic(neighbor.X, neighbor.Y, targetX, targetY)
				existingNode.F = existingNode.G + existingNode.H
				existingNode.Parent = current

				// Add to open set if not already there
				if existingNode.Index == -1 {
					heap.Push(openSet, existingNode)
				} else {
					heap.Fix(openSet, existingNode.Index)
				}
			}
		}
	}

	// No path found
	return nil
}

// heuristic calculates the Manhattan distance heuristic
func (m *Monster) heuristic(x1, y1, x2, y2 int) float64 {
	return math.Abs(float64(x1-x2)) + math.Abs(float64(y1-y2))
}

// moveCost calculates the cost to move from one position to another
func (m *Monster) moveCost(x1, y1, x2, y2 int) float64 {
	// Diagonal movement costs more
	if x1 != x2 && y1 != y2 {
		return 1.4 // sqrt(2) approximation
	}
	return 1.0
}

// getNeighbors returns all valid neighboring positions
func (m *Monster) getNeighbors(x, y int, level LevelCollisionChecker) []Node {
	neighbors := []Node{}

	// Check all 8 directions
	for dx := -1; dx <= 1; dx++ {
		for dy := -1; dy <= 1; dy++ {
			if dx == 0 && dy == 0 {
				continue // Skip current position
			}

			nx, ny := x+dx, y+dy
			if m.CanMoveTo(nx, ny, level) {
				neighbors = append(neighbors, Node{X: nx, Y: ny})
			}
		}
	}

	return neighbors
}

// reconstructPath reconstructs the path from the target back to the start
func (m *Monster) reconstructPath(node *Node) []Node {
	path := []Node{}
	current := node

	for current != nil {
		path = append([]Node{{X: current.X, Y: current.Y}}, path...)
		current = current.Parent
	}

	return path
}

// nodeKey creates a unique key for a node position
func (m *Monster) nodeKey(x, y int) string {
	return fmt.Sprintf("%d,%d", x, y)
}

// FindPathToPlayer finds an optimal path to the player using A*
func (m *Monster) FindPathToPlayer(player *Player, level LevelCollisionChecker) []Node {
	return m.AStar(player.Position.X, player.Position.Y, level)
}

// MoveAlongPath moves the monster along a given path
func (m *Monster) MoveAlongPath(path []Node, level LevelCollisionChecker) bool {
	if len(path) < 2 {
		return false
	}

	// Get the next position in the path (skip current position)
	next := path[1]

	// Check if we can move to the next position
	if m.CanMoveTo(next.X, next.Y, level) {
		dx := next.X - m.Position.X
		dy := next.Y - m.Position.Y
		m.Position.Move(dx, dy)
		return true
	}

	return false
}
