package routing

const (
	ArmyMovesPrefix = "army_moves"

	WarRecognitionsPrefix = "war"

	PauseKey = "pause"

	GameLogSlug = "game_logs"
)

const (
	ExchangePerilDirect = "peril_direct"
	ExchangePerilTopic  = "peril_topic"
)

// MoveMessage represents a message sent when a player moves units
type MoveMessage struct {
	Username    string      `json:"username"`
	Units       interface{} `json:"units"`       // Using interface{} as we don't know the exact type
	Destination string      `json:"destination"` // Location name
}

// PlayingState represents the paused/resumed state of the game
// type PlayingState struct {
// 	IsPaused bool `json:"is_paused"`
// }
