package dungeon

import (
	"math/rand"

	"github.com/yuru-sha/gorogue/internal/utils/logger"
)

// GridCell represents a single cell in the 3x3 grid
type GridCell struct {
	X, Y        int   // Grid coordinates (0-2)
	Room        *Room // The room in this cell (nil if no room)
	Connected   bool  // Whether this cell is connected to the dungeon
	HasRoom     bool  // Whether this cell has an actual room
	IsGone      bool  // Whether this is a "gone room" (corridor only)
	Connections []int // Connected grid cell indices
}

// GridGenerator implements the original Rogue 3x3 grid system
type GridGenerator struct {
	level          *Level
	gridWidth      int
	gridHeight     int
	cellWidth      int
	cellHeight     int
	grid           []*GridCell
	connectedCells map[int]bool
}

// NewGridGenerator creates a new grid-based room generator
func NewGridGenerator(level *Level) *GridGenerator {
	// Original Rogue uses a 3x3 grid
	gridWidth := 3
	gridHeight := 3
	cellWidth := level.Width / gridWidth
	cellHeight := level.Height / gridHeight

	generator := &GridGenerator{
		level:          level,
		gridWidth:      gridWidth,
		gridHeight:     gridHeight,
		cellWidth:      cellWidth,
		cellHeight:     cellHeight,
		grid:           make([]*GridCell, gridWidth*gridHeight),
		connectedCells: make(map[int]bool),
	}

	// Initialize grid cells
	for i := 0; i < gridWidth*gridHeight; i++ {
		x := i % gridWidth
		y := i / gridWidth
		generator.grid[i] = &GridCell{
			X:           x,
			Y:           y,
			Connected:   false,
			HasRoom:     false,
			IsGone:      false,
			Connections: make([]int, 0),
		}
	}

	return generator
}

// GenerateRooms generates rooms using the original Rogue 3x3 grid algorithm
func (g *GridGenerator) GenerateRooms() {
	logger.Info("Generating rooms using 3x3 grid system")

	// Step 1: Decide which cells will have rooms
	g.decideRoomPlacements()

	// Step 2: Create the actual rooms in the designated cells
	g.createRooms()

	// Step 3: Connect rooms using the original Rogue algorithm
	g.connectRooms()

	// Step 4: Generate corridors between connected rooms
	g.generateCorridors()

	logger.Info("Grid-based room generation completed",
		"total_cells", len(g.grid),
		"room_cells", g.countRoomCells(),
		"connected_cells", len(g.connectedCells),
	)
}

// decideRoomPlacements decides which grid cells will contain rooms
func (g *GridGenerator) decideRoomPlacements() {
	for i, cell := range g.grid {
		// Original Rogue: 70-80% chance of having a room in each cell
		if rand.Float64() < 0.75 {
			// 15% chance of being a "gone room" (corridor only)
			if rand.Float64() < 0.15 {
				cell.IsGone = true
				cell.HasRoom = false
				logger.Debug("Marked cell as gone room",
					"grid_x", cell.X,
					"grid_y", cell.Y,
					"index", i,
				)
			} else {
				cell.HasRoom = true
				logger.Debug("Marked cell for room placement",
					"grid_x", cell.X,
					"grid_y", cell.Y,
					"index", i,
				)
			}
		}
	}
}

// createRooms creates the actual rooms in the designated cells
func (g *GridGenerator) createRooms() {
	for i, cell := range g.grid {
		if cell.HasRoom {
			room := g.createRoomInCell(cell)
			if room != nil {
				cell.Room = room
				g.level.Rooms = append(g.level.Rooms, room)
				logger.Debug("Created room in grid cell",
					"grid_x", cell.X,
					"grid_y", cell.Y,
					"room_x", room.X,
					"room_y", room.Y,
					"width", room.Width,
					"height", room.Height,
				)
			}
		} else if cell.IsGone {
			// Create gone room (corridor space)
			g.createGoneRoomInCell(cell)
			logger.Debug("Created gone room in grid cell",
				"grid_x", cell.X,
				"grid_y", cell.Y,
				"index", i,
			)
		}
	}
}

