package utils

import "time"

func NowMoscow() time.Time {
	loc, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		loc = time.FixedZone("MSK", 3*3600)
	}
	return time.Now().In(loc)
}

func IsFirstWeekOfMonth(date time.Time) bool {
	year, week := date.ISOWeek()

	firstDayOfMonth := time.Date(
		date.Year(),
		date.Month(),
		1,
		0,
		0,
		0,
		0,
		date.Location(),
	)

	firstDayYear, firstDayWeek := firstDayOfMonth.ISOWeek()

	return year == firstDayYear && week == firstDayWeek
}
