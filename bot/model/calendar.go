package model

type Calendar string

const (
	CalendarGoogle  Calendar = "Google"
	CalendarOutlook Calendar = "Outlook"
	CalendarYandex  Calendar = "Yandex"
	CalendarApple   Calendar = "Apple"
)

func CalendarTypes() []Calendar {
	return []Calendar{
		CalendarGoogle,
		CalendarOutlook,
		CalendarYandex,
		CalendarApple,
	}
}
