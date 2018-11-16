package utils

import "time"

func NextDay(t time.Time) time.Time {
	return t.AddDate(0, 0, 1)
}

func DateZero(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

type TimeRange struct {
	From int64
	To   int64
}

// 30mins per range
// 8:00 - 11:00 3h -> 6
// 14:00 - 17:00 3h -> 6
func NormalTimeRange(t time.Time) []TimeRange {
	const min30 = time.Minute * 30

	t = DateZero(t)
	t8 := t.Add(time.Hour * 8)
	t14 := t.Add(time.Hour * 14)

	var trs []TimeRange
	for i := 0; i < 6; i++ {
		trs = append(trs, TimeRange{
			From: t8.Unix(),
			To:   t8.Add(min30).Unix() - 1,
		})

		t8 = t8.Add(min30)
	}

	for i := 0; i < 6; i++ {
		trs = append(trs, TimeRange{
			From: t14.Unix(),
			To:   t14.Add(min30).Unix() - 1,
		})

		t14 = t14.Add(min30)
	}

	return trs
}
