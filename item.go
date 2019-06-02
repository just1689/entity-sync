package main

import "time"

type ItemV1 struct {
	ID         string
	Closed     bool
	ClosedDate time.Time
}
