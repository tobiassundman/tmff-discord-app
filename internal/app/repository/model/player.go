package model

import "time"

type Player struct {
	Name      string    `db:"name"`
	Elo       float64   `db:"elo"`
	BGAID     string    `db:"bga_id"`
	CreatedAt time.Time `db:"created_at"`
}
