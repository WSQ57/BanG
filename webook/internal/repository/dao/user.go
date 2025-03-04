package domain

type User struct {
	Addr address
}

type address struct {
	Id     int64
	UserId int64
}