// createRoomInCell creates a room within the specified grid cell
func (g *GridGenerator) createRoomInCell(cell *GridCell) *Room {
	// Calculate the boundaries of this grid cell
	cellStartX := cell.X * g.cellWidth
	cellStartY := cell.Y * g.cellHeight
	cellEndX := cellStartX + g.cellWidth
	cellEndY := cellStartY + g.cellHeight

	// Leave margins for walls and corridors
	margin := 2
	maxWidth := g.cellWidth - margin*2
	maxHeight := g.cellHeight - margin*2

	// Ensure minimum room size
	if maxWidth < MinRoomSize || maxHeight < MinRoomSize {
		return nil
	}

	// Generate room size (smaller than the cell)
	width := MinRoomSize + rand.Intn(maxWidth-MinRoomSize+1)
	height := MinRoomSize + rand.Intn(maxHeight-MinRoomSize+1)

	// Position room within the cell (centered with some randomness)
	maxX := cellEndX - width - margin
	maxY := cellEndY - height - margin
	x := cellStartX + margin + rand.Intn(maxX-cellStartX-margin+1)
	y := cellStartY + margin + rand.Intn(maxY-cellStartY-margin+1)

	// Create the room
	room := &Room{
		X:         x,
		Y:         y,
		Width:     width,
		Height:    height,
		Connected: true, // All rooms created in grid system are connected
	}

	// Fill the room with floor tiles
	for dy := 0; dy < height; dy++ {
		for dx := 0; dx < width; dx++ {
			if g.level.IsInBounds(x+dx, y+dy) {
				g.level.SetTile(x+dx, y+dy, TileFloor)
			}
		}
	}

	return room
}

// createGoneRoomInCell creates a gone room (corridor space) in the specified cell
func (g *GridGenerator) createGoneRoomInCell(cell *GridCell) {
	// Calculate the boundaries of this grid cell
	cellStartX := cell.X * g.cellWidth
	cellStartY := cell.Y * g.cellHeight

	// Create a smaller corridor space in the center of the cell
	corridorWidth := 3 + rand.Intn(4)  // 3-6 tiles wide
	corridorHeight := 3 + rand.Intn(4) // 3-6 tiles high

	startX := cellStartX + (g.cellWidth-corridorWidth)/2
	startY := cellStartY + (g.cellHeight-corridorHeight)/2

	// Fill with floor tiles
	for dy := 0; dy < corridorHeight; dy++ {
		for dx := 0; dx < corridorWidth; dx++ {
			if g.level.IsInBounds(startX+dx, startY+dy) {
				g.level.SetTile(startX+dx, startY+dy, TileFloor)
			}
		}
	}
}

// connectRooms implements the original Rogue room connection algorithm
func (g *GridGenerator) connectRooms() {
	// Step 1: Choose a random starting cell that has a room or is a gone room
	startCell := g.chooseRandomActiveCell()
	if startCell == -1 {
		logger.Warn("No active cells found for connection")
		return
	}

	// Mark starting cell as connected
	g.grid[startCell].Connected = true
	g.connectedCells[startCell] = true

	// Step 2: Connect adjacent unconnected cells
	for {
		connectedAdjacent := false
		for cellIndex := range g.connectedCells {
			adjacent := g.getAdjacentCells(cellIndex)
			for _, adjIndex := range adjacent {
				if !g.grid[adjIndex].Connected && g.isActiveCell(adjIndex) {
					g.connectCells(cellIndex, adjIndex)
					g.grid[adjIndex].Connected = true
					g.connectedCells[adjIndex] = true
					connectedAdjacent = true
					break
				}
			}
			if connectedAdjacent {
				break
			}
		}
		if !connectedAdjacent {
			break
		}
	}

	// Step 3: Connect any remaining unconnected cells
	for i, cell := range g.grid {
		if !cell.Connected && g.isActiveCell(i) {
			nearestConnected := g.findNearestConnectedCell(i)
			if nearestConnected != -1 {
				g.connectCells(nearestConnected, i)
				cell.Connected = true
				g.connectedCells[i] = true
			}
		}
	}

	// Step 4: Add some extra connections for variety (0-2 additional connections)
	extraConnections := rand.Intn(3)
	for i := 0; i < extraConnections; i++ {
		g.addRandomConnection()
	}
}

// generateCorridors generates corridors between connected rooms
func (g *GridGenerator) generateCorridors() {
	for cellIndex, cell := range g.grid {
		if cell.Connected {
			for _, connectedIndex := range cell.Connections {
				g.createCorridorBetweenCells(cellIndex, connectedIndex)
			}
		}
	}
}

