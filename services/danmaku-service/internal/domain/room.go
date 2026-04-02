package domain

import (
	"time"
)

type RoomState int32

const (
	ROOM_STATE_UNSPECIFIED RoomState = iota
	ROOM_STATE_ACTIVE
	ROOM_STATE_CLOSED
	ROOM_STATE_HIDDEN
)

type RoomModel struct {
	ID           string
	HostId       string
	Title        string
	State        RoomState
	ViewerCount  int32
	CreatedAt    time.Time
	BlockedUsers []string
}
