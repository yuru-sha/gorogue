package dungeon

import (
	"math/rand"

	"github.com/yuru-sha/gorogue/internal/utils/logger"
)

// BSPNode represents a node in the binary space partitioning tree
type BSPNode struct {
	X, Y          int      // Position of the node
	Width, Height int      // Size of the node
	Room          *Room    // Room in this node (nil if not a leaf)
	LeftChild     *BSPNode // Left child node
	RightChild    *BSPNode // Right child node
	IsLeaf        bool     // Whether this is a leaf node
	SplitVertical bool     // Whether this node was split vertically
	Corridor      []Position // Corridor tiles for this node
}

// BSPGenerator generates dungeons using Binary Space Partitioning
type BSPGenerator struct {
	level    *Level
	root     *BSPNode
	minSize  int
	maxDepth int
}

// NewBSPGenerator creates a new BSP generator (PyRogue style)
func NewBSPGenerator(level *Level) *BSPGenerator {
	return &BSPGenerator{
		level:    level,
		minSize:  8,  // PyRogue MIN_SIZE: smaller for more rooms
		maxDepth: 12, // PyRogue DEPTH: increased for more room subdivision
	}
}

// GenerateRooms generates rooms using BSP algorithm
func (g *BSPGenerator) GenerateRooms() {
	logger.Info("Generating rooms using BSP algorithm")

	// Create root node covering the entire level (PyRogue style)
	g.root = &BSPNode{
		X:        0,
		Y:        0,
		Width:    g.level.Width,
		Height:   g.level.Height,
		IsLeaf:   true,
	}

	// Recursively split the space
	g.splitNode(g.root, 0)

	// Create rooms in leaf nodes
	g.createRooms(g.root)

	// Connect rooms with corridors (doors placed during corridor creation)
	g.connectRooms(g.root)

	logger.Info("BSP room generation completed",
		"total_rooms", len(g.level.Rooms),
		"tree_depth", g.calculateDepth(g.root),
	)
}

// splitNode recursively splits a node into two children
func (g *BSPGenerator) splitNode(node *BSPNode, depth int) {
	// Stop splitting if we've reached max depth or node is too small
	if depth >= g.maxDepth || node.Width < g.minSize*2 || node.Height < g.minSize*2 {
		return
	}

	// Determine split direction
	// Prefer splitting the longer dimension
	splitVertical := node.Width > node.Height
	if node.Width == node.Height {
		splitVertical = rand.Float64() < 0.5
	}

	var splitPos int
	if splitVertical {
		// Split vertically (left/right)
		minSplit := node.X + g.minSize
		maxSplit := node.X + node.Width - g.minSize
		if minSplit >= maxSplit {
			return // Can't split
		}
		splitPos = minSplit + rand.Intn(maxSplit-minSplit)
		
		// Create left and right children
		node.LeftChild = &BSPNode{
			X:      node.X,
			Y:      node.Y,
			Width:  splitPos - node.X,
			Height: node.Height,
			IsLeaf: true,
		}
		node.RightChild = &BSPNode{
			X:      splitPos,
			Y:      node.Y,
			Width:  node.X + node.Width - splitPos,
			Height: node.Height,
			IsLeaf: true,
		}
	} else {
		// Split horizontally (top/bottom)
		minSplit := node.Y + g.minSize
		maxSplit := node.Y + node.Height - g.minSize
		if minSplit >= maxSplit {
			return // Can't split
		}
		splitPos = minSplit + rand.Intn(maxSplit-minSplit)
		
		// Create top and bottom children
		node.LeftChild = &BSPNode{
			X:      node.X,
			Y:      node.Y,
			Width:  node.Width,
			Height: splitPos - node.Y,
			IsLeaf: true,
		}
		node.RightChild = &BSPNode{
			X:      node.X,
			Y:      splitPos,
			Width:  node.Width,
			Height: node.Y + node.Height - splitPos,
			IsLeaf: true,
		}
	}

	node.IsLeaf = false
	node.SplitVertical = splitVertical

	// Recursively split children
	g.splitNode(node.LeftChild, depth+1)
	g.splitNode(node.RightChild, depth+1)

	logger.Debug("Split BSP node",
		"x", node.X,
		"y", node.Y,
		"width", node.Width,
		"height", node.Height,
		"vertical", splitVertical,
		"split_pos", splitPos,
		"depth", depth,
	)
}

// createRooms creates rooms in leaf nodes
func (g *BSPGenerator) createRooms(node *BSPNode) {
	if node.IsLeaf {
		// Create a room in this leaf node
		room := g.createRoomInNode(node)
		if room != nil {
			node.Room = room
			g.level.Rooms = append(g.level.Rooms, room)
		}
	} else {
		// Recursively create rooms in children
		if node.LeftChild != nil {
			g.createRooms(node.LeftChild)
		}
		if node.RightChild != nil {
			g.createRooms(node.RightChild)
		}
	}
}

