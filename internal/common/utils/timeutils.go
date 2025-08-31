package utils

import "time"

func NowMoscow() time.Time {
	loc, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		loc = time.FixedZone("MSK", 3*3600)
	}
	return time.Now().In(loc)
}

func GetLastWeekDay(weekday time.Weekday) time.Time {
	now := time.Now()
	daysAgo := (now.Weekday() - weekday + 7) % 7
	return now.AddDate(0, 0, -int(daysAgo))
}
