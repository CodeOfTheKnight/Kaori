package main

type Notifications struct {
	Viewed Notification
	NotViewed Notification
}

type Notification struct {
	IdNotification string
	Time int64
	Description string
	Icon string
}