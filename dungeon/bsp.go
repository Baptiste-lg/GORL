package dungeon

import "math/rand"

const (
	MinRoomSize = 4
	MaxRoomSize = 10
	MinLeafSize = 8
)

// Room represents a rectangular room in the dungeon.
type Room struct {
	X, Y, W, H int
}

// CenterX returns the horizontal center of the room.
func (r Room) CenterX() int { return r.X + r.W/2 }

// CenterY returns the vertical center of the room.
func (r Room) CenterY() int { return r.Y + r.H/2 }

// bspNode is a node in the BSP tree.
type bspNode struct {
	x, y, w, h  int
	left, right *bspNode
	room        *Room
}

// split attempts to split this node into two children.
func (n *bspNode) split(rng *rand.Rand) bool {
	if n.left != nil {
		return false // already split
	}

	// Decide split direction: prefer splitting the longer axis
	splitH := rng.Intn(2) == 0
	if float64(n.w)/float64(n.h) >= 1.25 {
		splitH = false // wide → split vertically
	} else if float64(n.h)/float64(n.w) >= 1.25 {
		splitH = true // tall → split horizontally
	}

	maxSize := n.h - MinLeafSize
	if !splitH {
		maxSize = n.w - MinLeafSize
	}
	if maxSize < MinLeafSize {
		return false // too small to split
	}

	split := rng.Intn(maxSize-MinLeafSize+1) + MinLeafSize

	if splitH {
		n.left = &bspNode{x: n.x, y: n.y, w: n.w, h: split}
		n.right = &bspNode{x: n.x, y: n.y + split, w: n.w, h: n.h - split}
	} else {
		n.left = &bspNode{x: n.x, y: n.y, w: split, h: n.h}
		n.right = &bspNode{x: n.x + split, y: n.y, w: n.w - split, h: n.h}
	}
	return true
}

// createRoom places a random room within this leaf node.
func (n *bspNode) createRoom(rng *rand.Rand) {
	if n.left != nil || n.right != nil {
		return // not a leaf
	}

	w := rng.Intn(min(MaxRoomSize, n.w-2)-MinRoomSize+1) + MinRoomSize
	h := rng.Intn(min(MaxRoomSize, n.h-2)-MinRoomSize+1) + MinRoomSize
	x := n.x + rng.Intn(n.w-w-1) + 1
	y := n.y + rng.Intn(n.h-h-1) + 1

	n.room = &Room{X: x, Y: y, W: w, H: h}
}

// getRoom returns the room in this node or a descendant.
func (n *bspNode) getRoom(rng *rand.Rand) *Room {
	if n.room != nil {
		return n.room
	}
	if n.left == nil {
		return nil
	}
	lr := n.left.getRoom(rng)
	rr := n.right.getRoom(rng)
	if lr == nil {
		return rr
	}
	if rr == nil {
		return lr
	}
	if rng.Intn(2) == 0 {
		return lr
	}
	return rr
}

// generateBSP creates a BSP tree and returns the list of rooms.
func generateBSP(w, h int, rng *rand.Rand) []*Room {
	root := &bspNode{x: 0, y: 0, w: w, h: h}

	// Collect all nodes to process
	nodes := []*bspNode{root}
	for i := 0; i < len(nodes); i++ {
		n := nodes[i]
		if n.split(rng) {
			nodes = append(nodes, n.left, n.right)
		}
	}

	// Create rooms in leaf nodes
	var rooms []*Room
	for _, n := range nodes {
		if n.left == nil && n.right == nil {
			n.createRoom(rng)
			if n.room != nil {
				rooms = append(rooms, n.room)
			}
		}
	}

	return rooms
}

