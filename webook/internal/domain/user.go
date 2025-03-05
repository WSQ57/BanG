package domain

import "time"

// User领域对象，对应ddd中的entity
type User struct {
	Id       int64
	Email    string
	Password string
	Ctime    time.Time
	// Addr address
}

// type address struct {
// }