// createRoomInNode creates a room within the given node (PyRogue style)
func (g *BSPGenerator) createRoomInNode(node *BSPNode) *Room {
	// PyRogue style: fixed 2-tile margin from section boundaries
	margin := 2
	minRoomSize := 4
	
	// Available space after margin
	availableWidth := node.Width - margin*2
	availableHeight := node.Height - margin*2
	
	if availableWidth < minRoomSize || availableHeight < minRoomSize {
		return nil
	}
	
	// PyRogue style: room size within available space (with some randomization)
	width := minRoomSize + rand.Intn(availableWidth-minRoomSize+1)
	height := minRoomSize + rand.Intn(availableHeight-minRoomSize+1)
	
	// Ensure room doesn't exceed available space
	if width > availableWidth {
		width = availableWidth
	}
	if height > availableHeight {
		height = availableHeight
	}
	
	// PyRogue style: room position with fixed margin (centered within available space)
	maxXOffset := availableWidth - width
	maxYOffset := availableHeight - height
	if maxXOffset < 0 {
		maxXOffset = 0
	}
	if maxYOffset < 0 {
		maxYOffset = 0
	}
	x := node.X + margin + rand.Intn(maxXOffset+1)
	y := node.Y + margin + rand.Intn(maxYOffset+1)
	
	room := &Room{
		X:         x,
		Y:         y,
		Width:     width,
		Height:    height,
		Connected: true,
	}
	
	// Fill room with floor tiles
	for dy := 0; dy < height; dy++ {
		for dx := 0; dx < width; dx++ {
			if g.level.IsInBounds(x+dx, y+dy) {
				g.level.SetTile(x+dx, y+dy, TileFloor)
			}
		}
	}
	
	logger.Debug("Created room in BSP node",
		"room_x", x,
		"room_y", y,
		"room_width", width,
		"room_height", height,
		"node_x", node.X,
		"node_y", node.Y,
		"node_width", node.Width,
		"node_height", node.Height,
	)
	
	return room
}

// connectRooms connects rooms with corridors
func (g *BSPGenerator) connectRooms(node *BSPNode) {
	if node.IsLeaf {
		return
	}

	// Recursively connect children first
	if node.LeftChild != nil {
		g.connectRooms(node.LeftChild)
	}
	if node.RightChild != nil {
		g.connectRooms(node.RightChild)
	}

	// Connect the two child subtrees
	if node.LeftChild != nil && node.RightChild != nil {
		g.connectNodeSubtrees(node.LeftChild, node.RightChild, node.SplitVertical)
	}
}

// connectNodeSubtrees connects two subtrees with a corridor (PyRogue style)
func (g *BSPGenerator) connectNodeSubtrees(leftNode, rightNode *BSPNode, splitVertical bool) {
	// Find representative rooms from each subtree
	leftRoom := g.findRepresentativeRoom(leftNode)
	rightRoom := g.findRepresentativeRoom(rightNode)

	if leftRoom == nil || rightRoom == nil {
		logger.Debug("Skipping connection: one or both rooms not found")
		return
	}

	logger.Debug("Connecting rooms", 
		"left_room", leftRoom.X, leftRoom.Y, leftRoom.Width, leftRoom.Height,
		"right_room", rightRoom.X, rightRoom.Y, rightRoom.Width, rightRoom.Height,
		"split_vertical", splitVertical)

	// Create corridor between the rooms
	corridor := g.createCorridorBetweenRooms(leftRoom, rightRoom, splitVertical)
	
	// Store corridor information in the parent node
	// This will be used for door placement
	if leftNode.Corridor == nil {
		leftNode.Corridor = make([]Position, 0)
	}
	if rightNode.Corridor == nil {
		rightNode.Corridor = make([]Position, 0)
	}
	
	leftNode.Corridor = append(leftNode.Corridor, corridor...)
	rightNode.Corridor = append(rightNode.Corridor, corridor...)

	logger.Debug("Connected rooms with corridor", "corridor_length", len(corridor))
}

// findRepresentativeRoom finds a representative room from a subtree
func (g *BSPGenerator) findRepresentativeRoom(node *BSPNode) *Room {
	if node.IsLeaf {
		return node.Room
	}

	// Try left child first
	if node.LeftChild != nil {
		if room := g.findRepresentativeRoom(node.LeftChild); room != nil {
			return room
		}
	}

	// Try right child
	if node.RightChild != nil {
		if room := g.findRepresentativeRoom(node.RightChild); room != nil {
			return room
		}
	}

	return nil
}

