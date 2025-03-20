package domain

import "time"

// User领域对象，对应ddd中的entity
type User struct {
	Id       int64
	Email    string
	Password string
	Phone    string
	Nickname string
	AboutMe  string
	Birthday time.Time
	Ctime    time.Time

	WechatInfo WechatInfo
	// Addr address
}

// type address struct {
// }
