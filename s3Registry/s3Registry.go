package s3Registry

import "time"

type Reservation struct {
	Username     string    `json:"username"`
	Password     string    `json:"password"`
	ReserveDate  time.Time `json:"reserveDate"`
	ReserveCourt string    `json:"reserveCourt"`
	ReserveTime  string    `json:"reserveTime"`
}

type Registry struct {
	Bucket   string
	Dirname  string
	Filename string
	Content  []byte
}