// createCorridorBetweenRooms creates a corridor between two rooms (PyRogue style)
func (g *BSPGenerator) createCorridorBetweenRooms(room1, room2 *Room, splitVertical bool) []Position {
	var corridor []Position

	// PyRogue style: connect room centers with L-shaped corridors
	center1X := room1.X + room1.Width/2
	center1Y := room1.Y + room1.Height/2
	center2X := room2.X + room2.Width/2
	center2Y := room2.Y + room2.Height/2

	// PyRogue style: determine L-shape direction based on distance
	dx := abs(center2X - center1X)
	dy := abs(center2Y - center1Y)
	
	if dx > dy {
		// Horizontal distance is greater: go horizontal first, then vertical
		corridor = append(corridor, g.createHorizontalCorridor(center1X, center2X, center1Y)...)
		corridor = append(corridor, g.createVerticalCorridor(center1Y, center2Y, center2X)...)
	} else {
		// Vertical distance is greater: go vertical first, then horizontal
		corridor = append(corridor, g.createVerticalCorridor(center1Y, center2Y, center1X)...)
		corridor = append(corridor, g.createHorizontalCorridor(center1X, center2X, center2Y)...)
	}

	return corridor
}

// createHorizontalCorridor creates a horizontal corridor (PyRogue style: place doors while digging)
func (g *BSPGenerator) createHorizontalCorridor(x1, x2, y int) []Position {
	var corridor []Position
	
	minX := min(x1, x2)
	maxX := max(x1, x2)
	
	for x := minX; x <= maxX; x++ {
		if g.level.IsInBounds(x, y) {
			currentTile := g.level.GetTile(x, y)
			if currentTile.Type == TileWall {
				// PyRogue style: place door when breaking through room boundary
				if g.isRoomBoundaryWall(x, y) {
					doorType := g.selectDoorType()
					g.level.SetTile(x, y, doorType)
					logger.Debug("Placed door during corridor creation", "x", x, "y", y, "type", doorType)
				} else {
					g.level.SetTile(x, y, TileFloor)
				}
			}
			corridor = append(corridor, Position{X: x, Y: y})
		}
	}
	
	return corridor
}

// createVerticalCorridor creates a vertical corridor (PyRogue style: place doors while digging)
func (g *BSPGenerator) createVerticalCorridor(y1, y2, x int) []Position {
	var corridor []Position
	
	minY := min(y1, y2)
	maxY := max(y1, y2)
	
	for y := minY; y <= maxY; y++ {
		if g.level.IsInBounds(x, y) {
			currentTile := g.level.GetTile(x, y)
			if currentTile.Type == TileWall {
				// PyRogue style: place door when breaking through room boundary
				if g.isRoomBoundaryWall(x, y) {
					doorType := g.selectDoorType()
					g.level.SetTile(x, y, doorType)
					logger.Debug("Placed door during corridor creation", "x", x, "y", y, "type", doorType)
				} else {
					g.level.SetTile(x, y, TileFloor)
				}
			}
			corridor = append(corridor, Position{X: x, Y: y})
		}
	}
	
	return corridor
}

// placeDoors is now a no-op since doors are placed during corridor creation (PyRogue style)
func (g *BSPGenerator) placeDoors(node *BSPNode) {
	// PyRogue style: doors are placed during corridor creation
	// This method is kept for compatibility but does nothing
	logger.Debug("Door placement completed during corridor creation")
}

// selectDoorType selects door type based on PyRogue probabilities
func (g *BSPGenerator) selectDoorType() TileType {
	rand_val := rand.Float64()
	
	if rand_val < 0.1 {
		return TileSecretDoor // 10% secret doors
	} else if rand_val < 0.4 {
		return TileOpenDoor   // 30% open doors
	} else {
		return TileDoor       // 60% normal doors
	}
}

// PyRogue style: these functions are no longer needed since doors are placed during corridor creation

// isInsideRoom checks if a position is inside a room
func (g *BSPGenerator) isInsideRoom(x, y int, room *Room) bool {
	return x >= room.X && x < room.X+room.Width &&
		y >= room.Y && y < room.Y+room.Height
}

// isRoomBoundaryWall checks if a wall position is on the boundary of any room (PyRogue style)
func (g *BSPGenerator) isRoomBoundaryWall(x, y int) bool {
	for _, room := range g.level.Rooms {
		// Check if this wall is on the room's perimeter
		if (x == room.X-1 || x == room.X+room.Width) && y >= room.Y && y < room.Y+room.Height {
			return true
		}
		if (y == room.Y-1 || y == room.Y+room.Height) && x >= room.X && x < room.X+room.Width {
			return true
		}
	}
	return false
}

// calculateDepth calculates the depth of the BSP tree
func (g *BSPGenerator) calculateDepth(node *BSPNode) int {
	if node.IsLeaf {
		return 0
	}

	leftDepth := 0
	rightDepth := 0

	if node.LeftChild != nil {
		leftDepth = g.calculateDepth(node.LeftChild)
	}
	if node.RightChild != nil {
		rightDepth = g.calculateDepth(node.RightChild)
	}

	return 1 + max(leftDepth, rightDepth)
}

// GetRoot returns the root node for debugging/testing
func (g *BSPGenerator) GetRoot() *BSPNode {
	return g.root
}

// Helper functions for min/max are already defined in level.go
// abs function is already defined in room_connector.go