package main

type Notifications struct {
	Viewed Notification `json:"viewed"`
	NotViewed Notification `json:"not_viewed"`
}

type Notification struct {
	IdNotification string `json:"id_notification"`
	Time int64 `json:"time"`
	Description string `json:"description"`
	Icon string `json:"icon"`
}