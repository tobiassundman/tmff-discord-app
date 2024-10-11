package model

import "time"

type Player struct {
	Name      string    `db:"name"`
	BGAID     string    `db:"bga_id"`
	CreatedAt time.Time `db:"created_at"`
}
