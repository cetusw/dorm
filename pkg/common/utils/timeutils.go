package utils

import "time"

func NowMoscow() time.Time {
	loc, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		loc = time.FixedZone("MSK", 3*3600)
	}
	return time.Now().In(loc)
}
