package model

import "time"

type Player struct {
	ID        int       `db:"id"`
	Name      string    `db:"name"`
	BGAID     string    `db:"bga_id"`
	CreatedAt time.Time `db:"created_at"`
}
