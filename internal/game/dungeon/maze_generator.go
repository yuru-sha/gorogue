package dungeon

import (
	"math/rand"

	"github.com/yuru-sha/gorogue/internal/utils/logger"
)

// MazeGenerator generates maze-type dungeons
type MazeGenerator struct {
	level *Level
}

// NewMazeGenerator creates a new maze generator
func NewMazeGenerator(level *Level) *MazeGenerator {
	return &MazeGenerator{
		level: level,
	}
}

// GenerateMaze generates a maze-type dungeon (PyRogue style)
func (g *MazeGenerator) GenerateMaze() {
	logger.Info("Generating maze dungeon", "floor", g.level.FloorNumber)

	// Initialize all tiles as walls
	for y := 0; y < g.level.Height; y++ {
		for x := 0; x < g.level.Width; x++ {
			g.level.SetTile(x, y, TileWall)
		}
	}

	// Start carving from the center
	startX := g.level.Width / 2
	startY := g.level.Height / 2

	// Make sure start position is odd (for proper maze generation)
	if startX%2 == 0 {
		startX--
	}
	if startY%2 == 0 {
		startY--
	}

	// Carve the maze using recursive backtracking
	g.carveMaze(startX, startY)

	// Add some additional connections to make it less linear
	g.addRandomConnections()

	logger.Info("Maze generation completed",
		"width", g.level.Width,
		"height", g.level.Height,
		"start", startX, startY)
}

// carveMaze recursively carves maze passages
func (g *MazeGenerator) carveMaze(x, y int) {
	// Mark current position as floor
	g.level.SetTile(x, y, TileFloor)

	// Directions: up, right, down, left
	directions := []Position{
		{X: 0, Y: -2}, // up
		{X: 2, Y: 0},  // right
		{X: 0, Y: 2},  // down
		{X: -2, Y: 0}, // left
	}

	// Shuffle directions for randomness
	rand.Shuffle(len(directions), func(i, j int) {
		directions[i], directions[j] = directions[j], directions[i]
	})

	// Try each direction
	for _, dir := range directions {
		nx, ny := x+dir.X, y+dir.Y

		// Check if new position is valid and unvisited
		if g.level.IsInBounds(nx, ny) && g.level.GetTile(nx, ny).Type == TileWall {
			// Check if we can carve to this position
			if g.canCarveDirection(x, y, nx, ny) {
				// Carve the wall between current and new position
				wallX := x + dir.X/2
				wallY := y + dir.Y/2
				g.level.SetTile(wallX, wallY, TileFloor)

				// Recursively carve from new position
				g.carveMaze(nx, ny)
			}
		}
	}
}

// canCarveDirection checks if we can carve in a direction
func (g *MazeGenerator) canCarveDirection(fromX, fromY, toX, toY int) bool {
	// Check bounds
	if !g.level.IsInBounds(toX, toY) {
		return false
	}

	// Check if destination is a wall
	if g.level.GetTile(toX, toY).Type != TileWall {
		return false
	}

	// Check if we're not too close to the edge
	if toX < 2 || toY < 2 || toX >= g.level.Width-2 || toY >= g.level.Height-2 {
		return false
	}

	return true
}

// addRandomConnections adds some random connections to make the maze more interesting
func (g *MazeGenerator) addRandomConnections() {
	connectionCount := (g.level.Width * g.level.Height) / 200 // About 0.5% of total tiles

	for i := 0; i < connectionCount; i++ {
		// Pick a random wall
		x := 1 + rand.Intn(g.level.Width-2)
		y := 1 + rand.Intn(g.level.Height-2)

		// If it's a wall and connects two floor areas, make it a floor
		if g.level.GetTile(x, y).Type == TileWall && g.connectsFloorAreas(x, y) {
			g.level.SetTile(x, y, TileFloor)
		}
	}
}

// connectsFloorAreas checks if a wall position connects two different floor areas
func (g *MazeGenerator) connectsFloorAreas(x, y int) bool {
	floorCount := 0

	// Check 4 directions
	directions := []Position{
		{X: 0, Y: -1}, // up
		{X: 1, Y: 0},  // right
		{X: 0, Y: 1},  // down
		{X: -1, Y: 0}, // left
	}

	for _, dir := range directions {
		nx, ny := x+dir.X, y+dir.Y
		if g.level.IsInBounds(nx, ny) && g.level.GetTile(nx, ny).Type == TileFloor {
			floorCount++
		}
	}

	// If we have floors in at least 2 directions, this wall connects areas
	return floorCount >= 2
}

// findFloorTiles finds all floor tiles in the maze (for stair placement)
func (g *MazeGenerator) findFloorTiles() []Position {
	var floorTiles []Position

	for y := 0; y < g.level.Height; y++ {
		for x := 0; x < g.level.Width; x++ {
			if g.level.GetTile(x, y).Type == TileFloor {
				floorTiles = append(floorTiles, Position{X: x, Y: y})
			}
		}
	}

	return floorTiles
}
