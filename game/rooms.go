package game

// RoomRole identifies the purpose of a room on a floor.
type RoomRole int

const (
	RoleNormal    RoomRole = iota
	RoleSpawn              // Room 0, safe
	RoleStairs             // Contains stairs down
	RoleShop               // Shop NPC
	RoleTreasure           // Guaranteed item
	RoleChallenge          // Extra enemies, reward on clear
	RoleSecret             // Hidden, accessed via cracked walls
)
