package utils

import "time"

const TimeFormat = "2006-01-02 15:04:05"

func FormatTime(t time.Time, layout string) string {
	return t.Format(layout)
}
