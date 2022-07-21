package utils

import "time"

const TimeFormat = "2006-01-02 15:04:05"
const StatisticTimeFormat = "2006/01/02 15:04:05"
const DateFormat = "2006-01-02"
const StatisticDateFormat = "2006/01/02"
const HourFormat = "15:04:05"

func FormatTime(t time.Time, layout string) string {
	return t.Format(layout)
}

//转本地时间 - 需要指定格式
func FormatLocalTimeWithLayout(t time.Time, layout string) string {
	return t.In(time.Local).Format(layout)
}

//转本地时间 - 默认格式
func FormatLocalTime(t time.Time) string {
	return FormatLocalTimeWithLayout(t, TimeFormat)
}
