package domain

import (
	"time"
)

type Match struct {
	ID          int       `json:"id"`
	MatchNumber int       `json:"match_number"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	Winner      string    `json:"winner"`
	Turns       []Turn    `json:"turns"`
}

type Turn struct {
	ID          int       `json:"id"`
	TurnNumber  int       `json:"turn_number"`
	Time        time.Time `json:"time"`
	Player      string    `json:"player"`
	BallPower   int       `json:"ball_power"`
	RoutineID   string    `json:"routine_id"`
	MatchNumber int       `json:"match_number"`
}