// Helper functions

// chooseRandomActiveCell chooses a random cell that has a room or is a gone room
func (g *GridGenerator) chooseRandomActiveCell() int {
	activeCells := make([]int, 0)
	for i, cell := range g.grid {
		if cell.HasRoom || cell.IsGone {
			activeCells = append(activeCells, i)
		}
	}
	if len(activeCells) == 0 {
		return -1
	}
	return activeCells[rand.Intn(len(activeCells))]
}

// isActiveCell checks if a cell has a room or is a gone room
func (g *GridGenerator) isActiveCell(index int) bool {
	if index < 0 || index >= len(g.grid) {
		return false
	}
	cell := g.grid[index]
	return cell.HasRoom || cell.IsGone
}

// getAdjacentCells returns indices of adjacent cells (up, down, left, right)
func (g *GridGenerator) getAdjacentCells(index int) []int {
	cell := g.grid[index]
	adjacent := make([]int, 0)

	// Up
	if cell.Y > 0 {
		adjacent = append(adjacent, (cell.Y-1)*g.gridWidth+cell.X)
	}
	// Down
	if cell.Y < g.gridHeight-1 {
		adjacent = append(adjacent, (cell.Y+1)*g.gridWidth+cell.X)
	}
	// Left
	if cell.X > 0 {
		adjacent = append(adjacent, cell.Y*g.gridWidth+(cell.X-1))
	}
	// Right
	if cell.X < g.gridWidth-1 {
		adjacent = append(adjacent, cell.Y*g.gridWidth+(cell.X+1))
	}

	return adjacent
}

// connectCells connects two cells
func (g *GridGenerator) connectCells(from, to int) {
	g.grid[from].Connections = append(g.grid[from].Connections, to)
	g.grid[to].Connections = append(g.grid[to].Connections, from)
	logger.Debug("Connected cells",
		"from", from,
		"to", to,
	)
}

// findNearestConnectedCell finds the nearest connected cell
func (g *GridGenerator) findNearestConnectedCell(index int) int {
	minDistance := g.gridWidth * g.gridHeight
	nearest := -1

	cell := g.grid[index]
	for connectedIndex := range g.connectedCells {
		connectedCell := g.grid[connectedIndex]
		distance := abs(cell.X-connectedCell.X) + abs(cell.Y-connectedCell.Y)
		if distance < minDistance {
			minDistance = distance
			nearest = connectedIndex
		}
	}

	return nearest
}

// addRandomConnection adds a random connection between two connected cells
func (g *GridGenerator) addRandomConnection() {
	connectedIndices := make([]int, 0)
	for index := range g.connectedCells {
		connectedIndices = append(connectedIndices, index)
	}

	if len(connectedIndices) < 2 {
		return
	}

	// Pick two random connected cells
	from := connectedIndices[rand.Intn(len(connectedIndices))]
	to := connectedIndices[rand.Intn(len(connectedIndices))]

	if from != to {
		// Check if they're not already connected
		alreadyConnected := false
		for _, connection := range g.grid[from].Connections {
			if connection == to {
				alreadyConnected = true
				break
			}
		}
		if !alreadyConnected {
			g.connectCells(from, to)
		}
	}
}

// createCorridorBetweenCells creates a corridor between two connected cells
func (g *GridGenerator) createCorridorBetweenCells(from, to int) {
	fromCell := g.grid[from]
	toCell := g.grid[to]

	// Get cell centers
	fromX := fromCell.X*g.cellWidth + g.cellWidth/2
	fromY := fromCell.Y*g.cellHeight + g.cellHeight/2
	toX := toCell.X*g.cellWidth + g.cellWidth/2
	toY := toCell.Y*g.cellHeight + g.cellHeight/2

	// Create L-shaped corridor
	g.level.CreateHorizontalCorridor(fromX, toX, fromY)
	g.level.CreateVerticalCorridor(fromY, toY, toX)
}

// countRoomCells counts the number of cells with rooms
func (g *GridGenerator) countRoomCells() int {
	count := 0
	for _, cell := range g.grid {
		if cell.HasRoom {
			count++
		}
	}
	return count
}

// GetGrid returns the grid for debugging/testing
func (g *GridGenerator) GetGrid() []*GridCell {
	return g.grid
}
